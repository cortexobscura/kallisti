# Join other servers with your account

## Summary

As close to Discord as possible: someone sends an invite to a Kallisti
server, you sign in there with your existing account from another
Kallisti server, and from then on every Kallisti site and app you open
shows all your servers in the rail. Two things work across instances
(that allow it): your account and your server list.

Guiding rule of the project: it is Mumble, it is easy to use, and when the
two conflict, easy to use wins.

## Decisions (2026-10-27)

- **The home server holds your account and your server list** (federated,
  as Mastodon or Matrix). No central coordinator, no DHT for now. A hybrid
  later (an identity key in DNS or a DHT, so an account can move to another
  home) stays possible; a central directory is not wanted, at most later
  and optional, for finding servers only.
- Accepted costs: signing in to a new server needs the home server online
  (existing sessions keep working); the home server could act as you
  elsewhere (as in every federated system); moving homes means a new
  identity, for now.
- **Browser and apps together.**
- **Switching servers connects in place:** one connection per server from
  the page or app, one voice channel across all of them, voice goes on
  while you switch.

## Why not a coordinator or a DHT

Each server is its own site: a page on one cannot read another's cookies
or storage, and third-party cookies are blocked. Someone must hold the
account and the list for every site to ask. A coordinator is one party for
everyone (a single point of failure and trust, against self-hosting). A DHT
needs a private key per user: in the browser it is lost with the storage
and must be copied to each device (passkeys are bound to one domain), and
browsers reach a DHT only through HTTP relays. Both lead back to a server
holding it: the home server. iroh stays useful for servers without a
domain or public address, not for accounts.

## How it works

- **Identity:** `lain@chat.kallisti.chat`, found by webfinger. Every
  Kallisti server is a login provider with the Mastodon OAuth subset that
  our fediverse login already uses from the other side (instance, apps,
  authorize, token, verify_credentials). Another server's fediverse login
  then works for Kallisti accounts almost unchanged and makes a local
  account linked to the remote one ("name@home" only on a name conflict).
- **Server list:** on the home account. A server you joined reads and
  updates it server to server with the token from the login (scope: your
  identity and your list). Joining or leaving anywhere changes it at home.
- **Tokens:** the client keeps one session token per server (per site in
  the browser, per app). The first click on a server in a new browser opens
  a short popup through the home server, which approves at once after the
  first time; the token comes back by postMessage (deep link in the apps),
  never in the URL.
- **Invite links:** "Continue with your Kallisti account" on the invite
  page; the home address is typed once per browser and remembered by that
  site. The home client registers as the handler of `web+kallisti:` links
  (Chrome, Firefox), so invite pages offer "Open in my Kallisti" in one
  click; invite links clicked inside a Kallisti app are handled directly.
- **Servers allow:** token requests and WebSockets from any origin (safe
  with tokens: no cookies, no cross-site forgery); pages may connect to
  other servers (looser CSP).
- **Other home servers are accepted by default.** Admins can set a block
  list, or an allow list (then only those), in the server settings.
- **Mumble clients:** nothing changes; remote users are registered users.

## Steps

1. [Several servers in one client](multiple-servers.md): a token and a
   connection per server, the rail, one voice channel; adding a server with
   its normal login (password, guest, fediverse).
2. [Cross-origin access with tokens](cross-origin-tokens.md): servers
   accept token requests and WebSockets from any origin; the page's CSP
   allows other servers.
3. [Kallisti as a login provider](kallisti-login-provider.md): the
   Mastodon OAuth subset and webfinger, a consent page, approval at once
   after the first time; "Continue with your Kallisti account" on the
   login and invite pages; the block and allow lists.
4. [The server list on the home account](home-server-list.md): stored at
   home, read and updated server to server, the rail on every site.
5. [Tokens without typing](seamless-tokens.md): the popup per new browser,
   the `web+kallisti:` handler, invite links inside the apps.

## Not in scope

- Federation beyond joining (private messages or friend lists across
  servers).
- Moving an account to another home (see the hybrid above).
