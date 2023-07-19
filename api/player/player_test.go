package player_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gravitational/teleport/api/player"
	"github.com/gravitational/teleport/api/types/events"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

// test cases:
// - streams N events
// - streams N events with proper timing
// - can be canceled mid stream
// - can be paused and resumed

func TestBasicStream(t *testing.T) {
	clk := clockwork.NewFakeClock()
	p, err := player.New(&player.Config{
		Clock:     clk,
		Log:       nil, // TODO (nop)
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

// simpleStreamer streams a fake session that contains
// count events, one per second
type simpleStreamer struct {
	count int64
}

func (s *simpleStreamer) StreamSessionEvents(ctx context.Context, sessionID string, startIndex int64) (chan events.AuditEvent, chan error) {
	errors := make(chan error, 1)
	evts := make(chan events.AuditEvent)

	go func() {
		defer close(evts)

		for i := int64(0); i < s.count; i++ {
			select {
			case <-ctx.Done():
				return
			case evts <- &events.SessionPrint{
				Data:              []byte(fmt.Sprintf("event %d\n", i)),
				ChunkIndex:        i, // TODO (deprecate this?)
				DelayMilliseconds: i * 1000,
			}:
			}
		}
	}()

	return evts, errors
}
