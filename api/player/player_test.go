package player_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gravitational/teleport/api/player"
	apievents "github.com/gravitational/teleport/api/types/events"

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
	clk := clockwork.NewFakeClock()
	p, err := player.New(&player.Config{
		Clock:     clk,
		SessionID: "test-session",
		Streamer:  &simpleStreamer{count: 3, delay: 1000},
	})
	require.NoError(t, err)

	// var count atomic.Int64
	// go func() {
	// 	for range p.C() {
	// 		println("received event!!")
	// 		count.Add(1)
	// 	}
	// }()

	require.NoError(t, p.Play())

	clk.BlockUntil(1) // player is now waiting to emit event 0

	// advance to next event (player will have emitted event 0
	// and will be waiting to emit event 1)
	clk.Advance(1001 * time.Millisecond)
	clk.BlockUntil(1)
	evt := <-p.C()
	require.Equal(t, int64(0), evt.GetIndex())

	// repeat the process (emit event 1, wait for event 2)
	clk.Advance(1001 * time.Millisecond)
	clk.BlockUntil(1)
	evt = <-p.C()
	require.Equal(t, int64(1), evt.GetIndex())

	// advance the player to allow event 2 to be emitted
	clk.Advance(1001 * time.Millisecond)
	evt = <-p.C()
	require.Equal(t, int64(2), evt.GetIndex())

	// channel should be closed
	_, ok := <-p.C()
	require.False(t, ok, "player should be closed")
}

// simpleStreamer streams a fake session that contains
// count events, emitted at a particular interval
type simpleStreamer struct {
	count int64
	delay int64 // milliseconds
}

func (s *simpleStreamer) StreamSessionEvents(ctx context.Context, sessionID string, startIndex int64) (chan apievents.AuditEvent, chan error) {
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
					Type:  "session.print", //events.SessionPrintEvent,
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
