# Tasks

## Ranked queue

No active tasks.

## completed

- **Add a standalone Docker Compose DNS-01 variant for Cloudflare.** Implemented in `deploy/compose-dns01.yaml` with the custom Caddy build, scoped token-file mount, certificate sync, deployment guidance, and offline checks. Verified with `sh tests/deploy/dns01.sh`, `sh tests/deploy/cert-sync.sh`, ShellCheck, and Compose configuration checks; public certificate issuance requires an operator test. Committed as "feat(deploy): add cloudflare dns-01 compose variant".
