package whip

import (
	"math"
	"sync/atomic"
)

// lag follows how much later than at its start a track's timestamps fall
// behind their arrival: the least gap per window (network delay only adds
// to it) against the first window's.
type lag struct {
	started, known      bool
	first, least, until int64
	drift               int64
}

const lagWindow = 2_000_000 // µs

// add notes a timestamp and its arrival (µs) and returns the drift (µs),
// once the first window is over.
func (l *lag) add(stamp, arrival int64) (int64, bool) {
	gap := arrival - stamp
	if !l.started {
		l.started, l.least, l.until = true, math.MaxInt64, arrival+lagWindow
	}
	l.least = min(l.least, gap)
	if arrival >= l.until {
		if !l.known {
			l.known, l.first = true, l.least
		}
		l.drift = l.least - l.first
		l.least, l.until = math.MaxInt64, arrival+lagWindow
	}
	return l.drift, l.known
}

// avSync keeps a stream's pictures on its sound's timeline. Senders stamp
// sound by sample count (the capture device's clock) and pictures by the
// system clock, and the two drift apart; viewers play both on one clock and
// drop sound that falls too far behind. Pictures can move a few
// milliseconds unseen; sound cannot. Sound comes from one goroutine,
// pictures from another.
type avSync struct {
	sound      lag // the sound's goroutine
	soundDrift atomic.Int64
	soundKnown atomic.Bool
	pictures   lag   // the pictures' goroutine
	shift      int64 // the pictures', now
}

// maxShiftStep: per picture, at most (µs), so pictures stay in order.
const maxShiftStep = 5_000

func (a *avSync) noteSound(stamp, arrival int64) {
	if d, ok := a.sound.add(stamp, arrival); ok {
		a.soundDrift.Store(d)
		a.soundKnown.Store(true)
	}
}

// picture returns a picture's timestamp on the sound's timeline.
func (a *avSync) picture(stamp, arrival int64) int64 {
	if d, ok := a.pictures.add(stamp, arrival); ok && a.soundKnown.Load() {
		want := a.soundDrift.Load() - d
		a.shift += max(-maxShiftStep, min(maxShiftStep, want-a.shift))
	}
	return max(0, stamp-a.shift)
}

// syncState: the drifts (µs) and the pictures' shift now; known: both
// tracks are past their first window.
type syncState struct {
	sound, pictures, shift int64
	known                  bool
}

// state, from the pictures' goroutine (for the log).
func (a *avSync) state() syncState {
	return syncState{sound: a.soundDrift.Load(), pictures: a.pictures.drift, shift: a.shift, known: a.soundKnown.Load() && a.pictures.known}
}
