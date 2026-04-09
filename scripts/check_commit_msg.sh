#!/bin/sh

set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <commit-msg-file>" >&2
  exit 1
fi

msg_file="$1"

if [ ! -f "$msg_file" ]; then
  echo "commit message file not found: $msg_file" >&2
  exit 1
fi

first_line=$(sed -n '1{s/\r$//;p;}' "$msg_file")

if [ -z "$first_line" ]; then
  echo "invalid commit message: first line is empty" >&2
  exit 1
fi

case "$first_line" in
  Merge\ *|Revert\ *)
    exit 0
    ;;
esac

pattern='^(feat|fix|docs|refactor|chore|perf)\([a-z0-9][a-z0-9-]*\): [a-z][^.]*[^.]$'

if printf '%s\n' "$first_line" | grep -Eq "$pattern"; then
  exit 0
fi

cat >&2 <<'EOF'
invalid commit message format

expected:
  type(scope): summary

rules:
  - allowed types: feat, fix, docs, refactor, chore, perf
  - scope is required and must use lowercase letters, digits, or dashes
  - summary must start with a lowercase letter
  - summary must not end with a period

example:
  refactor(printer): use memobird textrender for text prints
EOF

exit 1
