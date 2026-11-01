# Invite-only registration

## Contract

An optional, persisted admin setting closes public account creation and new
guest admission. Default is open, including upgrades. Existing accounts,
sessions, registered Mumble users, QR logins and federated logins keep their
existing authentication and authorization checks. Existing temporary accounts
may still become permanent. No automatic eviction or account deletion.

Only admins of the built-in server manage this policy and account invitation
links. Moderators, channel moderators and ordinary members cannot do so.
Account links are distinct from private-channel invitations: neither confers
the other's privilege. External Murmur admission remains owned by Murmur.

## Implementation boundaries

- `store/registration.go`: additive tables, default-open policy, hashed
  invitations, expiry/revocation/use limits. Account insertion, federated
  identity linking and invitation consumption share one transaction. Policy
  is checked again inside the transaction after expensive authentication work.
- `auth`: early refusal before password hashing/certificate generation;
  invitation accepted only by account-creation endpoints. Existing login,
  session, QR and identity-linking paths are unchanged.
- `mumbled`: unregistered admission and subsequent guest registration check
  the current policy. Authenticated registrations and web accounts reach their
  existing paths first. Database failures do not open guest admission.
- `sidecar`: admin-authorized policy and invitation operations, separate
  message types; new clients hide controls when capability is absent.
- Web client: Registration settings, account invite links at
  `/register#<token>`, invitation preserved across the federated login round
  trip, and an invitation field for first-time federation and desktop signup.
  The fragment keeps tokens out of page requests and Referer headers.

## Compatibility and proof

Before changes, `pnpm test` passed (143 protocol tests, 1285 web tests, all Go
packages). The targeted Go baseline also passed for store, auth, sidecar,
mumbled and cmd/server.

Regression coverage must demonstrate:

- Open default and upgrades retain accounts and behavior; policy persists
  through restart and can be disabled again.
- Closed signup, temporary guests and native anonymous joins are refused.
- Existing password/federated accounts, sessions, native certificate
  registrations and SuperUser retain access.
- Invitations expire, can be revoked, and never bypass existing name,
  password, browser-challenge or identity checks.
- Competing requests cannot exceed a use limit; failed account insertion or
  identity linking does not consume an invitation or leave a partial account.
- Channel invitations cannot create accounts, and registration invitations
  do not grant private-channel membership.
- Only admins can list/create/revoke links or change policy; raw tokens are
  returned only to their creator, not stored or broadcast.
- New clients work with older server messages; older clients remain usable
  on open servers. Closed servers enforce policy even for an older client.
- The full TypeScript/Go suite, type checks, Go vet and race tests pass.

## Limits

This is admission control, not a redesign of file authorization or an
eviction mechanism. Already connected anonymous guests remain until they
leave; existing temporary accounts retain their normal lifetime. Server
owners wanting everyone removed must use existing moderation controls.

No existing tables or protocol messages are removed. Older binaries ignore
the new policy, so downgrading reopens admission. Multi-server newcomers can
paste an invitation into the destination server's signup screen; account
invitations do not introduce a new cross-server routing protocol.

Local automated tests do not establish production deployment, browser voice
behavior, mobile packaging or live CI results.

## Local validation (2026-11-01)

- `pnpm test`: passed; 143 protocol tests, 1299 web tests and all Go
  packages. This includes mounted Svelte component tests for invitation
  signup, federated signup and the admin controls.
- `pnpm typecheck`: passed, including Go vet; Svelte reports no errors or
  warnings.
- `cd server && go test -race ./...`: passed for all packages.
- Final targeted race tests passed after adding explicit QR-login and
  native-password-login checks in invite-only mode.
- Reconciled with main at `825fc5c126b82054258463a444128215c1c155c2`;
  reran the full test suite, type checks, build, and race tests for auth,
  store, sidecar, mumbled and cmd/server. Invited federated signup also
  explicitly verifies that the upstream handoff consent is still required.
- `pnpm --filter @kallisti/web build`: passed. Vite reports its bundle-size
  advisory; no build failure.
- `git diff --check`: passed.

See [the local test guide](../../docs/testing-invite-only.md) for a disposable
Docker setup and manual acceptance checks. Compose configuration was
validated; the Docker deployment and manual browser checks have not been run.
No live server configuration has been changed.
