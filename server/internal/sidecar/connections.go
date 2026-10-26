package sidecar

import (
	"slices"
	"unicode/utf8"

	"kallisti.chat/server/internal/mumble"
)

// Connection is one connection over the recent window, for the admin view
// (see mumbled.Connection, which converts to it).
type Connection struct {
	Session       uint32   `json:"session"`
	Name          string   `json:"name"`
	Transport     string   `json:"transport"`
	Ping          *float64 `json:"ping"`       // ms (null: none measured)
	FromClient    *float64 `json:"fromClient"` // late or lost voice in percent (null: too little voice)
	FromServer    *float64 `json:"fromServer"`
	Dropped       uint64   `json:"dropped"`
	TextOverflows uint64   `json:"overflows"`
	Online        int64    `json:"online"` // seconds
	Client        string   `json:"client"`
	OS            string   `json:"os"`
	// Its voice lately: the Opus band ("narrow" … "full"; none yet: left
	// out), stereo, frame length and the bitrate of the Opus data.
	VoiceBand    string  `json:"voiceBand,omitempty"`
	VoiceStereo  bool    `json:"voiceStereo,omitempty"`
	VoiceFrameMs float64 `json:"voiceFrameMs,omitempty"`
	VoiceKbps    float64 `json:"voiceKbps,omitempty"`
	// Stream frames it missed as a viewer, its connection full.
	StreamPictures uint64 `json:"streamPictures"`
	StreamSound    uint64 `json:"streamSound"`
	// Its microphone, as a web client's browser says (see wireMic).
	MicName     string   `json:"micName,omitempty"`
	MicRate     int      `json:"micRate,omitempty"`
	MicChannels int      `json:"micChannels,omitempty"`
	MicFilters  []string `json:"micFilters,omitempty"`
}

// wireMic: what a web client's browser says about its microphone, for the
// admin view (a sample rate of 16 or 8 kHz: a Bluetooth headset in its
// phone mode), and the filters that run.
type wireMic struct {
	Name       string   `json:"name"`
	SampleRate int      `json:"sampleRate"`
	Channels   int      `json:"channels"`
	Filters    []string `json:"filters"`
}

const (
	maxMicName    = 100
	maxMicFilters = 8
	maxMicFilter  = 40
)

func (c *Conn) setMic(m *wireMic) {
	ok := m != nil && utf8.RuneCountInString(m.Name) <= maxMicName && m.SampleRate >= 0 && m.SampleRate <= 384_000 &&
		m.Channels >= 0 && m.Channels <= 32 && len(m.Filters) <= maxMicFilters &&
		!slices.ContainsFunc(m.Filters, func(f string) bool { return utf8.RuneCountInString(f) > maxMicFilter })
	if !ok {
		c.sendError("not a microphone description")
		return
	}
	c.mu.Lock()
	c.mic = m
	c.mu.Unlock()
}

// micOf: the microphone of a session's web connection (nil: none said).
func (h *Hub) micOf(session uint32) *wireMic {
	var mic *wireMic
	h.eachReady(func(conn *Conn, id mumble.Identity) {
		if id.Session == session {
			conn.mu.Lock()
			mic = conn.mic
			conn.mu.Unlock()
		}
	})
	return mic
}

// connections answers an admin's poll of the connections table.
func (c *Conn) connections(session uint32) {
	if c.hub.voice == nil {
		c.manageError("connections need the built-in Mumble server")
		return
	}
	conns, err := c.hub.voice.Connections(session)
	if err != nil {
		c.manageError(err.Error())
		return
	}
	if conns == nil {
		conns = []Connection{}
	}
	for i, conn := range conns {
		if m := c.hub.micOf(conn.Session); m != nil {
			conns[i].MicName, conns[i].MicRate, conns[i].MicChannels, conns[i].MicFilters = m.Name, m.SampleRate, m.Channels, m.Filters
		}
	}
	c.send(map[string]any{"type": "connections", "connections": conns})
}
