package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

// Polls: a message with a question and answers (see NewMessage.Poll).
// The votes keep who voted, also in anonymous polls (one vote each, and
// changes); the sidecar never sends names of those.

const pollsTables = `CREATE TABLE IF NOT EXISTS polls (
	message_id INTEGER PRIMARY KEY,
	question   TEXT NOT NULL,
	answers    TEXT NOT NULL, -- JSON: [{"text", "emoji"}]
	multiple   INTEGER NOT NULL DEFAULT 0,
	anonymous  INTEGER NOT NULL DEFAULT 0,
	ends_at    INTEGER NOT NULL,
	ended_at   INTEGER
);
CREATE INDEX IF NOT EXISTS polls_open ON polls(ends_at) WHERE ended_at IS NULL;
CREATE TABLE IF NOT EXISTS poll_votes (
	message_id INTEGER NOT NULL,
	user_id    INTEGER NOT NULL, -- registered users only
	answer     INTEGER NOT NULL, -- index in answers
	user       TEXT NOT NULL,    -- the name when voting
	created_at INTEGER NOT NULL,
	PRIMARY KEY (message_id, user_id, answer)
)`

type Poll struct {
	Question  string
	Answers   []PollAnswer
	Multiple  bool // more than one answer per voter
	Anonymous bool
	EndsAt    time.Time
	EndedAt   time.Time // zero: still open
}

type PollAnswer struct {
	Text   string  `json:"text"`
	Emoji  string  `json:"emoji,omitempty"`
	Voters []Voter `json:"-"` // in the order they voted
}

type Voter struct {
	ID   uint32
	Name string
}

var (
	ErrBadVote   = errors.New("not a vote for this poll")
	ErrPollEnded = errors.New("the poll has ended")
)

func addPoll(tx *sql.Tx, id int64, p Poll) error {
	answers, err := json.Marshal(p.Answers)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO polls (message_id, question, answers, multiple, anonymous, ends_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, p.Question, string(answers), p.Multiple, p.Anonymous, p.EndsAt.UnixMilli())
	return err
}

// withPolls fills in messages' polls, with their votes.
func withPolls(q interface {
	Query(string, ...any) (*sql.Rows, error)
}, msgs []Message) error {
	if len(msgs) == 0 {
		return nil
	}
	at := map[int64]int{}
	ids := make([]int64, len(msgs))
	for i, m := range msgs {
		at[m.ID], ids[i] = i, m.ID
	}
	in, args := inList(ids)
	rows, err := q.Query(`SELECT message_id, question, answers, multiple, anonymous, ends_at, ended_at FROM polls WHERE message_id IN (`+in+`)`, args...)
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var id, ends int64
		var ended sql.NullInt64
		var answers string
		p := &Poll{}
		if err := rows.Scan(&id, &p.Question, &answers, &p.Multiple, &p.Anonymous, &ends, &ended); err != nil {
			rows.Close()
			return err
		}
		if err := json.Unmarshal([]byte(answers), &p.Answers); err != nil {
			rows.Close()
			return err
		}
		p.EndsAt = time.UnixMilli(ends).UTC()
		if ended.Valid {
			p.EndedAt = time.UnixMilli(ended.Int64).UTC()
		}
		msgs[at[id]].Poll, found = p, true
	}
	rows.Close()
	if err := rows.Err(); err != nil || !found {
		return err
	}
	rows, err = q.Query(`SELECT message_id, answer, user_id, user FROM poll_votes WHERE message_id IN (`+in+`) ORDER BY created_at, rowid`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var answer int
		var v Voter
		if err := rows.Scan(&id, &answer, &v.ID, &v.Name); err != nil {
			return err
		}
		if p := msgs[at[id]].Poll; p != nil && answer >= 0 && answer < len(p.Answers) {
			p.Answers[answer].Voters = append(p.Answers[answer].Voters, v)
		}
	}
	return rows.Err()
}

// Vote replaces a voter's votes in an open poll (none: taken back) and
// returns the message.
func (s *Store) Vote(message int64, v Voter, answers []int, at time.Time) (Message, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Message{}, err
	}
	defer tx.Rollback()
	var raw string
	var multiple bool
	var ends int64
	var ended sql.NullInt64
	err = tx.QueryRow(`SELECT answers, multiple, ends_at, ended_at FROM polls WHERE message_id = ?`, message).Scan(&raw, &multiple, &ends, &ended)
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, ErrNotFound
	} else if err != nil {
		return Message{}, err
	}
	if ended.Valid || at.UnixMilli() >= ends {
		return Message{}, ErrPollEnded
	}
	var list []PollAnswer
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return Message{}, err
	}
	sorted := slices.Sorted(slices.Values(answers))
	if len(answers) > 1 && !multiple || len(slices.Compact(slices.Clone(sorted))) != len(sorted) ||
		len(sorted) > 0 && (sorted[0] < 0 || sorted[len(sorted)-1] >= len(list)) {
		return Message{}, ErrBadVote
	}
	if _, err := tx.Exec(`DELETE FROM poll_votes WHERE message_id = ? AND user_id = ?`, message, v.ID); err != nil {
		return Message{}, err
	}
	for _, a := range answers {
		if _, err := tx.Exec(`INSERT INTO poll_votes (message_id, user_id, answer, user, created_at) VALUES (?, ?, ?, ?, ?)`,
			message, v.ID, a, v.Name, at.UnixMilli()); err != nil {
			return Message{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}
	return s.MessageByID(message)
}

// EndPoll ends an open poll; ended is false if it had ended already (or
// is no poll).
func (s *Store) EndPoll(message int64, at time.Time) (ended bool, err error) {
	res, err := s.db.Exec(`UPDATE polls SET ended_at = ? WHERE message_id = ? AND ended_at IS NULL`, at.UnixMilli(), message)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// DuePolls are the open polls whose time is up, the oldest first.
func (s *Store) DuePolls(now time.Time) ([]int64, error) {
	rows, err := s.db.Query(`SELECT message_id FROM polls WHERE ended_at IS NULL AND ends_at <= ? ORDER BY ends_at`, now.UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
