// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pgbk

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gravitational/trace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/backend"
	pgcommon "github.com/gravitational/teleport/lib/backend/pgbk/common"
	"github.com/gravitational/teleport/lib/defaults"
)

func (b *Backend) backgroundExpiry(ctx context.Context) {
	defer b.log.Info("Exited expiry loop.")

	for ctx.Err() == nil {
		// "DELETE FROM kv WHERE expires <= now()" but more complicated: logical
		// decoding can become really really slow if a transaction is big enough
		// to spill on disk - max_changes_in_memory (4096) changes before
		// Postgres 13, or logical_decoding_work_mem (64MiB) bytes of total size
		// in Postgres 13 and later; thankfully, we can just limit our
		// transactions to a small-ish number of affected rows (1000 seems to
		// work ok) as we don't need atomicity for this; we run a tight loop
		// here because it could be possible to have more than ExpiryBatchSize
		// new items expire every ExpiryInterval, so we could end up not ever
		// catching up
		for i := 0; i < backend.DefaultRangeLimit/b.cfg.ExpiryBatchSize; i++ {
			t0 := time.Now()
			// TODO(espadolini): try getting keys in a read-only deferrable
			// transaction and deleting them later to reduce potential
			// serialization issues
			deleted, err := pgcommon.RetryIdempotent(ctx, b.log, func() (int64, error) {
				// LIMIT without ORDER BY might get executed poorly because the
				// planner doesn't have any idea of how many rows will be chosen
				// or skipped, and it's not necessary but it's a nice touch that
				// we'll be deleting expired items in expiration order
				tag, err := b.pool.Exec(ctx,
					"DELETE FROM kv WHERE key IN (SELECT key FROM kv"+
						" WHERE expires IS NOT NULL AND expires <= now()"+
						" ORDER BY expires LIMIT $1 FOR UPDATE)",
					b.cfg.ExpiryBatchSize,
				)
				if err != nil {
					return 0, trace.Wrap(err)
				}
				return tag.RowsAffected(), nil
			})
			if err != nil {
				b.log.WithError(err).Error("Failed to delete expired items.")
				break
			}

			if deleted > 0 {
				b.log.WithFields(logrus.Fields{
					"deleted": deleted,
					"elapsed": time.Since(t0).String(),
				}).Debug("Deleted expired items.")
			}

			if deleted < int64(b.cfg.ExpiryBatchSize) {
				break
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(b.cfg.ExpiryInterval)):
		}
	}
}

func (b *Backend) backgroundChangeFeed(ctx context.Context) {
	defer b.log.Info("Exited change feed loop.")
	defer b.buf.Close()

	for ctx.Err() == nil {
		b.log.Info("Starting change feed stream.")
		err := b.runChangeFeed(ctx)
		if ctx.Err() != nil {
			break
		}
		b.log.WithError(err).Error("Change feed stream lost.")

		select {
		case <-ctx.Done():
			return
		case <-time.After(defaults.HighResPollingPeriod):
		}
	}
}

// runChangeFeed will connect to the database, start a change feed and emit
// events. Assumes that b.buf is not initialized but not closed, and will reset
// it before returning.
func (b *Backend) runChangeFeed(ctx context.Context) error {
	// we manually copy the pool configuration and connect because we don't want
	// to hit a connection limit or mess with the connection pool stats; we need
	// a separate, long-running connection here anyway.
	poolConfig := b.pool.Config()
	if poolConfig.BeforeConnect != nil {
		if err := poolConfig.BeforeConnect(ctx, poolConfig.ConnConfig); err != nil {
			return trace.Wrap(err)
		}
	}
	conn, err := pgx.ConnectConfig(ctx, poolConfig.ConnConfig)
	if err != nil {
		return trace.Wrap(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := conn.Close(ctx); err != nil && ctx.Err() != nil {
			b.log.WithError(err).Warn("Error closing change feed connection.")
		}
	}()

	// reading from a replication slot adds to the postgres log at "log" level
	// (right below "fatal") for every poll, and we poll every second here, so
	// we try to silence the logs for this connection; this can fail because of
	// permission issues, which would delete the temporary slot (it's deleted on
	// any error), so we have to do it before that
	if _, err := conn.Exec(ctx, "SET log_min_messages TO fatal", pgx.QueryExecModeExec); err != nil {
		b.log.WithError(err).Debug("Failed to silence log messages for change feed session.")
	}

	// this can be useful if we're some sort of admin but we haven't gotten the
	// REPLICATION attribute yet
	// HACK(espadolini): ALTER ROLE CURRENT_USER REPLICATION just crashes postgres on Azure
	if _, err := conn.Exec(ctx,
		fmt.Sprintf("ALTER ROLE \"%v\" REPLICATION", poolConfig.ConnConfig.User),
		pgx.QueryExecModeExec,
	); err != nil {
		b.log.WithError(err).Debug("Failed to enable replication for the current user.")
	}

	u := uuid.New()
	slotName := hex.EncodeToString(u[:])

	b.log.WithField("slot_name", slotName).Info("Setting up change feed.")
	if _, err := conn.Exec(ctx,
		"SELECT * FROM pg_create_logical_replication_slot($1, 'wal2json', true)",
		pgx.QueryExecModeExec, slotName,
	); err != nil {
		return trace.Wrap(err)
	}

	b.log.WithField("slot_name", slotName).Info("Change feed started.")
	b.buf.SetInit()
	defer b.buf.Reset()

	for ctx.Err() == nil {
		events, err := b.pollChangeFeed(ctx, conn, slotName)
		if err != nil {
			return trace.Wrap(err)
		}

		// tight loop if we hit the batch size
		if events >= int64(b.cfg.ChangeFeedBatchSize) {
			continue
		}

		select {
		case <-ctx.Done():
			return trace.Wrap(ctx.Err())
		case <-time.After(time.Duration(b.cfg.ChangeFeedPollInterval)):
		}
	}
	return trace.Wrap(err)
}

type wal2jsonValue struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value *string `json:"value"`
}
type wal2jsonMessage struct {
	Action   string          `json:"action"`
	Schema   string          `json:"schema"`
	Table    string          `json:"table"`
	Columns  []wal2jsonValue `json:"columns"`
	Identity []wal2jsonValue `json:"identity"`

	Transactional bool   `json:"transactional"`
	Prefix        string `json:"prefix"`
	Content       string `json:"content"`
}

// pollChangeFeed will poll the change feed and emit any fetched events, if any.
// It returns the count of received/emitted events.
func (b *Backend) pollChangeFeed(ctx context.Context, conn *pgx.Conn, slotName string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	t0 := time.Now()
	rows, _ := conn.Query(ctx,
		"SELECT data FROM pg_logical_slot_get_changes($1, NULL, $2,"+
			" 'format-version', '2', 'add-tables', 'public.kv', 'include-transaction', 'false')",
		slotName, b.cfg.ChangeFeedBatchSize)

	var data string
	tag, err := pgx.ForEachRow(rows, []any{&data}, func() error {
		var m wal2jsonMessage
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			return trace.Wrap(err)
		}

		switch m.Action {
		case "I", "U", "D", "T":
			if m.Schema != "public" || m.Table != "kv" {
				b.log.WithFields(logrus.Fields{
					"schema": m.Schema,
					"table":  m.Table,
				}).Debug("Received WAL message for unexpected table (should not happen).")
				return nil
			}
		}

		switch m.Action {
		case "I", "U":
			if len(m.Columns) != 4 ||
				m.Columns[0].Name != "key" || m.Columns[0].Type != "bytea" || m.Columns[0].Value == nil ||
				m.Columns[1].Name != "value" || m.Columns[1].Type != "bytea" || m.Columns[1].Value == nil ||
				m.Columns[2].Name != "expires" || m.Columns[2].Type != "timestamp with time zone" ||
				m.Columns[3].Name != "revision" || m.Columns[3].Type != "uuid" || m.Columns[3].Value == nil {
				return trace.BadParameter("invalid schema on insert/update")
			}
		}

		switch m.Action {
		case "I":
			if len(m.Identity) != 0 {
				return trace.BadParameter("unexpected identity on insert")
			}
		case "U", "D":
			if len(m.Identity) != 1 ||
				m.Identity[0].Name != "key" || m.Identity[0].Type != "bytea" || m.Identity[0].Value == nil {
				return trace.BadParameter("invalid schema on update/delete")
			}
			if m.Action == "U" && *m.Columns[0].Value != *m.Identity[0].Value {
				return trace.BadParameter("item renaming is not supported")
			}
		}

		switch m.Action {
		case "I", "U":
			key, err := hex.DecodeString(*m.Columns[0].Value)
			if err != nil {
				return trace.Wrap(err, "decoding key")
			}
			value, err := hex.DecodeString(*m.Columns[1].Value)
			if err != nil {
				return trace.Wrap(err, "decoding value")
			}
			var expires zeronull.Timestamptz
			if m.Columns[2].Value != nil {
				if err := expires.Scan(*m.Columns[2].Value); err != nil {
					return trace.Wrap(err, "scanning expires")
				}
			}
			revision, err := uuid.Parse(*m.Columns[3].Value)
			if err != nil {
				return trace.Wrap(err, "parsing revision")
			}

			b.buf.Emit(backend.Event{
				Type: types.OpPut,
				Item: backend.Item{
					Key:     key,
					Value:   value,
					Expires: time.Time(expires),
				},
			})
			_ = revision
			return nil
		case "D":
			key, err := hex.DecodeString(*m.Identity[0].Value)
			if err != nil {
				return trace.Wrap(err, "decoding key")
			}
			b.buf.Emit(backend.Event{
				Type: types.OpDelete,
				Item: backend.Item{
					Key: key,
				},
			})
			return nil
		case "M":
			b.log.WithField("prefix", m.Prefix).Debug("Received WAL message.")
			return nil
		case "B", "C":
			b.log.Debug("Received transaction message in change feed (should not happen).")
			return nil
		case "T":
			// it could be possible to just reset the event buffer and
			// continue from the next row but it's not worth the effort
			// compared to just killing this connection and reconnecting,
			// and this should never actually happen anyway - deleting
			// everything from the backend would leave Teleport in a very
			// broken state
			return trace.BadParameter("received truncate WAL message, can't continue")
		default:
			return trace.BadParameter("received unknown WAL message %q", m.Action)
		}
	})
	if err != nil {
		return 0, trace.Wrap(err)
	}

	events := tag.RowsAffected()

	if events > 0 {
		b.log.WithFields(logrus.Fields{
			"events":  events,
			"elapsed": time.Since(t0).String(),
		}).Debug("Fetched change feed events.")
	}

	return events, nil
}
