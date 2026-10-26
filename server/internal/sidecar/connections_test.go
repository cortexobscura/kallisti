package sidecar

import (
	"errors"
	"strings"
	"testing"
)

// Session 1 is an admin.
func (f *fakeVoice) Connections(session uint32) ([]Connection, error) {
	if session != 1 {
		return nil, errors.New("only admins see the connections")
	}
	ping := 42.0
	return []Connection{{Session: 2, Name: "bob", Transport: "udp", Ping: &ping, Online: 60}}, nil
}

func TestConnectionsAreForAdmins(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin := readyAs(t, h, 1, "admin")
	bob := readyAs(t, h, 2, "bob")
	bob.Handle([]byte(`{"type":"connections"}`))
	if m := next(t, bob); m["type"] != "error" || m["about"] != "manage" {
		t.Fatalf("bob got %v", m)
	}
	admin.Handle([]byte(`{"type":"connections"}`))
	m := next(t, admin)
	conns, _ := m["connections"].([]any)
	if m["type"] != "connections" || len(conns) != 1 {
		t.Fatalf("admin got %v", m)
	}
	c := conns[0].(map[string]any)
	if c["name"] != "bob" || c["ping"] != 42.0 || c["fromClient"] != nil || c["online"] != 60.0 || c["address"] != nil {
		t.Fatalf("connection %v", c)
	}
}

// A web client's microphone, as its browser says, for the connections view.
func TestMicrophoneInfoInTheConnections(t *testing.T) {
	h := newHub(t, "general")
	h.SetVoice(&fakeVoice{}, "")
	admin := readyAs(t, h, 1, "admin")
	bob := readyAs(t, h, 2, "bob")
	for _, bad := range []string{
		`{"type":"mic-info","mic":{"name":"` + strings.Repeat("x", 101) + `","sampleRate":48000,"channels":1,"filters":[]}}`,
		`{"type":"mic-info","mic":{"name":"m","sampleRate":-1,"channels":1,"filters":[]}}`,
		`{"type":"mic-info","mic":{"name":"m","sampleRate":48000,"channels":1,"filters":["a","b","c","d","e","f","g","h","i"]}}`,
	} {
		bob.Handle([]byte(bad))
		if m := next(t, bob); m["type"] != "error" {
			t.Fatalf("%s: %v", bad, m)
		}
	}
	bob.Handle([]byte(`{"type":"mic-info","mic":{"name":"Headset (Bluetooth)","sampleRate":16000,"channels":1,"filters":["browser noise suppression","echo cancellation"]}}`))
	admin.Handle([]byte(`{"type":"connections"}`))
	c := next(t, admin)["connections"].([]any)[0].(map[string]any)
	if c["micName"] != "Headset (Bluetooth)" || c["micRate"] != 16000.0 || c["micChannels"] != 1.0 || len(c["micFilters"].([]any)) != 2 {
		t.Fatalf("connection %v", c)
	}
	// Gone with the connection.
	bob.Close()
	admin.Handle([]byte(`{"type":"connections"}`))
	if c := next(t, admin)["connections"].([]any)[0].(map[string]any); c["micName"] != nil {
		t.Fatalf("after bob left %v", c)
	}
}
