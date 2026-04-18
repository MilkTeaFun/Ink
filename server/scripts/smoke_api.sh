#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
SERVER_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
REPO_DIR=$(CDPATH= cd -- "$SERVER_DIR/.." && pwd)

SMOKE_PORT="${SMOKE_PORT:-18080}"
SMOKE_BASE_URL="${SMOKE_BASE_URL:-http://127.0.0.1:${SMOKE_PORT}}"
SMOKE_TIMEOUT_SECONDS="${SMOKE_TIMEOUT_SECONDS:-30}"
BOOTSTRAP_DB="${INK_SMOKE_BOOTSTRAP_DB:-1}"

cleanup() {
  if [ "${SERVER_PID:-}" != "" ] && kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
  fi
  rm -f "${SERVER_LOG:-}" "${WORKSPACE_BEFORE:-}" "${WORKSPACE_AFTER:-}" "${WORKSPACE_RESTORED:-}" "${LOGIN_JSON:-}" "${UPDATED_PAYLOAD:-}"
}

trap cleanup EXIT INT TERM

"$SERVER_DIR/scripts/ensure_dev_env.sh"

if [ "$BOOTSTRAP_DB" = "1" ]; then
  (
    cd "$REPO_DIR"
    make dev-db
  )
fi

(
  cd "$SERVER_DIR"
  go run ./cmd/migrate up
  go run ./cmd/seed dev
)

CREDENTIALS_FILE="$SERVER_DIR/.dev-admin-password"
if [ ! -r "$CREDENTIALS_FILE" ]; then
  echo "Missing credentials file: $CREDENTIALS_FILE" >&2
  exit 1
fi

LOGIN_NAME=$(awk -F= '$1=="login" {print $2}' "$CREDENTIALS_FILE")
LOGIN_PASSWORD=$(awk -F= '$1=="password" {print $2}' "$CREDENTIALS_FILE")

if [ -z "$LOGIN_NAME" ] || [ -z "$LOGIN_PASSWORD" ]; then
  echo "Invalid credentials file: $CREDENTIALS_FILE" >&2
  exit 1
fi

SERVER_LOG=$(mktemp "${TMPDIR:-/tmp}/ink-smoke-server.XXXXXX.log")
WORKSPACE_BEFORE=$(mktemp "${TMPDIR:-/tmp}/ink-smoke-workspace-before.XXXXXX.json")
WORKSPACE_AFTER=$(mktemp "${TMPDIR:-/tmp}/ink-smoke-workspace-after.XXXXXX.json")
WORKSPACE_RESTORED=$(mktemp "${TMPDIR:-/tmp}/ink-smoke-workspace-restored.XXXXXX.json")
LOGIN_JSON=$(mktemp "${TMPDIR:-/tmp}/ink-smoke-login.XXXXXX.json")
UPDATED_PAYLOAD=$(mktemp "${TMPDIR:-/tmp}/ink-smoke-workspace-updated.XXXXXX.json")

(
  cd "$SERVER_DIR"
  PORT="$SMOKE_PORT" \
  SCHEDULER_POLL_INTERVAL=24h \
  go run ./cmd/api >"$SERVER_LOG" 2>&1
) &
SERVER_PID=$!

attempt=0
until curl --silent --show-error --fail "$SMOKE_BASE_URL/healthz" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge "$SMOKE_TIMEOUT_SECONDS" ]; then
    echo "API failed to start in time. Recent log output:" >&2
    tail -n 40 "$SERVER_LOG" >&2 || true
    exit 1
  fi
  sleep 1
done

curl --silent --show-error --fail \
  -H "Content-Type: application/json" \
  -X POST \
  -d "{\"email\":\"$LOGIN_NAME\",\"password\":\"$LOGIN_PASSWORD\"}" \
  "$SMOKE_BASE_URL/api/v1/auth/login" >"$LOGIN_JSON"

ACCESS_TOKEN=$(
  python3 - "$LOGIN_JSON" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    payload = json.load(fh)

token = payload.get("accessToken", "")
if not token:
    raise SystemExit("missing access token")

print(token)
PY
)

AUTH_HEADER="Authorization: Bearer $ACCESS_TOKEN"

curl --silent --show-error --fail \
  -H "$AUTH_HEADER" \
  "$SMOKE_BASE_URL/api/v1/workspace" >"$WORKSPACE_BEFORE"

python3 - "$WORKSPACE_BEFORE" "$UPDATED_PAYLOAD" <<'PY'
import json
import sys

source_path, target_path = sys.argv[1], sys.argv[2]
with open(source_path, "r", encoding="utf-8") as fh:
    workspace = json.load(fh)

preferences = workspace.setdefault("preferences", {})
preferences["sendConfirmationEnabled"] = not bool(preferences.get("sendConfirmationEnabled", False))

with open(target_path, "w", encoding="utf-8") as fh:
    json.dump(workspace, fh, ensure_ascii=False)
PY

curl --silent --show-error --fail \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -X PUT \
  --data-binary "@$UPDATED_PAYLOAD" \
  "$SMOKE_BASE_URL/api/v1/workspace" >"$WORKSPACE_AFTER"

python3 - "$WORKSPACE_BEFORE" "$WORKSPACE_AFTER" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    before = json.load(fh)
with open(sys.argv[2], "r", encoding="utf-8") as fh:
    after = json.load(fh)

before_flag = bool(before.get("preferences", {}).get("sendConfirmationEnabled", False))
after_flag = bool(after.get("preferences", {}).get("sendConfirmationEnabled", False))

if before_flag == after_flag:
    raise SystemExit("workspace update did not persist the expected change")
PY

curl --silent --show-error --fail \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -X PUT \
  --data-binary "@$WORKSPACE_BEFORE" \
  "$SMOKE_BASE_URL/api/v1/workspace" >"$WORKSPACE_RESTORED"

python3 - "$WORKSPACE_BEFORE" "$WORKSPACE_RESTORED" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    expected = json.load(fh)
with open(sys.argv[2], "r", encoding="utf-8") as fh:
    restored = json.load(fh)

if expected != restored:
    raise SystemExit("workspace state was not restored after smoke test")
PY

echo "Smoke test passed: login, workspace load, workspace save, and restore succeeded."
