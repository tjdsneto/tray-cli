#!/usr/bin/env bash
# End-to-end CLI exercise (creates trays/items; cleans up). Default: tmp/cli-e2e-report.md
set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/tmp"
mkdir -p "${OUT_DIR}"
REPORT_PATH="${1:-${OUT_DIR}/cli-e2e-report.md}"
cd "$ROOT_DIR" || exit 1

timestamp="$(date -u +"%Y-%m-%d %H:%M:%S UTC")"
run_id="$(date -u +"%Y%m%d%H%M%S")"

created_tray_ids=()
LAST_OUTPUT=""
LAST_EXIT=0

append_report() {
  local cmd="$1"
  local exit_code="$2"
  local output="$3"
  {
    echo "## \`$cmd\`"
    echo
    echo "- Exit code: \`$exit_code\`"
    echo
    echo '```text'
    if [ -n "$output" ]; then
      printf '%s\n' "$output"
    fi
    echo '```'
    echo
  } >> "$REPORT_PATH"
}

run_cmd() {
  local cmd="$1"
  local output
  output="$(
    set +e
    eval "$cmd" 2>&1
    echo "__EXIT_CODE__$?"
  )"
  LAST_EXIT="$(printf '%s\n' "$output" | awk -F'__EXIT_CODE__' 'NF>1{print $2}' | tail -n1)"
  LAST_OUTPUT="$(printf '%s\n' "$output" | sed '/__EXIT_CODE__/d')"
  append_report "$cmd" "$LAST_EXIT" "$LAST_OUTPUT"
}

first_id_from_json() {
  python3 - <<'PY' "$1"
import json, sys
s = sys.argv[1]
try:
    obj = json.loads(s)
except Exception:
    print("")
    raise SystemExit(0)

if isinstance(obj, list) and obj:
    v = obj[0].get("id", "")
    print(v if isinstance(v, str) else "")
elif isinstance(obj, dict):
    v = obj.get("id", "")
    print(v if isinstance(v, str) else "")
else:
    print("")
PY
}

token_from_json() {
  python3 - <<'PY' "$1"
import json, sys
s = sys.argv[1]
try:
    obj = json.loads(s)
except Exception:
    print("")
    raise SystemExit(0)

if isinstance(obj, list) and obj:
    v = obj[0].get("invite_token", "")
    print(v if isinstance(v, str) else "")
elif isinstance(obj, dict):
    v = obj.get("invite_token", "")
    print(v if isinstance(v, str) else "")
else:
    print("")
PY
}

cleanup() {
  local tid
  for tid in "${created_tray_ids[@]}"; do
    if [ -n "$tid" ]; then
      run_cmd "./run.sh delete-tray \"$tid\""
    fi
  done
}
trap cleanup EXIT

{
  echo "# CLI E2E Report"
  echo
  echo "- Generated: $timestamp"
  echo "- Run ID: \`$run_id\`"
  echo "- Repo: \`$ROOT_DIR\`"
  echo "- Scope: end-to-end command exercise with real create/update/delete flows."
  echo
} > "$REPORT_PATH"

# 1) Baseline
run_cmd "./run.sh status"
run_cmd "./run.sh ls --format json"

# 2) Create trays
tray_a="e2e-${run_id}-alpha"
tray_b="e2e-${run_id}-beta"

run_cmd "./run.sh create \"$tray_a\" --format json"
tray_a_id="$(first_id_from_json "$LAST_OUTPUT")"
[ -n "$tray_a_id" ] && created_tray_ids+=("$tray_a_id")

run_cmd "./run.sh create \"$tray_b\" --format json"
tray_b_id="$(first_id_from_json "$LAST_OUTPUT")"
[ -n "$tray_b_id" ] && created_tray_ids+=("$tray_b_id")

run_cmd "./run.sh ls --format json"

# 3) Rename
tray_a_renamed="${tray_a}-renamed"
if [ -n "$tray_a_id" ]; then
  run_cmd "./run.sh rename \"$tray_a_id\" \"$tray_a_renamed\""
fi

# 4) Invite + rotate invite + join (self-token flow)
if [ -n "$tray_a_id" ]; then
  run_cmd "./run.sh invite \"$tray_a_id\" --format json"
  invite_token="$(token_from_json "$LAST_OUTPUT")"
  run_cmd "./run.sh rotate-invite \"$tray_a_id\" --format json"
  if [ -n "$invite_token" ]; then
    run_cmd "./run.sh join \"$invite_token\""
  fi
fi

# 5) Add several items for triage and lifecycle
create_item_and_capture_id() {
  local title="$1"
  local tray_ref="$2"
  run_cmd "./run.sh add \"$title\" \"$tray_ref\" --format json"
  first_id_from_json "$LAST_OUTPUT"
}

item_accept_id=""
item_decline_id=""
item_snooze_id=""
item_complete_id=""
item_archive_id=""
item_remove_id=""

if [ -n "$tray_a_id" ]; then
  item_accept_id="$(create_item_and_capture_id "E2E accept ${run_id}" "$tray_a_id")"
  item_decline_id="$(create_item_and_capture_id "E2E decline ${run_id}" "$tray_a_id")"
  item_snooze_id="$(create_item_and_capture_id "E2E snooze ${run_id}" "$tray_a_id")"
  item_complete_id="$(create_item_and_capture_id "E2E complete ${run_id}" "$tray_a_id")"
  item_archive_id="$(create_item_and_capture_id "E2E archive ${run_id}" "$tray_a_id")"
  item_remove_id="$(create_item_and_capture_id "E2E remove ${run_id}" "$tray_a_id")"
fi

run_cmd "./run.sh list --format json"
if [ -n "$tray_a_id" ]; then
  run_cmd "./run.sh list \"$tray_a_id\" --format json"
fi
run_cmd "./run.sh review --format json"
if [ -n "$tray_a_id" ]; then
  run_cmd "./run.sh review \"$tray_a_id\" --format json"
fi

# 6) Triage transitions
if [ -n "$item_accept_id" ]; then
  run_cmd "./run.sh accept \"$item_accept_id\""
fi
if [ -n "$item_decline_id" ]; then
  run_cmd "./run.sh decline \"$item_decline_id\" --reason \"e2e test\""
fi
if [ -n "$item_snooze_id" ]; then
  snooze_until="$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d '+1 hour' +"%Y-%m-%dT%H:%M:%SZ")"
  run_cmd "./run.sh snooze \"$item_snooze_id\" --until \"$snooze_until\""
fi
if [ -n "$item_complete_id" ]; then
  run_cmd "./run.sh complete \"$item_complete_id\" --message \"done in e2e\""
fi
if [ -n "$item_archive_id" ]; then
  run_cmd "./run.sh archive \"$item_archive_id\""
fi

# 7) Remove + contributed + members/revoke/leave + remote + listen once
if [ -n "$item_remove_id" ]; then
  run_cmd "./run.sh remove \"$item_remove_id\""
fi
run_cmd "./run.sh contributed --format json"
if [ -n "$tray_a_id" ]; then
  run_cmd "./run.sh members \"$tray_a_id\" --format json"
  run_cmd "./run.sh revoke \"$tray_a_id\" \"00000000-0000-0000-0000-000000000000\""
  run_cmd "./run.sh leave \"$tray_a_id\""
fi
if [ -n "$tray_b_id" ]; then
  run_cmd "./run.sh remote add \"e2e-${run_id}\" \"$tray_b_id\""
  run_cmd "./run.sh remote ls --format json"
  run_cmd "./run.sh remote remove \"e2e-${run_id}\""
fi
run_cmd "./run.sh listen --once --format json"

echo "Report written to: $REPORT_PATH"
