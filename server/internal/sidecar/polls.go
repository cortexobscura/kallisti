package sidecar

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"kallisti.chat/server/internal/mumble"
	"kallisti.chat/server/internal/store"
)

// Polls (see store.Poll): a message of its own, created with send. Its
// text (question and answers) is for what shows only text: Mumble, search,
// quotes, push. Registered users vote while it runs; the author ends it
// early, the minute loop on time (EndDuePolls), each with a result message.

const (
	maxPollQuestion = 300 // characters
	maxPollAnswer   = 55
	minPollAnswers  = 2
	maxPollAnswers  = 10
	maxPollHours    = 14 * 24
)

type wirePollRequest struct {
	Question  string           `json:"question"`
	Answers   []wirePollAnswer `json:"answers"`
	Multiple  bool             `json:"multiple"`
	Anonymous bool             `json:"anonymous"`
	Hours     int              `json:"hours"`
}

type wirePollAnswer struct {
	Text   string   `json:"text"`
	Emoji  string   `json:"emoji,omitempty"`
	Votes  int      `json:"votes"`
	Voters []string `json:"voters,omitempty"` // named polls
}

type wirePoll struct {
	Question  string           `json:"question"`
	Answers   []wirePollAnswer `json:"answers"`
	Multiple  bool             `json:"multiple,omitempty"`
	Anonymous bool             `json:"anonymous,omitempty"`
	Ends      int64            `json:"ends"`            // Unix milliseconds
	Ended     int64            `json:"ended,omitempty"` // Unix milliseconds
	Voters    int              `json:"voters"`          // who voted (for shares in multiple choice)
	// The reader's own answers, where the message goes to one reader (else
	// missing: the client keeps what it knew).
	Mine *[]int `json:"mine,omitempty"`
}

// newPoll checks a poll request: the poll for the store, or why not.
// usable: an answer's emoji in its normal form, or why not (see
// usableEmoji).
func newPoll(r wirePollRequest, now time.Time, usable func(string) (string, string)) (store.Poll, string) {
	p := store.Poll{Question: oneLine(r.Question), Multiple: r.Multiple, Anonymous: r.Anonymous, EndsAt: now.Add(time.Duration(r.Hours) * time.Hour)}
	if p.Question == "" || utf8.RuneCountInString(p.Question) > maxPollQuestion {
		return p, fmt.Sprintf("a poll's question has 1 to %d characters", maxPollQuestion)
	}
	if len(r.Answers) < minPollAnswers || len(r.Answers) > maxPollAnswers {
		return p, fmt.Sprintf("a poll has %d to %d answers", minPollAnswers, maxPollAnswers)
	}
	for _, a := range r.Answers {
		text, e := oneLine(a.Text), strings.TrimSpace(a.Emoji)
		if text == "" || utf8.RuneCountInString(text) > maxPollAnswer {
			return p, fmt.Sprintf("a poll's answer has 1 to %d characters", maxPollAnswer)
		}
		if e != "" {
			var why string
			if e, why = usable(e); why != "" {
				return p, why
			}
		}
		p.Answers = append(p.Answers, store.PollAnswer{Text: text, Emoji: e})
	}
	if r.Hours < 1 || r.Hours > maxPollHours {
		return p, "a poll runs from 1 hour to 2 weeks"
	}
	return p, ""
}

// oneLine: line breaks would fake answers in pollText.
func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// quiet: polls ping nobody. A zero-width space after each "@" keeps the
// text as it looks, but no mention matches it (pushes, @everyone, follows,
// the clients' highlights).
func quiet(s string) string { return strings.ReplaceAll(s, "@", "@\u200b") }

// pollText is a poll as text.
func pollText(p store.Poll) string {
	var b strings.Builder
	b.WriteString("📊 " + p.Question)
	for _, a := range p.Answers {
		b.WriteString("\n• " + strings.TrimSpace(a.Emoji+" "+a.Text))
	}
	b.WriteString("\n(a poll: vote in Kallisti)")
	return quiet(b.String())
}

// toWirePoll: the counts, names only in named polls; viewer: the reader
// (nil: not one reader, or a guest).
func toWirePoll(p store.Poll, viewer *uint32) *wirePoll {
	w := &wirePoll{Question: p.Question, Answers: make([]wirePollAnswer, len(p.Answers)), Multiple: p.Multiple, Anonymous: p.Anonymous, Ends: p.EndsAt.UnixMilli()}
	if !p.EndedAt.IsZero() {
		w.Ended = p.EndedAt.UnixMilli()
	}
	voters := map[uint32]bool{}
	mine := []int{}
	for i, a := range p.Answers {
		w.Answers[i] = wirePollAnswer{Text: a.Text, Emoji: a.Emoji, Votes: len(a.Voters)}
		for _, v := range a.Voters {
			voters[v.ID] = true
			if !p.Anonymous {
				w.Answers[i].Voters = append(w.Answers[i].Voters, v.Name)
			}
			if viewer != nil && v.ID == *viewer {
				mine = append(mine, i)
			}
		}
	}
	w.Voters = len(voters)
	if viewer != nil {
		w.Mine = &mine
	}
	return w
}

// viewerOf is a reader's user id for its own votes (guests: none).
func viewerOf(id mumble.Identity) *uint32 {
	if !id.Registered {
		return nil
	}
	return &id.UserID
}

func (c *Conn) viewer() *uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.identity == nil {
		return nil
	}
	return viewerOf(*c.identity)
}

// sendPoll sends the chat's readers a poll's counts, each with its votes
// (none for a poll deleted meanwhile).
func (h *Hub) sendPoll(ch chat, m store.Message) {
	if m.Poll == nil {
		return
	}
	h.eachReady(func(conn *Conn, id mumble.Identity) {
		if h.may(id, ch, false) {
			conn.send(map[string]any{"type": "poll", "channel": m.ChannelID, "message": m.ID, "poll": toWirePoll(*m.Poll, viewerOf(id))})
		}
	})
}

// pollAction: vote (answers; none takes the vote back) and end-poll (the
// author).
func (c *Conn) pollAction(id mumble.Identity, req request) {
	h := c.hub
	if !id.Registered {
		c.sendError("join with an account (or a guest account) to vote")
		return
	}
	m, err := h.store.MessageByID(req.Message)
	ch, ok := h.chatOf(m.ChannelID)
	if err != nil || m.Poll == nil || !ok || !h.may(id, ch, true) {
		c.sendError("unknown poll")
		return
	}
	if req.Type == "end-poll" {
		if !byUser(id, m) {
			c.sendError("only its author ends a poll early")
			return
		}
		h.endPoll(ch, m.ID)
		return
	}
	m, err = h.store.Vote(m.ID, store.Voter{ID: id.UserID, Name: id.Name}, req.Answers, h.now())
	switch {
	case errors.Is(err, store.ErrPollEnded):
		c.sendError("the poll has ended")
	case errors.Is(err, store.ErrBadVote):
		c.sendError("not a vote for this poll")
	case err != nil:
		c.sendError("cannot store the vote")
	default:
		h.sendPoll(ch, m)
	}
}

// EndDuePolls ends the polls whose time is up (the minute loop).
func (h *Hub) EndDuePolls() {
	due, err := h.store.DuePolls(h.now())
	if err != nil {
		slog.Error("due polls", "err", err)
		return
	}
	for _, id := range due {
		m, err := h.store.MessageByID(id)
		if ch, ok := h.chatOf(m.ChannelID); err == nil && ok {
			h.endPoll(ch, id)
		} else if _, err := h.store.EndPoll(id, h.now()); err == nil { // a chat gone: ended quietly
			slog.Info("poll ended without its chat", "poll", id)
		}
	}
}

// endPoll ends a poll once: its readers get the counts, then the result as
// a message of its author that answers it (no ping), also to Mumble.
func (h *Hub) endPoll(ch chat, id int64) {
	if ended, err := h.store.EndPoll(id, h.now()); err != nil || !ended {
		return
	}
	m, err := h.store.MessageByID(id)
	if err != nil || m.Poll == nil { // deleted meanwhile
		return
	}
	h.sendPoll(ch, m)
	// From the author, but not from a session (Mumble: from the server).
	author := mumble.Identity{Name: m.Author, Registered: m.Verified}
	if m.AuthorID != nil {
		author.UserID = *m.AuthorID
	}
	result, err := h.store.AddMessage(store.NewMessage{ChannelID: m.ChannelID, Thread: ch.parent != 0, Author: m.Author, Verified: m.Verified, AuthorID: author.UserID,
		Text: pollResult(*m.Poll), At: h.now(), ReplyTo: m.ID, NoPing: true})
	if err != nil {
		slog.Error("poll result", "poll", id, "err", err)
		return
	}
	h.deliver(author, ch, result, nil, "")
}

// pollResult names the answers with the most votes, with their share of
// the voters.
func pollResult(p store.Poll) string {
	w := toWirePoll(p, nil)
	best := 0
	for _, a := range w.Answers {
		best = max(best, a.Votes)
	}
	if best == 0 {
		return quiet(fmt.Sprintf("📊 The poll “%s” ended: no votes.", p.Question))
	}
	var top []string
	for _, a := range w.Answers {
		if a.Votes == best {
			top = append(top, strings.TrimSpace(a.Emoji+" "+a.Text))
		}
	}
	votes := "votes"
	if best == 1 {
		votes = "vote"
	}
	return quiet(fmt.Sprintf("📊 The poll “%s” ended: %s (%d %s, %d%%).", p.Question, strings.Join(top, ", "), best, votes, best*100/w.Voters))
}
