package store

import (
	"database/sql"
	"errors"
	"time"
)

// Stickers: large pictures sent as a message of their own. Kept like
// personal emoji images: a deleted sticker keeps its row without the
// image, so old messages still show its name; ids are never used again,
// so an image URL by id can be cached.

// MaxStickers is the most stickers a server has (deleted ones aside).
const MaxStickers = 200

var ErrTooManyStickers = errors.New("too many stickers")

// Sticker is a server sticker; Data only from StickerByID.
type Sticker struct {
	ID          int64
	Name        string
	Description string
	Emoji       string // a related emoji ("": none)
	Pack        string // its group in the picker ("": none)
	Type        string
	Data        []byte
	Author      string
	CreatedAt   time.Time
	Deleted     bool
}

const stickersTable = `CREATE TABLE IF NOT EXISTS stickers (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	emoji       TEXT NOT NULL DEFAULT '',
	pack        TEXT NOT NULL DEFAULT '',
	type        TEXT NOT NULL,
	data        BLOB,
	author      TEXT NOT NULL,
	created_at  INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS stickers_name ON stickers (name) WHERE data IS NOT NULL`

// AddSticker stores a sticker (the caller checks name and image).
func (s *Store) AddSticker(st Sticker, at time.Time) (Sticker, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Sticker{}, err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRow(`SELECT count(*) FROM stickers WHERE data IS NOT NULL`).Scan(&n); err != nil {
		return Sticker{}, err
	}
	if n >= MaxStickers {
		return Sticker{}, ErrTooManyStickers
	}
	err = tx.QueryRow(`INSERT INTO stickers (name, description, emoji, pack, type, data, author, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		st.Name, st.Description, st.Emoji, st.Pack, st.Type, st.Data, st.Author, at.UnixMilli()).Scan(&st.ID)
	if err != nil {
		return Sticker{}, nameTaken(err)
	}
	st.CreatedAt = at
	return st, tx.Commit()
}

// Stickers lists the stickers that are not deleted, by name, without
// their images.
func (s *Store) Stickers() ([]Sticker, error) {
	rows, err := s.db.Query(`SELECT id, name, description, emoji, pack, type, author, created_at FROM stickers WHERE data IS NOT NULL ORDER BY pack, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Sticker
	for rows.Next() {
		var st Sticker
		var at int64
		if err := rows.Scan(&st.ID, &st.Name, &st.Description, &st.Emoji, &st.Pack, &st.Type, &st.Author, &at); err != nil {
			return nil, err
		}
		st.CreatedAt = time.UnixMilli(at).UTC()
		out = append(out, st)
	}
	return out, rows.Err()
}

// StickerByID is a sticker with its image (none, and Deleted, once
// deleted).
func (s *Store) StickerByID(id int64) (Sticker, error) {
	var st Sticker
	var at int64
	err := s.db.QueryRow(`SELECT id, name, description, emoji, pack, type, data, author, created_at FROM stickers WHERE id = ?`, id).
		Scan(&st.ID, &st.Name, &st.Description, &st.Emoji, &st.Pack, &st.Type, &st.Data, &st.Author, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return Sticker{}, ErrNotFound
	}
	st.CreatedAt = time.UnixMilli(at).UTC()
	st.Deleted = err == nil && st.Data == nil
	return st, err
}

// UpdateSticker changes a sticker's name, description, emoji and pack.
func (s *Store) UpdateSticker(st Sticker) error {
	return rowChanged(s.db.Exec(`UPDATE stickers SET name = ?, description = ?, emoji = ?, pack = ? WHERE id = ? AND data IS NOT NULL`, st.Name, st.Description, st.Emoji, st.Pack, st.ID))
}

// DeleteSticker removes a sticker's image; its row and name stay.
func (s *Store) DeleteSticker(id int64) error {
	return rowChanged(s.db.Exec(`UPDATE stickers SET data = NULL WHERE id = ? AND data IS NOT NULL`, id))
}
