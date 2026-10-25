package store

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func postPoll(t *testing.T, s *Store, p Poll, at time.Time) Message {
	t.Helper()
	m, err := s.AddMessage(NewMessage{ChannelID: 1, Author: "alice", Verified: true, AuthorID: 7, Text: "poll", At: at, Poll: &p})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// votes: "answer:names" for each answer.
func votes(p *Poll) string {
	out := ""
	for i, a := range p.Answers {
		out += fmt.Sprint(i, ":")
		for _, v := range a.Voters {
			out += v.Name
		}
		out += " "
	}
	return out
}

func TestPollsAndVotes(t *testing.T) {
	s := open(t)
	s.EnsureChannels([]string{"general"})
	at := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	m := postPoll(t, s, Poll{Question: "Pizza?", Answers: []PollAnswer{{Text: "Yes", Emoji: "🍕"}, {Text: "No"}, {Text: "Maybe"}}, EndsAt: at.Add(time.Hour)}, at)
	if m.Poll == nil || m.Poll.Question != "Pizza?" || len(m.Poll.Answers) != 3 || m.Poll.Answers[0].Emoji != "🍕" || !m.Poll.EndsAt.Equal(at.Add(time.Hour)) {
		t.Fatalf("added %+v", m.Poll)
	}
	bob, carol := Voter{ID: 8, Name: "b"}, Voter{ID: 9, Name: "c"}
	for _, bad := range [][]int{{0, 1}, {3}, {-1}} { // single choice; out of range
		if _, err := s.Vote(m.ID, bob, bad, at); !errors.Is(err, ErrBadVote) {
			t.Fatalf("%v: %v", bad, err)
		}
	}
	s.Vote(m.ID, bob, []int{0}, at)
	s.Vote(m.ID, carol, []int{0}, at)
	got, err := s.Vote(m.ID, bob, []int{2}, at) // changed
	if err != nil || votes(got.Poll) != "0:c 1: 2:b " {
		t.Fatalf("votes %s %v", votes(got.Poll), err)
	}
	got, _ = s.Vote(m.ID, carol, nil, at) // taken back
	if votes(got.Poll) != "0: 1: 2:b " {
		t.Fatalf("taken back: %s", votes(got.Poll))
	}
	for _, read := range []func() (Message, error){
		func() (Message, error) { return s.MessageByID(m.ID) },
		func() (Message, error) { h, err := s.History(1, 0, 10); return h[0], err },
	} {
		if r, err := read(); err != nil || r.Poll == nil || votes(r.Poll) != "0: 1: 2:b " {
			t.Fatalf("read back %+v %v", r.Poll, err)
		}
	}
	if _, err := s.Vote(9999, bob, []int{0}, at); !errors.Is(err, ErrNotFound) {
		t.Fatalf("no poll: %v", err)
	}
}

func TestMultipleChoice(t *testing.T) {
	s := open(t)
	s.EnsureChannels([]string{"general"})
	at := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	m := postPoll(t, s, Poll{Question: "Which?", Answers: []PollAnswer{{Text: "a"}, {Text: "b"}}, Multiple: true, EndsAt: at.Add(time.Hour)}, at)
	got, err := s.Vote(m.ID, Voter{ID: 8, Name: "b"}, []int{1, 0}, at)
	if err != nil || votes(got.Poll) != "0:b 1:b " {
		t.Fatalf("%s %v", votes(got.Poll), err)
	}
	if _, err := s.Vote(m.ID, Voter{ID: 8, Name: "b"}, []int{1, 1}, at); !errors.Is(err, ErrBadVote) {
		t.Fatalf("twice: %v", err)
	}
}

func TestPollsEnd(t *testing.T) {
	s := open(t)
	s.EnsureChannels([]string{"general"})
	at := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	soon := postPoll(t, s, Poll{Question: "a", Answers: []PollAnswer{{Text: "x"}, {Text: "y"}}, EndsAt: at.Add(time.Hour)}, at)
	later := postPoll(t, s, Poll{Question: "b", Answers: []PollAnswer{{Text: "x"}, {Text: "y"}}, EndsAt: at.Add(2 * time.Hour)}, at)
	if due, err := s.DuePolls(at.Add(time.Hour)); err != nil || fmt.Sprint(due) != fmt.Sprint([]int64{soon.ID}) {
		t.Fatalf("due %v %v", due, err)
	}
	if ended, err := s.EndPoll(soon.ID, at.Add(time.Hour)); err != nil || !ended {
		t.Fatalf("end %v %v", ended, err)
	}
	if ended, _ := s.EndPoll(soon.ID, at.Add(time.Hour)); ended {
		t.Fatal("ended twice")
	}
	if due, _ := s.DuePolls(at.Add(3 * time.Hour)); fmt.Sprint(due) != fmt.Sprint([]int64{later.ID}) {
		t.Fatalf("due after the end: %v", due)
	}
	if _, err := s.Vote(soon.ID, Voter{ID: 8, Name: "b"}, []int{0}, at.Add(time.Hour)); !errors.Is(err, ErrPollEnded) {
		t.Fatalf("vote after the end: %v", err)
	}
	// Past its time but not ended yet by the loop: no votes either.
	if _, err := s.Vote(later.ID, Voter{ID: 8, Name: "b"}, []int{0}, at.Add(2*time.Hour)); !errors.Is(err, ErrPollEnded) {
		t.Fatalf("vote after the time: %v", err)
	}
	if m, _ := s.MessageByID(soon.ID); !m.Poll.EndedAt.Equal(at.Add(time.Hour)) {
		t.Fatalf("ended at %v", m.Poll.EndedAt)
	}
	if _, err := s.DeleteMessage(soon.ID); err != nil {
		t.Fatal(err)
	}
	var left int
	s.db.QueryRow(`SELECT (SELECT count(*) FROM polls WHERE message_id = ?) + (SELECT count(*) FROM poll_votes WHERE message_id = ?)`, soon.ID, soon.ID).Scan(&left)
	if left != 0 {
		t.Fatalf("rows left: %d", left)
	}
}
