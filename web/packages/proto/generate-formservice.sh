#!/usr/bin/env bash
# Regenerate TypeScript bindings for core/platform/formservice using protoc directly.
# We bypass `buf generate` because the BSR requires an API token we don't have.
# Dependency protos (buf.validate, googleapis) are pulled from the local buf cache
# that was populated by prior buf runs.
#
# Usage: run from the repo root or anywhere — script cd's to the backend directory.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_ROOT="$(cd "$HERE/../.." && pwd)"
BACKEND_ROOT="$(cd "$WEB_ROOT/../backend" && pwd)"

# Newest cached snapshot of each BSR module. protovalidate schemas change occasionally
# but any recent vintage produces identical TS output for the fields we use.
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

PATH="$WEB_ROOT/node_modules/.bin:$PATH" protoc \
  --proto_path=. \
  --proto_path="$PROTOVALIDATE_DIR" \
  --proto_path="$GOOGLEAPIS_DIR" \
  --es_out="$WEB_ROOT/packages/proto/src/gen" \
  --es_opt=target=ts \
  --es_opt=import_extension=.js \
  core/platform/formservice/proto/formservice.proto

echo "✓ Regenerated formservice_pb.ts"
