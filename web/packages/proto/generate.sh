#!/bin/bash
# =============================================================================
# Proto Codegen — TypeScript types & ConnectRPC service descriptors
# =============================================================================
# Uses protoc + protoc-gen-es (globally installed) since buf BSR auth is
# unavailable. Copies proto sources to a temp dir to avoid Unicode path issues.
#
# Prerequisites:
#   npm install -g @bufbuild/protoc-gen-es@latest
#   protoc installed at D:/softwares/protoc/
#
# Usage:
#   ./generate.sh                        # Generate ALL modules
#   ./generate.sh packages/proto         # Generate only shared packages
#   ./generate.sh business/masters/item  # Generate only item module
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# Hardcode backend path to avoid Unicode cd issues
BACKEND_DIR="/e/Brahma/Samavāya/backend"
FINAL_OUT="$SCRIPT_DIR/src/gen"

# Temp dirs (non-Unicode paths to avoid protoc issues)
TEMP_SRC="/e/tmp/proto-src"
TEMP_OUT="/e/tmp/proto-out"

# Plugin (must be globally installed to avoid Unicode in path)
PLUGIN="C:/Users/ppusa/AppData/Roaming/npm/protoc-gen-es.CMD"

# Include paths
PROTO_INCLUDE="D:/softwares/protoc/include"

# Find buf/validate protos in local cache
BUF_VALIDATE_DIR=""
if [ -n "$LOCALAPPDATA" ]; then
  VALIDATE_FILE=$(find "$LOCALAPPDATA/Buf" -name "validate.proto" -path "*/buf/validate/*" 2>/dev/null | head -1)
  if [ -n "$VALIDATE_FILE" ]; then
    # Go up 3 levels: buf/validate/validate.proto -> files/
    BUF_VALIDATE_DIR=$(dirname "$(dirname "$(dirname "$VALIDATE_FILE")")")
  fi
fi

echo "============================================"
echo "Proto Codegen for @samavāya/proto"
echo "============================================"
echo "Backend:     $BACKEND_DIR"
echo "Temp source: $TEMP_SRC"
echo "Temp output: $TEMP_OUT"
echo "Final:       $FINAL_OUT"
echo "Plugin:      $PLUGIN"
echo "Validate:    ${BUF_VALIDATE_DIR:-NOT FOUND}"
echo ""

# ─────────────────────────────────────────────────
# Step 1: Copy backend proto sources to temp dir
# ─────────────────────────────────────────────────
echo "Step 1: Copying proto sources to temp dir..."
rm -rf "$TEMP_SRC"
mkdir -p "$TEMP_SRC"
cp -r "$BACKEND_DIR/packages" "$TEMP_SRC/packages"
cp -r "$BACKEND_DIR/business" "$TEMP_SRC/business"
cp -r "$BACKEND_DIR/core" "$TEMP_SRC/core"
cp -r "$BACKEND_DIR/extension" "$TEMP_SRC/extension" 2>/dev/null || true
echo "  Done."
echo ""

# ─────────────────────────────────────────────────
# Step 2: Run protoc on all proto files
# ─────────────────────────────────────────────────
rm -rf "$TEMP_OUT"
mkdir -p "$TEMP_OUT"

ERRORS=0
GENERATED=0

generate_file() {
  local proto_file="$1"

  if protoc \
    --plugin=protoc-gen-es="$PLUGIN" \
    --es_out="$TEMP_OUT" \
    --es_opt=target=ts,import_extension=.js \
    --proto_path="$TEMP_SRC" \
    --proto_path="$PROTO_INCLUDE" \
    ${BUF_VALIDATE_DIR:+--proto_path="$BUF_VALIDATE_DIR"} \
    "$proto_file" 2>/dev/null; then
    GENERATED=$((GENERATED + 1))
  else
    echo "  FAIL: ${proto_file#$TEMP_SRC/}"
    ERRORS=$((ERRORS + 1))
  fi
}

generate_module() {
  local module_path="$1"
  local proto_dir="$TEMP_SRC/$module_path"

  # Some modules have proto/ subdirectory
  if [ -d "$proto_dir/proto" ]; then
    proto_dir="$proto_dir/proto"
  fi

  if [ ! -d "$proto_dir" ]; then
    return
  fi

  find "$proto_dir" -name "*.proto" 2>/dev/null | while read -r proto_file; do
    generate_file "$proto_file"
  done
}

# If a specific module is passed
if [ -n "$1" ]; then
  echo "Step 2: Generating $1..."
  generate_module "$1"
else
  echo "Step 2: Generating ALL modules..."

  # Generate all .proto files
  find "$TEMP_SRC" -name "*.proto" | while read -r proto_file; do
    generate_file "$proto_file"
  done
fi

# Also generate buf/validate types
if [ -n "$BUF_VALIDATE_DIR" ]; then
  echo ""
  echo "Step 2b: Generating buf/validate types..."
  find "$BUF_VALIDATE_DIR" -name "*.proto" | while read -r f; do
    protoc \
      --plugin=protoc-gen-es="$PLUGIN" \
      --es_out="$TEMP_OUT" \
      --es_opt=target=ts,import_extension=.js \
      --proto_path="$BUF_VALIDATE_DIR" \
      --proto_path="$PROTO_INCLUDE" \
      "$f" 2>/dev/null || true
  done
fi

echo ""

# ─────────────────────────────────────────────────
# Step 3: Copy generated files to final location
# ─────────────────────────────────────────────────
echo "Step 3: Copying generated files to $FINAL_OUT..."
rm -rf "$FINAL_OUT"
cp -r "$TEMP_OUT" "$FINAL_OUT"

TOTAL=$(find "$FINAL_OUT" -name "*.ts" | wc -l)

echo ""
echo "============================================"
echo "DONE"
echo "  Generated: $TOTAL TypeScript files"
echo "  Location:  $FINAL_OUT"
echo "============================================"

# ─────────────────────────────────────────────────
# Step 4: Cleanup temp files
# ─────────────────────────────────────────────────
rm -rf "$TEMP_SRC" "$TEMP_OUT"
echo "Temp files cleaned up."
