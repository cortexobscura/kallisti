package store

import (
	"database/sql"
	"errors"
	"time"
)

// Kallisti as a login provider (see auth's provider.go): other servers
// register as clients, users allow a site once, and its tokens (by hash,
// short-lived) read who they are.
const oauthTables = `CREATE TABLE IF NOT EXISTS oauth_clients (
	client_id    TEXT PRIMARY KEY,
	secret_hash  TEXT NOT NULL,
	name         TEXT NOT NULL,
	redirect_uri TEXT NOT NULL,
	scopes       TEXT NOT NULL,
	website      TEXT NOT NULL,
	created_at   INTEGER NOT NULL,
	used         INTEGER NOT NULL DEFAULT 0 -- the last login (0: none)
);
CREATE TABLE IF NOT EXISTS oauth_grants (
	account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	redirect   TEXT NOT NULL, -- the site's redirect URI
	created_at INTEGER NOT NULL,
	PRIMARY KEY (account_id, redirect)
);
CREATE TABLE IF NOT EXISTS oauth_tokens (
	token_hash TEXT PRIMARY KEY,
	account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	client_id  TEXT NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
	scopes     TEXT NOT NULL,
	expires_at INTEGER NOT NULL
)`

// OAuthClient is another server's app here (its secret by hash).
type OAuthClient struct {
	ID, SecretHash, Name, RedirectURI, Scopes, Website string
	// Used: a login went through.
	Used bool
}

func (s *Store) CreateOAuthClient(c OAuthClient, now time.Time) error {
	_, err := s.db.Exec(`INSERT INTO oauth_clients (client_id, secret_hash, name, redirect_uri, scopes, website, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.SecretHash, c.Name, c.RedirectURI, c.Scopes, c.Website, now.UnixMilli())
	return err
}

func (s *Store) OAuthClient(id string) (OAuthClient, error) {
	c := OAuthClient{ID: id}
	var used int64
	err := s.db.QueryRow(`SELECT secret_hash, name, redirect_uri, scopes, website, used FROM oauth_clients WHERE client_id = ?`, id).
		Scan(&c.SecretHash, &c.Name, &c.RedirectURI, &c.Scopes, &c.Website, &used)
	c.Used = used > 0
	if errors.Is(err, sql.ErrNoRows) {
		return OAuthClient{}, ErrNotFound
	}
	return c, err
}

func (s *Store) MarkOAuthClientUsed(id string, now time.Time) error {
	_, err := s.db.Exec(`UPDATE oauth_clients SET used = ? WHERE client_id = ?`, now.UnixMilli(), id)
	return err
}

// DeleteUnusedOAuthClients removes clients (anyone may register one) made
// before unused that no login used, and those without a login since idle.
func (s *Store) DeleteUnusedOAuthClients(unused, idle time.Time) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM oauth_clients WHERE (used = 0 AND created_at < ?) OR (used > 0 AND used < ?)`, unused.UnixMilli(), idle.UnixMilli())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteLeastUsedOAuthClient makes room: the client whose last login (or
// making, without one) is oldest goes.
func (s *Store) DeleteLeastUsedOAuthClient() error {
	_, err := s.db.Exec(`DELETE FROM oauth_clients WHERE client_id = (SELECT client_id FROM oauth_clients ORDER BY MAX(used, created_at) LIMIT 1)`)
	return err
}

func (s *Store) CountOAuthClients() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM oauth_clients`).Scan(&n)
	return n, err
}

// GrantOAuth records that the account allowed the site with this
// redirect URI (its apps may change, the address stays).
func (s *Store) GrantOAuth(accountID int64, redirect string, now time.Time) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO oauth_grants (account_id, redirect, created_at) VALUES (?, ?, ?)`, accountID, redirect, now.UnixMilli())
	return err
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
