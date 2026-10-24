package whip

import (
	"math/rand/v2"
	"testing"
)

// Sound 0.1 % slow (the capture device's clock), pictures on the system
// clock, both with up to 30 ms of jitter on the way: after the first
// windows, a picture and the sound of the same moment keep about the same
// timestamp for an hour.
func TestPicturesFollowTheSound(t *testing.T) {
	var s avSync
	r := rand.New(rand.NewPCG(1, 2))
	jitter := func() int64 { return r.Int64N(30_000) }
	soundAt := func(real int64) int64 { return real * 999 / 1000 }
	worst := int64(0)
	for real := int64(0); real < 3600_000_000; real += 10_000 {
		if real%20_000 == 0 {
			s.noteSound(soundAt(real), real+jitter())
		}
		if real%40_000 == 0 {
			got := s.picture(real, real+jitter())
			if real > 10_000_000 {
				worst = max(worst, abs(got-soundAt(real)))
			}
		}
	}
	if worst > 50_000 {
		t.Fatalf("a picture %d ms off the sound of its moment", worst/1000)
	}
}

// Picture timestamps only move forward.
func TestPictureShiftKeepsOrder(t *testing.T) {
	var s avSync
	last := int64(-1)
	for real := int64(0); real < 600_000_000; real += 10_000 {
		if real%20_000 == 0 {
			s.noteSound(real*99/100, real) // 1 %: far more than devices drift
		}
		if real%33_000 == 0 {
			got := s.picture(real, real)
			if got <= last {
				t.Fatalf("at %d s: %d after %d", real/1_000_000, got, last)
			}
			last = got
		}
	}
}

func abs(x int64) int64 { return max(x, -x) }

// state reports what the correction does (for the log): the pictures 30 ms
// later on the way than at the start, the sound not.
func TestSyncState(t *testing.T) {
	var s avSync
	for real := int64(0); real < 20_000_000; real += 20_000 {
		s.noteSound(real, real)
		late := int64(0)
		if real >= 4_000_000 {
			late = 30_000
		}
		s.picture(real, real+late)
	}
	st := s.state()
	if !st.known || st.sound != 0 || st.pictures != 30_000 || st.shift != -30_000 {
		t.Fatalf("%+v", st)
	}
}
