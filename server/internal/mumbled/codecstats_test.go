package mumbled

import (
	"testing"
	"time"
)

// The first byte of an Opus packet (RFC 6716, 3.1): config << 3 | stereo << 2 | frames.
func toc(config byte, stereo bool) byte {
	b := config << 3
	if stereo {
		b |= 4
	}
	return b
}

func TestOpusTOC(t *testing.T) {
	for _, c := range []struct {
		config byte
		band   string
		frame  time.Duration
	}{
		{1, "narrow", 20 * time.Millisecond},      // SILK
		{5, "medium", 20 * time.Millisecond},      // SILK
		{9, "wide", 20 * time.Millisecond},        // SILK
		{13, "super-wide", 20 * time.Millisecond}, // hybrid
		{15, "full", 20 * time.Millisecond},       // hybrid
		{19, "narrow", 20 * time.Millisecond},     // CELT
		{23, "wide", 20 * time.Millisecond},
		{27, "super-wide", 20 * time.Millisecond},
		{31, "full", 20 * time.Millisecond},
		{28, "full", 2500 * time.Microsecond},
		{0, "narrow", 10 * time.Millisecond},
		{3, "narrow", 60 * time.Millisecond},
	} {
		band, stereo, frame := opusTOC(toc(c.config, true))
		if bands[band] != c.band || !stereo || frame != c.frame {
			t.Errorf("config %d: %s %v %v", c.config, bands[band], stereo, frame)
		}
	}
}

// What a speaker sent lately: the band most of its packets had, the last
// packet's channels and frame length, and the bitrate of the Opus data.
func TestRecentCodecStats(t *testing.T) {
	var m connStats
	now := time.Unix(100, 0)
	if c := m.codec(); c.Band != "" {
		t.Fatalf("no voice yet: %+v", c)
	}
	for i := range 300 {
		cfg := byte(31) // CELT full band, 20 ms
		if i%10 == 0 {
			cfg = 9 // a few SILK wide band
		}
		opus := make([]byte, 100) // 100 bytes per 20 ms: 40 kbps
		opus[0] = toc(cfg, false)
		m.voice(audioPacket{opus: opus, frame: uint64(i * 2)}, true, len(opus)+4, now)
	}
	c := m.codec()
	if c.Band != "full" || c.Stereo || c.FrameMs != 20 || c.Kbps != 40 {
		t.Fatalf("%+v", c)
	}
}

func TestConnectionsShowTheVoiceCodec(t *testing.T) {
	st := ConnStats{Session: 3, Name: "niffwe", Codec: Codec{Band: "narrow", Stereo: false, FrameMs: 20, Kbps: 12.5}}
	c := connection(st, sample{})
	if c.VoiceBand != "narrow" || c.VoiceFrameMs != 20 || c.VoiceKbps != 12.5 || c.VoiceStereo {
		t.Fatalf("%+v", c)
	}
}
