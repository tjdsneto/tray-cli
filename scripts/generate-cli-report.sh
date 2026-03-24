#!/usr/bin/env bash
# Captures CLI output for mostly non-mutating commands (--help / read-only). Default: tmp/cli-report.md
set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/tmp"
mkdir -p "${OUT_DIR}"
REPORT_PATH="${1:-${OUT_DIR}/cli-report.md}"

cd "$ROOT_DIR" || exit 1

timestamp="$(date -u +"%Y-%m-%d %H:%M:%S UTC")"

commands=(
  "./run.sh --help"
  "./run.sh login --help"
  "./run.sh status"
  "./run.sh create --help"
  "./run.sh ls"
  "./run.sh rename --help"
  "./run.sh delete-tray --help"
  "./run.sh invite --help"
  "./run.sh rotate-invite --help"
  "./run.sh join --help"
  "./run.sh add --help"
  "./run.sh list"
  "./run.sh remove --help"
  "./run.sh contributed"
  "./run.sh remote --help"
  "./run.sh remote ls"
  "./run.sh members --help"
  "./run.sh revoke --help"
  "./run.sh leave --help"
  "./run.sh review"
  "./run.sh accept --help"
  "./run.sh decline --help"
  "./run.sh snooze --help"
  "./run.sh complete --help"
  "./run.sh archive --help"
  "./run.sh listen --once"
)

{
  echo "# CLI Command Report"
  echo
  echo "- Generated: $timestamp"
  echo "- Repo: \`$ROOT_DIR\`"
  echo "- Notes: Non-mutating commands are executed directly; mutating commands are exercised with \`--help\` to avoid side effects."
  echo
} > "$REPORT_PATH"

for cmd in "${commands[@]}"; do
  output="$(
    set +e
    eval "$cmd" 2>&1
    echo "__EXIT_CODE__$?"
  )"

  exit_code="$(printf '%s\n' "$output" | awk -F'__EXIT_CODE__' 'NF>1{print $2}' | tail -n1)"
  body="$(printf '%s\n' "$output" | sed '/__EXIT_CODE__/d')"

  {
    echo "## \`$cmd\`"
    echo
    echo "- Exit code: \`$exit_code\`"
    echo
    echo '```text'
    if [ -n "$body" ]; then
      printf '%s\n' "$body"
    fi
    echo '```'
    echo
  } >> "$REPORT_PATH"
done

echo "Report written to: $REPORT_PATH"
