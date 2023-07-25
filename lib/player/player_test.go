package player_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/lib/events"
	"github.com/gravitational/teleport/lib/player"
	"github.com/gravitational/teleport/lib/session"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

func TestBasicStream(t *testing.T) {
	clk := clockwork.NewFakeClock()
	p, err := player.New(&player.Config{
		Clock:     clk,
		SessionID: "test-session",
		Streamer:  &simpleStreamer{count: 3},
	})
	require.NoError(t, err)

	require.NoError(t, p.Play())

	count := 0
	for range p.C() {
		count++
	}

	require.Equal(t, 3, count)
}

func TestPlayPause(t *testing.T) {
	clk := clockwork.NewFakeClock()
	p, err := player.New(&player.Config{
		Clock:     clk,
		SessionID: "test-session",
		Streamer:  &simpleStreamer{count: 3},
	})
	require.NoError(t, err)

	// pausing an already paused player should be a no-op
	require.NoError(t, p.Pause())
	require.NoError(t, p.Pause())

	// toggling back and forth between play and pause
	// should not impact our ability to receive all
	// 3 events
	require.NoError(t, p.Play())
	require.NoError(t, p.Pause())
	require.NoError(t, p.Play())

	count := 0
	for range p.C() {
		count++
	}

	require.Equal(t, 3, count)
}

func TestAppliesTiming(t *testing.T) {
	for _, test := range []struct {
		desc    string
		speed   float64
		advance time.Duration
	}{
		{
			desc:    "half speed",
			speed:   0.5,
			advance: 2000 * time.Millisecond,
		},
		{
			desc:    "normal speed",
			speed:   1.0,
			advance: 1000 * time.Millisecond,
		},
		{
			desc:    "double speed",
			speed:   2.0,
			advance: 500 * time.Millisecond,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			clk := clockwork.NewFakeClock()
			p, err := player.New(&player.Config{
				Clock:     clk,
				SessionID: "test-session",
				Streamer:  &simpleStreamer{count: 3, delay: 1000},
			})
			require.NoError(t, err)

			require.NoError(t, p.SetSpeed(test.speed))
			require.NoError(t, p.Play())

			clk.BlockUntil(1) // player is now waiting to emit event 0

			// advance to next event (player will have emitted event 0
			// and will be waiting to emit event 1)
			clk.Advance(test.advance)
			clk.BlockUntil(1)
			evt := <-p.C()
			require.Equal(t, int64(0), evt.GetIndex())

			// repeat the process (emit event 1, wait for event 2)
			clk.Advance(test.advance)
			clk.BlockUntil(1)
			evt = <-p.C()
			require.Equal(t, int64(1), evt.GetIndex())

			// advance the player to allow event 2 to be emitted
			clk.Advance(test.advance)
			evt = <-p.C()
			require.Equal(t, int64(2), evt.GetIndex())

			// channel should be closed
			_, ok := <-p.C()
			require.False(t, ok, "player should be closed")
		})
	}
}

func TestClose(t *testing.T) {
	clk := clockwork.NewFakeClock()
	p, err := player.New(&player.Config{
		Clock:     clk,
		SessionID: "test-session",
		Streamer:  &simpleStreamer{count: 2, delay: 1000},
	})
	require.NoError(t, err)

	require.NoError(t, p.Play())

	clk.BlockUntil(1) // player is now waiting to emit event 0

	// advance to next event (player will have emitted event 0
	// and will be waiting to emit event 1)
	clk.Advance(1001 * time.Millisecond)
	clk.BlockUntil(1)
	evt := <-p.C()
	require.Equal(t, int64(0), evt.GetIndex())

	require.NoError(t, p.Close())

	// channel should have been closed
	_, ok := <-p.C()
	require.False(t, ok, "player channel should have been closed")
}

// simpleStreamer streams a fake session that contains
// count events, emitted at a particular interval
type simpleStreamer struct {
	count int64
	delay int64 // milliseconds
}

func (s *simpleStreamer) StreamSessionEvents(ctx context.Context, sessionID session.ID, startIndex int64) (chan apievents.AuditEvent, chan error) {
	errors := make(chan error, 1)
	evts := make(chan apievents.AuditEvent)

	go func() {
		defer close(evts)

		for i := int64(0); i < s.count; i++ {
			select {
			case <-ctx.Done():
				return
			case evts <- &apievents.SessionPrint{
				Metadata: apievents.Metadata{
					Type:  events.SessionPrintEvent,
					Index: i,
					ID:    strconv.Itoa(int(i)),
				},
				Data:              []byte(fmt.Sprintf("event %d\n", i)),
				ChunkIndex:        i, // TODO (deprecate this?)
				DelayMilliseconds: (i + 1) * s.delay,
			}:
			}
		}
	}()

	return evts, errors
}
