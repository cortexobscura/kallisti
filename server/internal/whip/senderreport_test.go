package whip

import "testing"

// ntpAt: the NTP time (64-bit fixed point) of a wall time in µs.
func ntpAt(us int64) uint64 { return uint64(us/1_000_000)<<32 | uint64(us%1_000_000)<<32/1_000_000 }

// Sound and picture of the same moment, but the first picture arrives
// 300 ms after the first sound (an encoder that starts slowly): the
// server's clocks put the picture 300 ms late. The sender reports, on the
// sender's clock, show it.
func TestSenderReportsShowTheOffset(t *testing.T) {
	sound, picture := &clock{rate: 48_000}, &clock{rate: 90_000}
	var sa, sv srClock
	const t0 = int64(3_900_000_000_000_000) // the sender's wall clock (µs)
	// Media time 0 is rtp 1000 (sound) and 5000 (picture).
	sound.micros(1000, 10_000)    // arrives at 10 ms
	picture.micros(5000, 310_000) // arrives at 310 ms
	sound.micros(1000+48_000, 1_010_000)
	picture.micros(5000+90_000, 1_310_000)
	// At media time 1 s (sender wall t0+1 s): sound rtp 49000, picture rtp 95000.
	sa.report(ntpAt(t0+1_000_000), 1000+48_000)
	sv.report(ntpAt(t0+1_000_000), 5000+90_000)
	sa.apply(sound)
	sv.apply(picture)
	got, ok := avError(&sa, &sv)
	if !ok || got < 299_000 || got > 301_000 {
		t.Fatalf("picture minus sound: %d µs (%v), want 300 ms", got, ok)
	}
}

func TestSenderReportsUnknownUntilBoth(t *testing.T) {
	var sa, sv srClock
	sa.report(ntpAt(1_000_000), 1)
	sa.apply(&clock{rate: 48_000, started: true})
	if _, ok := avError(&sa, &sv); ok {
		t.Fatal("known with one report")
	}
}
