package whip

import "sync/atomic"

// The sender's RTCP sender reports map each track's RTP timestamps to its
// own wall clock (NTP). With them the server can tell how far apart its
// clocks put picture and sound of the same moment: they start at each
// track's first arrival, and a picture that comes later than its sound
// (an encoder that starts slowly) is stamped late for the whole stream.

// srClock is one track's latest sender report, and from it the offset of
// the server's timestamps from the sender's clock.
type srClock struct {
	pending atomic.Pointer[senderReport] // from the RTCP goroutine
	offset  atomic.Int64                 // server µs minus sender µs
	known   atomic.Bool
}

type senderReport struct {
	ntp uint64
	rtp uint32
}

// report: a sender report came (any goroutine).
func (s *srClock) report(ntp uint64, rtp uint32) { s.pending.Store(&senderReport{ntp, rtp}) }

// apply takes a pending report on the track's clock (from the goroutine
// that owns the clock).
func (s *srClock) apply(c *clock) {
	r := s.pending.Swap(nil)
	if r == nil || !c.started {
		return
	}
	at := c.base + (c.ticks+int64(int32(r.rtp-c.last)))*1_000_000/c.rate
	s.offset.Store(at - ntpMicros(r.ntp))
	s.known.Store(true)
}

func ntpMicros(ntp uint64) int64 {
	return int64(ntp>>32)*1_000_000 + int64((ntp&0xffffffff)*1_000_000>>32)
}

// avError: how much later than its sound the server stamps a picture of
// the same moment (µs; before the pictures' drift shift), once both
// tracks had a report.
func avError(sound, picture *srClock) (int64, bool) {
	if !sound.known.Load() || !picture.known.Load() {
		return 0, false
	}
	return picture.offset.Load() - sound.offset.Load(), true
}
