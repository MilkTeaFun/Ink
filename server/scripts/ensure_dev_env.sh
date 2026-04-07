#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
SERVER_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
TEMPLATE_FILE="$SERVER_DIR/.env.example"
ENV_FILE="$SERVER_DIR/.env"

if [ -f "$ENV_FILE" ]; then
  exit 0
fi

if [ ! -r "$TEMPLATE_FILE" ]; then
  echo "Missing template: $TEMPLATE_FILE" >&2
  exit 1
fi

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required to generate JWT_SECRET for $ENV_FILE" >&2
  exit 1
fi

JWT_SECRET=$(openssl rand -hex 32)
umask 077
TMP_ENV_FILE=$(mktemp "${ENV_FILE}.tmp.XXXXXX")
trap 'rm -f "$TMP_ENV_FILE"' EXIT INT TERM
sed "s/^JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$TEMPLATE_FILE" >"$TMP_ENV_FILE"
chmod 600 "$TMP_ENV_FILE"
mv "$TMP_ENV_FILE" "$ENV_FILE"
trap - EXIT INT TERM

echo "Created $ENV_FILE with local development defaults."
