package mumbled

import "time"

// What a speaker's voice looks like at the server, for the admin view of
// connections: the Opus bandwidth (narrow band sounds muffled, often a
// Bluetooth headset in its phone mode), channels, frame length and
// bitrate, over its last recentPackets packets (about 5 s of speech).

var bands = [...]string{"narrow", "medium", "wide", "super-wide", "full"}

// opusTOC reads an Opus packet's first byte (RFC 6716, 3.1): the index of
// its band in bands, stereo, and the length of one frame.
func opusTOC(b byte) (band int, stereo bool, frame time.Duration) {
	config := int(b >> 3)
	stereo = b&4 != 0
	switch {
	case config < 12: // SILK: narrow, medium, wide; 10, 20, 40, 60 ms
		return config / 4, stereo, []time.Duration{10, 20, 40, 60}[config%4] * time.Millisecond
	case config < 16: // hybrid: super-wide, full; 10, 20 ms
		return 3 + (config-12)/2, stereo, []time.Duration{10, 20}[config%2] * time.Millisecond
	default: // CELT: narrow, wide, super-wide, full; 2.5, 5, 10, 20 ms
		return []int{0, 2, 3, 4}[(config-16)/4], stereo, []time.Duration{2500, 5000, 10000, 20000}[config%4] * time.Microsecond
	}
}

const recentPackets = 250

type packetCodec struct {
	band  uint8
	bytes uint16
	frame time.Duration
}

// codecRing holds the last recentPackets packets' codec (under connStats.mu).
type codecRing struct {
	packets [recentPackets]packetCodec
	next, n int
	stereo  bool // the last packet's
}

func (r *codecRing) add(opus []byte) {
	if len(opus) == 0 {
		return
	}
	band, stereo, frame := opusTOC(opus[0])
	r.packets[r.next] = packetCodec{band: uint8(band), bytes: uint16(min(len(opus), 1<<16-1)), frame: frame}
	r.next, r.n, r.stereo = (r.next+1)%recentPackets, min(r.n+1, recentPackets), stereo
}

// Codec is a speaker's recent voice (Band "": none yet).
type Codec struct {
	Band    string
	Stereo  bool
	FrameMs float64
	Kbps    float64 // of the Opus data, without headers
}

func (r *codecRing) codec() Codec {
	if r.n == 0 {
		return Codec{}
	}
	var count [len(bands)]int
	var bytes int
	var length time.Duration
	for _, p := range r.packets[:r.n] {
		count[p.band]++
		bytes += int(p.bytes)
		length += p.frame
	}
	most := 0
	for i := range count {
		if count[i] > count[most] {
			most = i
		}
	}
	last := r.packets[(r.next+recentPackets-1)%recentPackets]
	return Codec{Band: bands[most], Stereo: r.stereo, FrameMs: float64(last.frame) / float64(time.Millisecond),
		Kbps: float64(bytes*8) / length.Seconds() / 1000}
}

// codec: what the connection's voice looked like lately.
func (m *connStats) codec() Codec {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.recent.codec()
}
