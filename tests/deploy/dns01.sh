#!/bin/sh
# Offline checks for the standalone Cloudflare DNS-01 deployment variant.
set -eu
cd "$(dirname "$0")/../.."

export DOMAIN=chat.example.org
export CLOUDFLARE_API_TOKEN_FILE=/dev/null
compose='deploy/compose-dns01.yaml'

# Check the effective topology without starting services or loading a real token.
docker compose -f "$compose" config --format json | python3 -c '
import json
import sys

config = json.load(sys.stdin)
caddy = config["services"]["caddy"]
app = config["services"]["kallisti"]
copier = config["services"]["cert-sync"]
assert config["name"] == "deploy"
assert {port["published"] for port in caddy["ports"]} == {"443"}
assert {port["published"] for port in app["ports"]} == {"64738", "8189"}
assert caddy["read_only"] and caddy["secrets"]
assert config["secrets"]["cloudflare_api_token"]["file"] == "/dev/null"
assert "cloudflare_api_token" in caddy["command"][0]
assert config["volumes"]["caddy-data"]["name"] == "deploy_caddy-data"
assert config["volumes"]["kallisti-data"]["name"] == "deploy_kallisti-data"
assert any(mount["source"] == "caddy-data" for mount in caddy["volumes"])
assert any(mount["source"] == "caddy-data" and mount["read_only"] for mount in copier["volumes"])
assert any(mount["source"] == "kallisti-data" for mount in app["volumes"])
'

docker compose -f "$compose" build --quiet caddy
docker run --rm deploy-caddy caddy list-modules | grep -Fxq dns.providers.cloudflare
docker run --rm --read-only -e DOMAIN="$DOMAIN" \
  -e CF_API_TOKEN=0123456789abcdef0123456789abcdef01234567 \
  -v "$PWD/deploy/Caddyfile.dns01:/etc/caddy/Caddyfile:ro" \
  deploy-caddy caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
