package store

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTooManyWebhooks = errors.New("too many webhooks in this channel")

// Webhook posts into a text channel from outside (see the sidecar's
// webhooks). Only its token's hash is kept, as for invites.
type Webhook struct {
	ID        int64
	Channel   int64
	Name      string
	Creator   string
	CreatedAt time.Time
	Picture   bool // it has one (see SetWebhookPicture)
}

const webhooksTable = `CREATE TABLE IF NOT EXISTS webhooks (
	id         INTEGER PRIMARY KEY AUTOINCREMENT, -- not used again (a stale delete)
	token_hash TEXT NOT NULL UNIQUE,
	channel    INTEGER NOT NULL, -- the text channel
	name       TEXT NOT NULL,
	creator    TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	picture    BLOB, -- a PNG square (see the files handler); NULL: none
	picture_at INTEGER NOT NULL DEFAULT 0
)`

const webhookSelect = `SELECT id, channel, name, creator, created_at, picture IS NOT NULL FROM webhooks `

func scanWebhook(row interface{ Scan(...any) error }) (Webhook, error) {
	var w Webhook
	var created int64
	err := row.Scan(&w.ID, &w.Channel, &w.Name, &w.Creator, &created, &w.Picture)
	if errors.Is(err, sql.ErrNoRows) {
		return Webhook{}, ErrNotFound
	}
	w.CreatedAt = time.UnixMilli(created).UTC()
	return w, err
}

// AddWebhook keeps a webhook by its token's hash; ErrTooManyWebhooks: its
// channel has limit already.
func (s *Store) AddWebhook(hash string, w Webhook, limit int) (Webhook, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Webhook{}, err
	}
	defer tx.Rollback()
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM webhooks WHERE channel = ?`, w.Channel).Scan(&count); err != nil {
		return Webhook{}, err
	}
	if count >= limit {
		return Webhook{}, ErrTooManyWebhooks
	}
	w.CreatedAt = w.CreatedAt.Truncate(time.Millisecond).UTC()
	if err := tx.QueryRow(`INSERT INTO webhooks (token_hash, channel, name, creator, created_at) VALUES (?, ?, ?, ?, ?) RETURNING id`,
		hash, w.Channel, w.Name, w.Creator, w.CreatedAt.UnixMilli()).Scan(&w.ID); err != nil {
		return Webhook{}, err
	}
	return w, tx.Commit()
}

// Webhooks lists a channel's webhooks, the oldest first.
func (s *Store) Webhooks(channel int64) ([]Webhook, error) {
	rows, err := s.db.Query(webhookSelect+`WHERE channel = ? ORDER BY id`, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Webhook{}
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// WebhookByToken is the webhook with this id and token hash (ErrNotFound:
// none).
func (s *Store) WebhookByToken(id int64, hash string) (Webhook, error) {
	return scanWebhook(s.db.QueryRow(webhookSelect+`WHERE id = ? AND token_hash = ?`, id, hash))
}

// RenameWebhook renames a webhook and says whose channel it is.
func (s *Store) RenameWebhook(id int64, name string) (channel int64, err error) {
	return webhookChannel(s.db.QueryRow(`UPDATE webhooks SET name = ? WHERE id = ? RETURNING channel`, name, id))
}

// DeleteWebhook deletes a webhook and says whose channel it was.
func (s *Store) DeleteWebhook(id int64) (channel int64, err error) {
	return webhookChannel(s.db.QueryRow(`DELETE FROM webhooks WHERE id = ? RETURNING channel`, id))
}

// WebhookChannel says whose channel a webhook is.
func (s *Store) WebhookChannel(id int64) (int64, error) {
	return webhookChannel(s.db.QueryRow(`SELECT channel FROM webhooks WHERE id = ?`, id))
}

func webhookChannel(row *sql.Row) (channel int64, err error) {
	if err = row.Scan(&channel); errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return channel, err
}

// SetWebhookPicture sets (nil: removes) a webhook's picture and says whose
// channel it is.
func (s *Store) SetWebhookPicture(id int64, png []byte, at time.Time) (channel int64, err error) {
	return webhookChannel(s.db.QueryRow(`UPDATE webhooks SET picture = ?, picture_at = ? WHERE id = ? RETURNING channel`, png, at.UnixMilli(), id))
}

// WebhookPicture is a webhook's picture and when it was set (ErrNotFound:
// none).
func (s *Store) WebhookPicture(id int64) ([]byte, time.Time, error) {
	var png []byte
	var at int64
	err := s.db.QueryRow(`SELECT picture, picture_at FROM webhooks WHERE id = ? AND picture IS NOT NULL`, id).Scan(&png, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, time.Time{}, ErrNotFound
	}
	return png, time.UnixMilli(at).UTC(), err
}
