package store

import "errors"

const friendsTable = `CREATE TABLE IF NOT EXISTS account_friends (
	account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	origin     TEXT NOT NULL,    -- the friend's server
	user_id    INTEGER NOT NULL, -- the friend's registered user id there
	name       TEXT NOT NULL,    -- the name last seen
	position   INTEGER NOT NULL,
	PRIMARY KEY (account_id, origin, user_id)
)`

var ErrTooManyFriends = errors.New("too many friends")

// Friend is an account on a server (this one or another), in an account's
// friends list: the same person on two servers is two friends.
type Friend struct {
	Origin string `json:"origin"`
	User   uint32 `json:"user"`
	Name   string `json:"name"`
}

// Friends is the account's list, in the order added.
func (s *Store) Friends(accountID int64) ([]Friend, error) {
	rows, err := s.db.Query(`SELECT origin, user_id, name FROM account_friends WHERE account_id = ? ORDER BY position`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Friend{}
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.Origin, &f.User, &f.Name); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

// AddFriend adds one at the end (already there: the new name, the same
// place); ErrTooManyFriends: the list has limit friends.
func (s *Store) AddFriend(accountID int64, f Friend, limit int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE account_friends SET name = ? WHERE account_id = ? AND origin = ? AND user_id = ?`, f.Name, accountID, f.Origin, f.User)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var count, last int
		if err := tx.QueryRow(`SELECT COUNT(*), COALESCE(MAX(position), 0) FROM account_friends WHERE account_id = ?`, accountID).Scan(&count, &last); err != nil {
			return err
		}
		if count >= limit {
			return ErrTooManyFriends
		}
		if _, err := tx.Exec(`INSERT INTO account_friends (account_id, origin, user_id, name, position) VALUES (?, ?, ?, ?, ?)`, accountID, f.Origin, f.User, f.Name, last+1); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RemoveFriend takes one out of the list (not there: no error).
func (s *Store) RemoveFriend(accountID int64, origin string, user uint32) error {
	_, err := s.db.Exec(`DELETE FROM account_friends WHERE account_id = ? AND origin = ? AND user_id = ?`, accountID, origin, user)
	return err
}
