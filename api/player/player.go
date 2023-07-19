// Package player includes an API to play back recorded sessions.
package player

import (
	"context"

	"github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
)

type Player struct {
	// read only config fields
	clock     clockwork.Clock
	log       logrus.FieldLogger
	sessionID string
	streamer  Streamer

	emit chan events.AuditEvent

	// playPause holds a channel to be closed when
	// the player transitions from paused to playing,
	// or nil if the player is already playing
	playPause chan chan struct{}
}

type Streamer interface {
	// TODO: need to change session ID type to avoid importing lib/session here
	StreamSessionEvents(ctx context.Context, sessionID string, startIndex int64) (chan events.AuditEvent, chan error)
}

type Config struct {
	Clock     clockwork.Clock
	Log       logrus.FieldLogger
	SessionID string
	Streamer  Streamer
}

func New(cfg *Config) (*Player, error) {
	if cfg.Streamer == nil {
		return nil, trace.BadParameter("missing Streamer")
	}

	if cfg.SessionID == "" {
		return nil, trace.BadParameter("missing SessionID")
	}

	clk := cfg.Clock
	if clk == nil {
		clk = clockwork.NewRealClock()
	}

	log := cfg.Log
	if log == nil {
		log = logrus.New().WithField(trace.Component, "player")
	}

	p := &Player{
		clock:     clk,
		log:       log,
		sessionID: cfg.SessionID,
		streamer:  cfg.Streamer,
		emit:      make(chan events.AuditEvent, 64),
		playPause: make(chan chan struct{}, 1),
	}

	// start in a paused state
	p.playPause <- make(chan struct{})

	go p.stream()

	return p, nil
}

func (p *Player) stream() {
	eventsC, errC := p.streamer.StreamSessionEvents(context.TODO(), p.sessionID, 0)
	for {
		select {
		case err := <-errC:
			// TODO: figure out how to surface the error
			// (probably close the chan and expose a method)
			p.log.Warn(err)
			return
		case evt := <-eventsC:
			if evt == nil {
				p.log.Debug("reached end of playback")
				close(p.emit)
				return
			}

			p.log.Debug("maybe waiting while paused")
			if err := p.waitWhilePaused(); err != nil {
				p.log.Warn(err)
				close(p.emit)
				return
			}
			p.log.Debug("unpaused, now playing")

			// TODO: add time delay

			p.log.Debugf("playing %v (%v)", evt.GetType(), evt.GetID())
			select {
			case p.emit <- evt:
			default:
				p.log.Warnf("dropped event %v, reader too slow", evt.GetID())
			}
		}
	}
}

// Close shuts down the player and cancels any streams that are
// in progress.
func (p *Player) Close() error {
	// TODO: either hold on to a context that we cancel, or
	// have a separate done chan
	return nil
}

// C returns a read only channel of recorded session events.
// The player manages the timing of events and writes them to the channel
// when they should be rendered. The channel is closed when the player
// has reached the end of playback.
func (p *Player) C() <-chan events.AuditEvent {
	return p.emit
}

// TODO: add an Err() method to be checked after C is closed

// Pause temporarily stops the player from emitting events.
// It is a no-op if playback is currently paused.
func (p *Player) Pause() error {
	p.setPlaying(false)
	return nil
}

// Play starts emitting events. It is used to start playback
// for the first time and to resume playing after the player
// is paused.
func (p *Player) Play() error {
	p.setPlaying(true)
	return nil
}

func (p *Player) setPlaying(play bool) {
	// Inspired by "Rethinking Classical Concurrency Patterns"
	// by Bryan C. Mills (GopherCon 2018)
	// <link here>

	ch := <-p.playPause
	alreadyPlaying := ch == nil

	if alreadyPlaying && !play {
		ch = make(chan struct{})
	} else if !alreadyPlaying && play {
		// signal waiters who are paused that it's time to resume playing
		close(ch)
		ch = nil
	}

	p.playPause <- ch
}

// waitWhilePaused blocks while the player is in a paused state.
// It returns immediately if the player is currently playing.
func (p *Player) waitWhilePaused() error {
	ch := <-p.playPause
	p.playPause <- ch

	if alreadyPlaying := ch == nil; !alreadyPlaying {
		select {
		// TODO: add cancelable case
		case <-ch:
		}
	}
	return nil
}

// LastPlayed returns the time of the last played event,
// expressed as milliseconds since the start of the session.
func (p *Player) LastPlayed() int64 {
	return 0
}

// TODO: set playback speed
// TODO: seek forward
// TODO: seek backward
