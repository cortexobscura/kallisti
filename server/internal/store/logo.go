package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"maps"
	"slices"
)

// The server's logo (admins upload it): one PNG per icon size, and a
// version (from the pictures) for URLs that browsers may cache for ever.
const logoTable = `CREATE TABLE IF NOT EXISTS server_logo (
	size    INTEGER PRIMARY KEY,
	png     BLOB NOT NULL,
	version TEXT NOT NULL
)`

// SetLogo replaces the logo with these pictures (by size); it gives the
// new version.
func (s *Store) SetLogo(icons map[int][]byte) (string, error) {
	h := sha256.New()
	sizes := slices.Sorted(maps.Keys(icons))
	for _, size := range sizes {
		h.Write(icons[size])
	}
	version := hex.EncodeToString(h.Sum(nil))[:16]
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM server_logo`); err != nil {
		return "", err
	}
	for _, size := range sizes {
		if _, err := tx.Exec(`INSERT INTO server_logo (size, png, version) VALUES (?, ?, ?)`, size, icons[size], version); err != nil {
			return "", err
		}
	}
	return version, tx.Commit()
}

// Logo is the logo's picture of this size and its version (ErrNotFound:
// none).
func (s *Store) Logo(size int) (png []byte, version string, err error) {
	err = s.db.QueryRow(`SELECT png, version FROM server_logo WHERE size = ?`, size).Scan(&png, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	return png, version, err
}

// LogoVersion is the logo's version ("": no logo).
func (s *Store) LogoVersion() (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT version FROM server_logo LIMIT 1`).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

func (s *Store) DeleteLogo() error {
	_, err := s.db.Exec(`DELETE FROM server_logo`)
	return err
}
