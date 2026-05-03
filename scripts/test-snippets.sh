#!/usr/bin/env bash
# test-snippets.sh — transpile every testdata/snippets/*.md and diff the output
# against the corresponding *.tex golden file.
#
# Per-snippet options are read from an optional <name>.opts sidecar file.
# Each non-empty line is one token:
#   standalone   — pass -standalone to the binary
#
# Usage:
#   ./scripts/test-snippets.sh            # compare against golden files
#   ./scripts/test-snippets.sh --update   # regenerate golden files

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SNIPPET_DIR="$REPO_ROOT/testdata/snippets"
BIN="$REPO_ROOT/bin/marktex"
UPDATE=false

for arg in "$@"; do
  case "$arg" in
    --update|-u) UPDATE=true ;;
    --help|-h)
      echo "Usage: $0 [--update]"
      echo "  --update  Overwrite .tex golden files with current transpiler output."
      exit 0
      ;;
    *) echo "Unknown flag: $arg" >&2; exit 2 ;;
  esac
done

echo "Building marktex..."
(cd "$REPO_ROOT" && go build -o "$BIN" ./cmd/marktex)

pass=0
fail=0

for md_file in "$SNIPPET_DIR"/*.md; do
  base="${md_file%.md}"
  tex_file="${base}.tex"
  opts_file="${base}.opts"
  name="$(basename "$base")"

  # Read per-snippet options
  extra_flags=""
  if [[ -f "$opts_file" ]]; then
    while IFS= read -r line; do
      token="$(echo "$line" | tr '[:upper:]' '[:lower:]' | xargs)"
      case "$token" in
        standalone) extra_flags="$extra_flags -standalone" ;;
      esac
    done < "$opts_file"
  fi

  # shellcheck disable=SC2086
  actual="$("$BIN" $extra_flags "$md_file")"

  if $UPDATE; then
    printf '%s\n' "$actual" > "$tex_file"
    echo "  UPDATED  $name"
    continue
  fi

  if [[ ! -f "$tex_file" ]]; then
    echo "  MISSING  $name  (no golden file — run with --update to create it)"
    (( fail++ )) || true
    continue
  fi

  expected="$(cat "$tex_file")"

  if [[ "$actual" == "$expected" ]]; then
    echo "  PASS     $name"
    (( pass++ )) || true
  else
    echo "  FAIL     $name"
    diff <(echo "$expected") <(echo "$actual") | head -30 | sed 's/^/           /'
    (( fail++ )) || true
  fi
done

echo ""
if $UPDATE; then
  echo "Golden files updated."
  exit 0
fi

echo "Results: $pass passed, $fail failed."
[[ $fail -eq 0 ]]
