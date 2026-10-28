package store

import (
	"errors"
)

// The servers an account joined (the rail of several servers): the home
// server keeps the list, every device of the account shows it.
const accountServersTable = `CREATE TABLE IF NOT EXISTS account_servers (
	account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	origin     TEXT NOT NULL,
	name       TEXT NOT NULL,
	position   INTEGER NOT NULL,
	PRIMARY KEY (account_id, origin)
)`

var ErrTooManyServers = errors.New("too many servers")

// AccountServer is a server in an account's list, by its origin, with the
// name the account has there.
type AccountServer struct {
	Origin string `json:"origin"`
	Name   string `json:"name"`
}

// AccountServers is the account's list, in the order added.
func (s *Store) AccountServers(accountID int64) ([]AccountServer, error) {
	rows, err := s.db.Query(`SELECT origin, name FROM account_servers WHERE account_id = ? ORDER BY position`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []AccountServer{}
	for rows.Next() {
		var x AccountServer
		if err := rows.Scan(&x.Origin, &x.Name); err != nil {
			return nil, err
		}
		list = append(list, x)
	}
	return list, rows.Err()
}

// AddAccountServer adds it at the end (already there: a new name, the same
// place); ErrTooManyServers: the list has limit servers.
func (s *Store) AddAccountServer(accountID int64, x AccountServer, limit int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE account_servers SET name = ? WHERE account_id = ? AND origin = ?`, x.Name, accountID, x.Origin)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var count, last int
		if err := tx.QueryRow(`SELECT COUNT(*), COALESCE(MAX(position), 0) FROM account_servers WHERE account_id = ?`, accountID).Scan(&count, &last); err != nil {
			return err
		}
		if count >= limit {
			return ErrTooManyServers
		}
		if _, err := tx.Exec(`INSERT INTO account_servers (account_id, origin, name, position) VALUES (?, ?, ?, ?)`, accountID, x.Origin, x.Name, last+1); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RemoveAccountServer takes it out of the list (not there: no error).
func (s *Store) RemoveAccountServer(accountID int64, origin string) error {
	_, err := s.db.Exec(`DELETE FROM account_servers WHERE account_id = ? AND origin = ?`, accountID, origin)
	return err
}
