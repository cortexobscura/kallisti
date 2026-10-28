package store

import (
	"time"
)

// Kallisti as a login provider (see auth's provider.go): users allow a
// site (its redirect URI) once, and its tokens (by hash, short-lived) read
// who they are. Apps are their redirect URI: nothing kept of them.
const oauthTables = `CREATE TABLE IF NOT EXISTS oauth_grants (
	account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	redirect   TEXT NOT NULL, -- the site's redirect URI
	created_at INTEGER NOT NULL,
	PRIMARY KEY (account_id, redirect)
);
CREATE TABLE IF NOT EXISTS oauth_tokens (
	token_hash TEXT PRIMARY KEY,
	account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	client_id  TEXT NOT NULL,
	scopes     TEXT NOT NULL,
	expires_at INTEGER NOT NULL
)`

// GrantOAuth records that the account allowed the site with this
// redirect URI.
func (s *Store) GrantOAuth(accountID int64, redirect string, now time.Time) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO oauth_grants (account_id, redirect, created_at) VALUES (?, ?, ?)`, accountID, redirect, now.UnixMilli())
	return err
}

// OAuthGrant: a site the account allowed, since when.
type OAuthGrant struct {
	Redirect  string
	CreatedAt time.Time
}

// OAuthGrants are the sites the account allowed, newest first.
func (s *Store) OAuthGrants(accountID int64) ([]OAuthGrant, error) {
	rows, err := s.db.Query(`SELECT redirect, created_at FROM oauth_grants WHERE account_id = ? ORDER BY created_at DESC, redirect`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OAuthGrant{}
	for rows.Next() {
		var g OAuthGrant
		var at int64
		if err := rows.Scan(&g.Redirect, &at); err != nil {
			return nil, err
		}
		g.CreatedAt = time.UnixMilli(at).UTC()
		out = append(out, g)
	}
	return out, rows.Err()
}

// RevokeOAuth takes back the account's grant for the site and ends its
// tokens there (clientID: the site's).
func (s *Store) RevokeOAuth(accountID int64, redirect, clientID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM oauth_grants WHERE account_id = ? AND redirect = ?`, accountID, redirect); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM oauth_tokens WHERE account_id = ? AND client_id = ?`, accountID, clientID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) HasOAuthGrant(accountID int64, redirect string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM oauth_grants WHERE account_id = ? AND redirect = ?`, accountID, redirect).Scan(&n)
	return n > 0, err
}

func (s *Store) CreateOAuthToken(tokenHash string, accountID int64, clientID, scopes string, expires time.Time) error {
	_, err := s.db.Exec(`INSERT INTO oauth_tokens (token_hash, account_id, client_id, scopes, expires_at) VALUES (?, ?, ?, ?, ?)`,
		tokenHash, accountID, clientID, scopes, expires.UnixMilli())
	return err
}

// AccountByOAuthToken is the account of a valid token and its scopes.
func (s *Store) AccountByOAuthToken(tokenHash string, now time.Time) (Account, string, error) {
	var scopes string
	row := s.db.QueryRow(`SELECT `+accountColumns+`, t.scopes FROM oauth_tokens t JOIN accounts a ON a.id = t.account_id
		WHERE t.token_hash = ? AND t.expires_at > ?`, tokenHash, now.UnixMilli())
	acc, err := scanAccountWith(row, &scopes)
	return acc, scopes, err
}

func (s *Store) DeleteExpiredOAuthTokens(now time.Time) error {
	_, err := s.db.Exec(`DELETE FROM oauth_tokens WHERE expires_at <= ?`, now.UnixMilli())
	return err
}
