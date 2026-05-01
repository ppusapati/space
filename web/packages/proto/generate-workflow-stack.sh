#!/usr/bin/env bash
# Regenerate TypeScript bindings for the workflow/approval/audit proto stack.
# We bypass `buf generate` because the BSR requires an API token we don't have.
# Dependency protos (buf.validate, googleapis) are pulled from the local buf cache
# that was populated by prior buf runs.
#
# Usage: run from anywhere — script cd's to the backend directory.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_ROOT="$(cd "$HERE/../.." && pwd)"
BACKEND_ROOT="$(cd "$WEB_ROOT/../backend" && pwd)"

BUF_CACHE="${LOCALAPPDATA:-$HOME/.local/share}/buf/v3/modules/b5"
if [[ ! -d "$BUF_CACHE" ]]; then
  BUF_CACHE="/c/Users/$USER/AppData/Local/buf/v3/modules/b5"
fi

PROTOVALIDATE_DIR="$(find "$BUF_CACHE/buf.build/bufbuild/protovalidate" -maxdepth 2 -type d -name files 2>/dev/null | sort | tail -1)"
GOOGLEAPIS_DIR="$(find "$BUF_CACHE/buf.build/googleapis/googleapis" -maxdepth 2 -type d -name files 2>/dev/null | sort | tail -1)"

if [[ -z "$PROTOVALIDATE_DIR" || -z "$GOOGLEAPIS_DIR" ]]; then
  echo "ERROR: buf cache missing — run 'buf dep update' with a valid BSR token once, then retry." >&2
  exit 1
fi

cd "$BACKEND_ROOT"

# Proto files needed for the approval inbox + submission history UIs.
# Keep this list minimal — every entry is a file the frontend consumes directly
# or via transitive TS imports.
PROTO_FILES=(
  # core approval service (registered in monolith, serves approval inbox)
  core/workflow/approval/proto/approval.proto

  # workflow engine (user tasks, execution history)
  core/workflow/workflow/proto/workflow.proto

  # audit read service (submission history, change log)
  core/audit/audit/proto/audit.proto

  # form instance (may still be needed for future submission-list RPC)
  core/workflow/formbuilder/proto/forminstance.proto
)

for proto in "${PROTO_FILES[@]}"; do
  if [[ ! -f "$proto" ]]; then
    echo "⚠️  Skipping missing proto: $proto" >&2
    continue
  fi
done

OUT_REL="../web/packages/proto/src/gen"

PATH="$WEB_ROOT/node_modules/.bin:$PATH" protoc \
  --proto_path=. \
  --proto_path="$PROTOVALIDATE_DIR" \
  --proto_path="$GOOGLEAPIS_DIR" \
  --es_out="$OUT_REL" \
  --es_opt=target=ts \
  --es_opt=import_extension=.js \
  "${PROTO_FILES[@]}"

echo "✓ Regenerated workflow stack proto bindings"
for p in "${PROTO_FILES[@]}"; do
  name="$(basename "$p" .proto)"
  echo "  - $name"
done
