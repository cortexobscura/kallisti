package sidecar

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"kallisti.chat/server/internal/msgrules"
	"kallisti.chat/server/internal/mumble"
	"kallisti.chat/server/internal/store"
)

const pizzaPoll = `{"question":"Pizza?","answers":[{"text":"Yes","emoji":"🍕"},{"text":"No"}],"hours":1}`

func sendPoll(t *testing.T, c *Conn, poll string) map[string]any {
	t.Helper()
	c.Handle([]byte(`{"type":"send","channel":1,"text":"","poll":` + poll + `}`))
	return next(t, c)
}

func pollOf(m map[string]any) map[string]any {
	if msg, ok := m["message"].(map[string]any); ok && m["type"] != "poll" {
		p, _ := msg["poll"].(map[string]any)
		return p
	}
	p, _ := m["poll"].(map[string]any)
	return p
}

// "votes:voters" for each answer.
func tally(p map[string]any) string {
	out := []string{}
	for _, a := range p["answers"].([]any) {
		a := a.(map[string]any)
		out = append(out, fmt.Sprint(a["votes"], ":", a["voters"]))
	}
	return strings.Join(out, " ")
}

func pollUsers(t *testing.T, h *Hub) (alice, bob *Conn) {
	alice = joinAs(t, h, mumble.Identity{Session: 1, Name: "alice", Registered: true, UserID: 7})
	bob = joinAs(t, h, mumble.Identity{Session: 2, Name: "bob", Registered: true, UserID: 8})
	return alice, bob
}

func TestCreatePolls(t *testing.T) {
	h := newHub(t, "general")
	alice, bob := pollUsers(t, h)
	for _, bad := range []string{
		`{"question":"Q","answers":[{"text":"only"}],"hours":1}`,
		`{"question":"Q","answers":[` + strings.Repeat(`{"text":"a"},`, 10) + `{"text":"a"}],"hours":1}`,
		`{"question":"` + strings.Repeat("q", 301) + `","answers":[{"text":"a"},{"text":"b"}],"hours":1}`,
		`{"question":"Q","answers":[{"text":"` + strings.Repeat("a", 56) + `"},{"text":"b"}],"hours":1}`,
		`{"question":"Q","answers":[{"text":" "},{"text":"b"}],"hours":1}`,
		`{"question":" ","answers":[{"text":"a"},{"text":"b"}],"hours":1}`,
		`{"question":"Q","answers":[{"text":"a"},{"text":"b"}],"hours":0}`,
		`{"question":"Q","answers":[{"text":"a"},{"text":"b"}],"hours":337}`,
	} {
		if m := sendPoll(t, alice, bad); m["type"] != "error" {
			t.Fatalf("%s: %v", bad, m)
		}
	}
	m := sendPoll(t, alice, pizzaPoll)
	p := pollOf(m)
	if m["type"] != "message" || p["question"] != "Pizza?" || p["ends"] != float64(clock.Add(time.Hour).UnixMilli()) || tally(p) != "0:<nil> 0:<nil>" {
		t.Fatalf("got %v", m)
	}
	// For what knows only text (Mumble, search, push).
	if text := m["message"].(map[string]any)["text"]; !strings.Contains(text.(string), "Pizza?") || !strings.Contains(text.(string), "🍕 Yes") {
		t.Fatalf("text %q", text)
	}
	next(t, bob)
	// Guests do not create polls; a poll comes on its own.
	guest := joinAs(t, h, mumble.Identity{Session: 3, Name: "guest"})
	drain(alice, bob)
	if m := sendPoll(t, guest, pizzaPoll); m["type"] != "error" {
		t.Fatalf("guest: %v", m)
	}
	alice.Handle([]byte(`{"type":"send","channel":1,"text":"","sticker":1,"poll":` + pizzaPoll + `}`))
	if m := next(t, alice); m["type"] != "error" {
		t.Fatalf("with a sticker: %v", m)
	}
	// Not editable.
	id := fmt.Sprint(m["message"].(map[string]any)["id"])
	alice.Handle([]byte(`{"type":"edit-message","message":` + id + `,"text":"x"}`))
	if m := next(t, alice); m["type"] != "error" {
		t.Fatalf("edited: %v", m)
	}
}

func TestVoting(t *testing.T) {
	for _, anonymous := range []bool{false, true} {
		h := newHub(t, "general")
		alice, bob := pollUsers(t, h)
		poll := strings.Replace(pizzaPoll, `"hours"`, fmt.Sprintf(`"anonymous":%v,"hours"`, anonymous), 1)
		id := fmt.Sprint(sendPoll(t, alice, poll)["message"].(map[string]any)["id"])
		next(t, bob)
		bob.Handle([]byte(`{"type":"vote","message":` + id + `,"answers":[1]}`))
		mine, theirs := next(t, bob), next(t, alice)
		want := "0:<nil> 1:[bob]"
		if anonymous {
			want = "0:<nil> 1:<nil>"
		}
		if mine["type"] != "poll" || tally(pollOf(mine)) != want || fmt.Sprint(pollOf(mine)["mine"]) != "[1]" || pollOf(mine)["voters"] != 1.0 {
			t.Fatalf("anonymous %v: bob got %v", anonymous, mine)
		}
		if fmt.Sprint(pollOf(theirs)["mine"]) != "[]" || tally(pollOf(theirs)) != want {
			t.Fatalf("anonymous %v: alice got %v", anonymous, theirs)
		}
		// The history has each reader's own votes.
		bob.Handle([]byte(`{"type":"history","channel":1}`))
		if h := next(t, bob)["messages"].([]any); fmt.Sprint(h[0].(map[string]any)["poll"].(map[string]any)["mine"]) != "[1]" {
			t.Fatalf("history: %v", h)
		}
		// Single choice; taken back.
		bob.Handle([]byte(`{"type":"vote","message":` + id + `,"answers":[0,1]}`))
		if m := next(t, bob); m["type"] != "error" {
			t.Fatalf("two answers: %v", m)
		}
		bob.Handle([]byte(`{"type":"vote","message":` + id + `,"answers":[]}`))
		if m := next(t, bob); tally(pollOf(m)) != "0:<nil> 0:<nil>" {
			t.Fatalf("taken back: %v", m)
		}
		next(t, alice)
		guest := joinAs(t, h, mumble.Identity{Session: 3, Name: "guest"})
		drain(alice, bob)
		guest.Handle([]byte(`{"type":"vote","message":` + id + `,"answers":[0]}`))
		if m := next(t, guest); m["type"] != "error" {
			t.Fatalf("guest voted: %v", m)
		}
	}
}

func TestEndingPolls(t *testing.T) {
	h := newHub(t, "general")
	now := clock
	h.now = func() time.Time { return now }
	alice, bob := pollUsers(t, h)
	early := fmt.Sprint(sendPoll(t, alice, pizzaPoll)["message"].(map[string]any)["id"])
	onTime := fmt.Sprint(sendPoll(t, alice, pizzaPoll)["message"].(map[string]any)["id"])
	drain(bob)
	bob.Handle([]byte(`{"type":"vote","message":` + early + `,"answers":[0]}`))
	drain(alice, bob)
	// Only the author ends it early.
	bob.Handle([]byte(`{"type":"end-poll","message":` + early + `}`))
	if m := next(t, bob); m["type"] != "error" {
		t.Fatalf("bob ended it: %v", m)
	}
	alice.Handle([]byte(`{"type":"end-poll","message":` + early + `}`))
	ended, result := next(t, alice), next(t, alice)
	if ended["type"] != "poll" || pollOf(ended)["ended"] != float64(now.UnixMilli()) {
		t.Fatalf("ended: %v", ended)
	}
	r := result["message"].(map[string]any)
	if result["type"] != "message" || r["author"] != "alice" || fmt.Sprint(r["reply"].(map[string]any)["id"]) != early || r["noPing"] != true ||
		!strings.Contains(r["text"].(string), "Yes") || !strings.Contains(r["text"].(string), "1 vote") {
		t.Fatalf("result: %v", result)
	}
	drain(bob)
	bob.Handle([]byte(`{"type":"vote","message":` + early + `,"answers":[1]}`))
	if m := next(t, bob); m["type"] != "error" {
		t.Fatalf("voted after the end: %v", m)
	}
	// On time, once, by the minute loop.
	now = now.Add(time.Hour)
	h.EndDuePolls()
	if m := next(t, alice); m["type"] != "poll" || fmt.Sprint(m["message"]) != onTime || pollOf(m)["ended"] == nil {
		t.Fatalf("on time: %v", m)
	}
	if m := next(t, alice); !strings.Contains(m["message"].(map[string]any)["text"].(string), "no votes") {
		t.Fatalf("result: %v", m)
	}
	h.EndDuePolls()
	empty(t, alice)
}

func TestPollChecks(t *testing.T) {
	h := newHub(t, "general")
	alice, _ := pollUsers(t, h)
	for _, bad := range []string{
		`{"question":"Q","answers":[{"text":"a","emoji":"no emoji"},{"text":"b"}],"hours":1}`,
	} {
		if m := sendPoll(t, alice, bad); m["type"] != "error" {
			t.Fatalf("%s: %v", bad, m)
		}
	}
	// Line breaks would fake answers in the text.
	m := sendPoll(t, alice, `{"question":"Q\n• fake","answers":[{"text":"a\nb"},{"text":"c"}],"hours":1}`)
	if p := pollOf(m); p["question"] != "Q • fake" || p["answers"].([]any)[0].(map[string]any)["text"] != "a b" {
		t.Fatalf("got %v", p)
	}
}

// The rules' word replacements reach the poll itself, not only its text.
func TestPollsGetTheWordReplacements(t *testing.T) {
	h := newHub(t, "general")
	f := &fakeVoice{}
	h.SetVoice(f, "")
	f.rules = msgrules.Rules{Words: []string{}, Domains: []string{}, Replacements: []msgrules.Replacement{{Phrase: "darn", With: []string{"oh"}}}}
	alice := joinAs(t, h, mumble.Identity{Session: 2, Name: "alice", Registered: true, UserID: 7})
	alice.Handle([]byte(`{"type":"send","channel":1,"text":"","poll":{"question":"darn?","answers":[{"text":"darn"},{"text":"no"}],"hours":1}}`))
	m := nextOf(t, alice, "message")
	p := pollOf(m)
	if p["question"] != "oh?" || p["answers"].([]any)[0].(map[string]any)["text"] != "oh" || strings.Contains(m["message"].(map[string]any)["text"].(string), "darn") {
		t.Fatalf("got %v", m)
	}
}

// As reactions: Unicode emoji, the server's, one's personal ones.
func TestPollAnswersWithCustomEmoji(t *testing.T) {
	h := newHub(t, "general")
	alice, _ := pollUsers(t, h)
	addEmoji(t, h, "party")
	h.EmojiChanged()
	img := addUserEmoji(t, h, 7, "cool", "c")
	ref := fmt.Sprintf("<:cool:%d>", img.ImageID)
	drain(alice)
	m := sendPoll(t, alice, `{"question":"Q","answers":[{"text":"a","emoji":":party:"},{"text":"b","emoji":"`+ref+`"},{"text":"c","emoji":"👍"}],"hours":1}`)
	answers := pollOf(m)["answers"].([]any)
	if m["type"] != "message" || answers[0].(map[string]any)["emoji"] != ":party:" || answers[1].(map[string]any)["emoji"] != ref {
		t.Fatalf("got %v", m)
	}
	for _, bad := range []string{":nope:", "<:x:999999>", "no emoji"} {
		if m := sendPoll(t, alice, `{"question":"Q","answers":[{"text":"a","emoji":"`+bad+`"},{"text":"b"}],"hours":1}`); m["type"] != "error" {
			t.Fatalf("%s: %v", bad, m)
		}
	}
}

// Polls ping nobody, nor does their result.
func TestPollsDoNotPing(t *testing.T) {
	h := newHub(t, "general")
	alice, _ := pollUsers(t, h)
	m := sendPoll(t, alice, `{"question":"@bob @everyone?","answers":[{"text":"@bob"},{"text":"b"}],"hours":1}`)
	msg := m["message"].(map[string]any)
	stored := mustMessage(t, h, int64(msg["id"].(float64)))
	if msg["mentionAll"] != nil || pings(stored, false, "bob", 8) || pollOf(m)["question"] != "@bob @everyone?" {
		t.Fatalf("got %v", m)
	}
	result := store.Message{Text: pollResult(*stored.Poll), Author: "alice"}
	if pings(result, false, "bob", 8) {
		t.Fatalf("the result pings: %q", result.Text)
	}
}
