#!/bin/bash
# Samavaya WASM Build Script
# Builds all Rust crates to WebAssembly using wasm-pack

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG_DIR="$SCRIPT_DIR/pkg"
CRATES_DIR="$SCRIPT_DIR/crates"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Samavaya WASM Build System${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"

# Check for wasm-pack
if ! command -v wasm-pack &> /dev/null; then
    echo -e "${RED}Error: wasm-pack is not installed${NC}"
    echo -e "${YELLOW}Install it with: cargo install wasm-pack${NC}"
    exit 1
fi

# Check for Rust
if ! command -v cargo &> /dev/null; then
    echo -e "${RED}Error: Rust/Cargo is not installed${NC}"
    echo -e "${YELLOW}Install it from: https://rustup.rs${NC}"
    exit 1
fi

# Parse arguments
TARGET="web"
BUILD_MODE="release"
CRATE_FILTER=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --target)
            TARGET="$2"
            shift 2
            ;;
        --dev)
            BUILD_MODE="dev"
            shift
            ;;
        --crate)
            CRATE_FILTER="$2"
            shift 2
            ;;
        --help)
            echo "Usage: build.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --target <target>  Build target: web, nodejs, bundler (default: web)"
            echo "  --dev              Build in development mode (faster, larger)"
            echo "  --crate <name>     Build only specific crate"
            echo "  --help             Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Create output directory
mkdir -p "$PKG_DIR"

# Build mode flag
if [ "$BUILD_MODE" = "dev" ]; then
    BUILD_FLAG="--dev"
    echo -e "${YELLOW}Building in DEVELOPMENT mode${NC}"
else
    BUILD_FLAG="--release"
    echo -e "${GREEN}Building in RELEASE mode${NC}"
fi

echo -e "${BLUE}Target: ${TARGET}${NC}"
echo ""

# Count crates
TOTAL_CRATES=0
BUILT_CRATES=0
FAILED_CRATES=0

for crate_dir in "$CRATES_DIR"/*/; do
    if [ -f "$crate_dir/Cargo.toml" ]; then
        TOTAL_CRATES=$((TOTAL_CRATES + 1))
    fi
done

# Build each crate
for crate_dir in "$CRATES_DIR"/*/; do
    if [ -f "$crate_dir/Cargo.toml" ]; then
        crate_name=$(basename "$crate_dir")

        # Skip if filter is set and doesn't match
        if [ -n "$CRATE_FILTER" ] && [ "$crate_name" != "$CRATE_FILTER" ]; then
            continue
        fi

        echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
        echo -e "${BLUE}Building: ${crate_name}${NC}"

        out_dir="$PKG_DIR/$crate_name"

        if wasm-pack build "$crate_dir" \
            --target "$TARGET" \
            --out-dir "$out_dir" \
            $BUILD_FLAG \
            --out-name "samavaya_${crate_name}" 2>&1; then

            echo -e "${GREEN}✓ Built: ${crate_name}${NC}"
            BUILT_CRATES=$((BUILT_CRATES + 1))

            # Clean up unnecessary files
            rm -f "$out_dir/.gitignore"
            rm -f "$out_dir/package-lock.json"
        else
            echo -e "${RED}✗ Failed: ${crate_name}${NC}"
            FAILED_CRATES=$((FAILED_CRATES + 1))
        fi
    fi
done

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Build Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Successful: ${BUILT_CRATES}${NC}"
echo -e "${RED}Failed: ${FAILED_CRATES}${NC}"
echo -e "${BLUE}Total: ${TOTAL_CRATES}${NC}"

# Generate index file
INDEX_FILE="$PKG_DIR/index.ts"
echo "// Auto-generated WASM module index" > "$INDEX_FILE"
echo "// Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$INDEX_FILE"
echo "" >> "$INDEX_FILE"

for crate_dir in "$CRATES_DIR"/*/; do
    if [ -f "$crate_dir/Cargo.toml" ]; then
        crate_name=$(basename "$crate_dir")
        if [ -d "$PKG_DIR/$crate_name" ]; then
            # Convert kebab-case to camelCase for export name
            export_name=$(echo "$crate_name" | sed -r 's/(^|-)(\w)/\U\2/g' | sed 's/^./\l&/')
            echo "export * as ${export_name} from './${crate_name}/samavaya_${crate_name}';" >> "$INDEX_FILE"
        fi
    fi
done

echo ""
echo -e "${GREEN}Generated: ${INDEX_FILE}${NC}"

if [ $FAILED_CRATES -eq 0 ]; then
    echo ""
    echo -e "${GREEN}Build completed successfully!${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}Build completed with ${FAILED_CRATES} failures${NC}"
    exit 1
fi
