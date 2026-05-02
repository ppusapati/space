#!/usr/bin/env bash
# tools/check-trace.sh — verifies traceability between
#   plan/requirements.md (REQ-… IDs)
#   plan/todo.md         (TASK-… IDs and their Trace blocks)
#
# Reports:
#   1. Orphan tasks      — TASK-… entries with no REQ-… in their Trace line.
#   2. Coverage gaps     — REQ-… IDs in requirements.md that no TASK-… references.
#   3. OQ-mirror drift   — OQ-… IDs in requirements.md §9 vs todo.md §10.
#   4. Stub-language     — forbidden tokens (TODO|FIXME|XXX|tbd|figure out)
#                          inside plan/todo.md (REQ-CONST-010), excluding lines
#                          that quote the tokens (back-tick-fenced) or refer to
#                          them as forbidden.
#
# Exit code: 0 on full pass; 1 on any failure.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REQ="$ROOT/plan/requirements.md"
TODO="$ROOT/plan/todo.md"
FAIL=0

if [[ ! -f "$REQ" || ! -f "$TODO" ]]; then
  echo "ERROR: missing $REQ or $TODO" >&2
  exit 2
fi

# --- Extract sets ---------------------------------------------------------

# REQ IDs declared in requirements.md (LHS of the | column tables).
mapfile -t REQ_IDS < <(grep -oE '\| REQ-[A-Z0-9-]+' "$REQ" \
  | sed -E 's/^\| //' | sort -u)

# TASK IDs declared in todo.md — only real ### TASK-Pn-… or ### TASK-XC-… headings;
# the doc also shows a literal template form ### TASK-<phase>-<module>-<nnn> in §1
# which must be skipped.
mapfile -t TASK_IDS < <(grep -E '^### TASK-' "$TODO" \
  | sed -E 's/^### //; s/:.*$//' \
  | grep -E '^TASK-(P[0-6]|XC)-' \
  | sort -u)

# REQ IDs referenced by any task's Trace line.
mapfile -t REFERENCED_REQS < <(grep -E '^\*\*Trace:\*\*' "$TODO" \
  | grep -oE 'REQ-[A-Z0-9-]+' | sort -u)

# OQ IDs in requirements.md §9 and todo.md §10.
mapfile -t REQ_OQS  < <(grep -oE '^\| OQ-[0-9]+' "$REQ"  | sed -E 's/^\| //' | sort -u)
mapfile -t TODO_OQS < <(grep -oE '^\| OQ-[0-9]+' "$TODO" | sed -E 's/^\| //' | sort -u)

# --- Check 1: orphan tasks ------------------------------------------------
# A task heading is orphan iff the *next* `**Trace:**` line within its block
# (before the next `### TASK-` heading) contains no REQ-… reference.

ORPHAN_TASKS=()
for task_id in "${TASK_IDS[@]}"; do
  # extract the slice from this heading to the next TASK heading or EOF
  trace=$(awk -v id="$task_id" '
    $0 ~ "^### "id":" { capture=1; next }
    capture && /^### TASK-/ { exit }
    capture && /^\*\*Trace:\*\*/ { print; exit }
  ' "$TODO")
  if [[ -z "$trace" || ! "$trace" =~ REQ- ]]; then
    ORPHAN_TASKS+=("$task_id")
  fi
done

if (( ${#ORPHAN_TASKS[@]} > 0 )); then
  echo "FAIL: orphan tasks (no REQ- in Trace):"
  printf '  - %s\n' "${ORPHAN_TASKS[@]}"
  FAIL=1
else
  echo "PASS: every TASK references at least one REQ"
fi

# --- Check 2: coverage gaps -----------------------------------------------

UNCOVERED=()
for req in "${REQ_IDS[@]}"; do
  found=0
  for ref in "${REFERENCED_REQS[@]}"; do
    if [[ "$ref" == "$req" ]]; then found=1; break; fi
  done
  if (( ! found )); then UNCOVERED+=("$req"); fi
done

if (( ${#UNCOVERED[@]} > 0 )); then
  echo "FAIL: requirements with no covering task:"
  printf '  - %s\n' "${UNCOVERED[@]}"
  FAIL=1
else
  echo "PASS: every REQ in requirements.md is covered by at least one TASK"
fi

# --- Check 3: OQ mirror ---------------------------------------------------

diff_lr=$(comm -23 <(printf '%s\n' "${REQ_OQS[@]}") <(printf '%s\n' "${TODO_OQS[@]}") || true)
diff_rl=$(comm -13 <(printf '%s\n' "${REQ_OQS[@]}") <(printf '%s\n' "${TODO_OQS[@]}") || true)
if [[ -n "$diff_lr" ]]; then
  echo "FAIL: OQ in requirements.md but missing in todo.md §10:"
  printf '  - %s\n' $diff_lr
  FAIL=1
fi
if [[ -n "$diff_rl" ]]; then
  echo "FAIL: OQ in todo.md §10 but missing in requirements.md §9:"
  printf '  - %s\n' $diff_rl
  FAIL=1
fi
if [[ -z "$diff_lr" && -z "$diff_rl" ]]; then
  echo "PASS: OQ list mirrored exactly between requirements.md §9 and todo.md §10"
fi

# --- Check 4: forbidden stub language ------------------------------------
# Banned: TODO, FIXME, XXX, tbd, "figure out".
# (We do NOT ban the word "placeholder" because it legitimately appears in
# infra/Helm context like "placeholder hypertables" meaning empty hypertables.)
# We also exclude:
#   - back-tick-quoted token mentions (e.g. `TODO`)
#   - meta sentences in §1 that *describe* the forbidden tokens.

# Use case-sensitive grep so file paths like `plan/todo.md` don't match TODO.
# Quoted forms (`TODO`, "TODO") are excluded along with the §1 meta-sentence.
STUB_HITS=$(grep -nE '\b(TODO|FIXME|XXX|figure out)\b' "$TODO" \
  | grep -vE '`(TODO|FIXME|XXX|tbd|figure out)`' \
  | grep -vE '"(TODO|FIXME|XXX|tbd|figure out)"' \
  | grep -vE 'forbidden|MUST NOT contain|never mention|stub-language|stub language|No task in this document contains' \
  || true)
TBD_HITS=$(grep -nE '\bTBD\b|\btbd\b' "$TODO" \
  | grep -vE '`tbd`|"tbd"' \
  | grep -vE 'forbidden|MUST NOT contain|never mention|No task in this document contains' \
  || true)
if [[ -n "$STUB_HITS" || -n "$TBD_HITS" ]]; then
  echo "FAIL: forbidden stub language in todo.md (REQ-CONST-010):"
  [[ -n "$STUB_HITS" ]] && echo "$STUB_HITS"
  [[ -n "$TBD_HITS" ]]  && echo "$TBD_HITS"
  FAIL=1
else
  echo "PASS: no forbidden stub language in todo.md"
fi

# --- Summary --------------------------------------------------------------

echo
if (( FAIL )); then
  echo "RESULT: FAIL"
  exit 1
else
  echo "RESULT: PASS"
fi
