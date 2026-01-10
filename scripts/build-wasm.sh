#!/bin/bash

# Build WASM for browser usage
# This script builds the phonid package as WebAssembly and prepares necessary files

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OUTPUT_DIR="dist/wasm"
WASM_FILE="phonid.wasm"
PACKAGE="./cmd/phonid-wasm"

echo -e "${BLUE}Building Phonid WASM...${NC}"

# Get git short SHA for cache busting
GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
export GIT_SHA

echo -e "${GREEN}→${NC} Using git SHA: $GIT_SHA"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build WASM
echo -e "${GREEN}→${NC} Building WASM module..."
GOOS=js GOARCH=wasm go build -o "$OUTPUT_DIR/phonid.$GIT_SHA.wasm" "$PACKAGE"

# Copy wasm_exec.js helper
echo -e "${GREEN}→${NC} Copying wasm_exec.js..."
GOROOT=$(go env GOROOT)

# Try multiple possible locations for wasm_exec.js
if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" "$OUTPUT_DIR/wasm_exec.$GIT_SHA.js"
elif [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/lib/wasm/wasm_exec.js" "$OUTPUT_DIR/wasm_exec.$GIT_SHA.js"
else
    echo -e "${GREEN}→${NC} Downloading wasm_exec.js from Go repository..."
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    curl -sL "https://raw.githubusercontent.com/golang/go/release-branch.go${GO_VERSION%.*}/misc/wasm/wasm_exec.js" -o "$OUTPUT_DIR/wasm_exec.$GIT_SHA.js"
    if [ ! -f "$OUTPUT_DIR/wasm_exec.$GIT_SHA.js" ]; then
        echo "Error: Could not find or download wasm_exec.js"
        exit 1
    fi
fi

# Create a simple HTML file for testing
echo -e "${GREEN}→${NC} Creating test HTML from template..."
envsubst < "$(dirname "$0")/wasm-index.html" > "$OUTPUT_DIR/index.html"

# Minify inline JavaScript if terser is available
if command -v terser &> /dev/null; then
    echo -e "${GREEN}→${NC} Minifying inline JavaScript..."
    # Extract and minify JavaScript between <script> tags
    # Note: This is a basic approach - for production, consider a proper HTML minifier
    temp_html="$OUTPUT_DIR/index.html.tmp"
    temp_js="$OUTPUT_DIR/temp.js"

    # Extract JavaScript (between last <script> tags that don't have src attribute)
    awk '/<script>/{flag=1; next} /<\/script>/{flag=0} flag' "$OUTPUT_DIR/index.html" > "$temp_js"

    if [ -s "$temp_js" ]; then
        # Minify the extracted JavaScript
        terser "$temp_js" --compress --mangle --output "$temp_js.min" 2>/dev/null || cp "$temp_js" "$temp_js.min"

        # Replace the JavaScript in HTML with minified version
        awk -v minjs="$temp_js.min" '
            /<script>/ && !/<script src=/ {
                print $0
                while ((getline line < minjs) > 0) print line
                close(minjs)
                # Skip original JS content
                while (getline && !/<\/script>/) {}
                print "</script>"
                next
            }
            {print}
        ' "$OUTPUT_DIR/index.html" > "$temp_html"

        mv "$temp_html" "$OUTPUT_DIR/index.html"
        rm -f "$temp_js" "$temp_js.min"
        echo -e "${GREEN}→${NC} JavaScript minified successfully"
    fi
else
    echo -e "${BLUE}→${NC} terser not found - skipping JavaScript minification (install with: npm install -g terser)"
fi

# Create README for WASM usage
cat > "$OUTPUT_DIR/README.md" <<'EOF'
# Phonid WASM Build

This directory contains the WebAssembly build of Phonid.

## Files

- `phonid.wasm` - The compiled WebAssembly module
- `wasm_exec.js` - Go's WebAssembly runtime helper
- `index.html` - Test/demo HTML page

## Usage

### Serve Locally

You need to serve these files over HTTP (not file://):

```bash
# Using Python
python3 -m http.server -d dist/wasm 8080

# Using npx
npx serve dist/wasm
```

Then open http://localhost:8080 in your browser.

### Integration

To integrate into your web application:

1. Include `wasm_exec.js` in your HTML
2. Load and instantiate `phonid.wasm`
3. Implement JavaScript bridge functions in Go using `syscall/js`

See index.html for a basic example.

## Next Steps

The current build is a basic WASM compilation. To make it functional:

1. Create a browser-specific entry point (e.g., `cmd/phonid-wasm/main.go`)
2. Use `syscall/js` to expose Go functions to JavaScript
3. Handle file uploads instead of filesystem access
4. Update the HTML/JS to call the exposed functions

EOF

echo ""
echo -e "${GREEN}✓ Build complete!${NC}"
echo ""
echo "Files created in: $OUTPUT_DIR/"
echo "  - phonid.$GIT_SHA.wasm ($(du -h "$OUTPUT_DIR/phonid.$GIT_SHA.wasm" | cut -f1))"
echo "  - wasm_exec.$GIT_SHA.js"
echo "  - index.html"
echo "  - README.md"
echo ""
echo "To test:"
echo "  python3 -m http.server -d $OUTPUT_DIR 8080"
echo "  Then open http://localhost:8080"
echo ""
