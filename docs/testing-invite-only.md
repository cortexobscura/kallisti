# Test invite-only registration locally

Use the feature branch and a disposable database. The Compose project below
is separate from the deployment projects: its only published ports are
localhost HTTP 18080 and Mumble TCP/UDP 64739. It never mounts your existing
server database. It needs Docker Compose and available ports 18080/64739.

## Automated checks

From the repository root (Node/pnpm and Go installed):

```sh
pnpm install --frozen-lockfile
pnpm test
pnpm typecheck
(cd server && go test -race ./...)
pnpm --filter @kallisti/web build
```

To run just the new admission and UI regressions:

```sh
pnpm --filter @kallisti/web exec vitest run test/registration.test.ts test/registration-ui.test.ts
(cd server && go test -race ./internal/auth ./internal/mumbled ./internal/store ./internal/sidecar ./cmd/server -run 'Test(Registration|InviteOnly|BuiltinUsesSavedRegistrationPolicy)' -count=1)
```

The Go tests use temporary databases, local sockets and a fake federated
identity provider. They cover open defaults, upgrade/restart behavior,
single-use races, rollback, expiry, revocation, browser challenges, admin
permissions, password/federated/QR login and native Mumble admission.

## Start a disposable server

In Bash, from the repository root:

```bash
read -rsp 'Choose a disposable SuperUser password: ' MUMBLED_SUPERUSER_PASSWORD
printf '\n'
export MUMBLED_SUPERUSER_PASSWORD
docker compose -f deploy/compose.invite-test.yml up -d --build
docker compose -f deploy/compose.invite-test.yml logs --tail=30 server
```

Keep this terminal for restart and cleanup commands. The password is for
this test server only; it is not written into a tracked file or command-line
argument. Leave the variable set until cleanup because Compose validates it
on each invocation.

Open **http://localhost:18080** on the same computer. Use `localhost`
consistently; invitation links are specific to their server origin. In the
**Log in** tab, use `SuperUser` and the password you just chose. The first
build may take several minutes. This setup disables federation and external
previews/push; federation is covered by the automated tests above. Testing
with a real federated provider requires a reachable HTTPS deployment.

## Manual acceptance checks

1. **Open default:** Before enabling the setting, use a separate browser
   profile to create `before-invite-mode`. Verify normal login and chat work.
   Keep that session open. A guest should also be able to join initially.
2. **Close registration:** As SuperUser, open **Server settings →
   Registration**, enable **Invite-only server**, and wait for confirmation.
   The existing user's session should still work. Log that user out and back
   in; the existing account should still be accepted.
3. **Uninvited access:** In a fresh browser profile, visit the bare site URL.
   The Guest option should be absent and signup should require an invitation.
   The HTTP checks below also prove enforcement without relying on the UI.
4. **Invited signup:** As SuperUser, create an invitation with **1 use** and
   **7 days**. Copy the full link, including its fragment after `#`. Open it
   in a logged-out browser profile and create a new account. It should join
   normally. In another fresh profile, try the same link for a different
   account; signup must fail. Refresh the admin invitation list to see one use.
5. **Revocation:** Create another invitation, revoke it, then try creating an
   account with its link. Signup must fail. Revoking a used link must not
   remove the account that already used it.
6. **Persistence:** Restart the test server with the command below. Existing
   accounts must still log in, invite-only must remain enabled, and a used
   or revoked link must remain unusable. An unused valid link must still work.
7. **Restore open registration:** Disable invite-only as SuperUser. In a new
   browser profile, signup without an invitation and guest admission should
   work again.
8. **Channel separation:** A private-channel invitation must not satisfy the
   account-invitation field. A newly invited account must not gain access to
   private channels until separately authorized.

With invite-only enabled, both commands must return HTTP **403** and
`"code":"invite_required"`:

```sh
curl -i http://localhost:18080/api/accounts \
  -H 'Origin: http://localhost:18080' -H 'Content-Type: application/json' \
  --data '{"name":"uninvited-probe","password":"disposable-test-only"}'
curl -i http://localhost:18080/api/guest \
  -H 'Origin: http://localhost:18080' -H 'Content-Type: application/json' \
  --data '{"name":"uninvited-guest"}'
```

Native Mumble: connect to `localhost:64739`. A new unregistered user should
be rejected while invite-only is enabled. An existing Kallisti username and
password, or a certificate registered before the switch, must still work.
The disposable server uses a self-signed Mumble certificate.

Already connected anonymous guests are not evicted, but cannot register
after the switch. Existing temporary accounts retain their normal lifetime.
This feature does not change file URL authorization.

## Restart and cleanup

Restart while keeping test data:

```sh
docker compose -f deploy/compose.invite-test.yml restart server
```

Stop and **delete only this disposable project's database and uploads**:

```sh
docker compose -f deploy/compose.invite-test.yml down --volumes
unset MUMBLED_SUPERUSER_PASSWORD
```

For a clean repeat, start again and choose a test password. To retain test
accounts, omit `--volumes` when stopping.
