package store

import (
	"database/sql"
	"errors"
)

const channelListTable = `CREATE TABLE IF NOT EXISTS channel_lists (
	user_id INTEGER PRIMARY KEY, -- the Mumble user id
	list    TEXT NOT NULL        -- JSON, checked by the sidecar
)`

// SetChannelList keeps a registered user's own channel list (hidden
// channels and an own order, see the sidecar's channelList).
func (s *Store) SetChannelList(userID uint32, list string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO channel_lists (user_id, list) VALUES (?, ?)`, userID, list)
	return err
}

// ChannelList is a user's kept channel list ("" if there is none).
func (s *Store) ChannelList(userID uint32) (string, error) {
	var list string
	err := s.db.QueryRow(`SELECT list FROM channel_lists WHERE user_id = ?`, userID).Scan(&list)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return list, err
}
