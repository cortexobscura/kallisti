#!/bin/sh
# Offline checks: the Compose files run Kallisti as PUID:PGID (default 1000:1000).
set -eu
cd "$(dirname "$0")/../.."

unset PUID PGID
export DOMAIN=chat.example.org PROXY_NETWORK=proxy MUMBLE_HOST=murmur.example.org
user_of() { docker compose -f "$1" config --format json | python3 -c 'import json,sys; print(json.load(sys.stdin)["services"]["kallisti"]["user"])'; }

for file in deploy/compose.yml deploy/compose.behind-proxy.yml deploy/compose.murmur.yml; do
  [ "$(user_of "$file")" = "1000:1000" ] || { echo "$file: default user is not 1000:1000" >&2; exit 1; }
  [ "$(PUID=1234 PGID=5678 user_of "$file")" = "1234:5678" ] || { echo "$file: PUID/PGID not used" >&2; exit 1; }
done
grep -q 'adduser -D -H -u 1000 -G kallisti kallisti' Dockerfile || { echo "Dockerfile: the user's id is not fixed at 1000" >&2; exit 1; }
echo ok
