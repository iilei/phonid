#!/bin/bash

# Helper script for creating public presets
# This guides users through the preset creation process

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}Phonid Public Preset Creator${NC}"
echo "================================"
echo ""

# Check if phonid is installed, if not build it locally
if ! command -v phonid &> /dev/null; then
    echo -e "${YELLOW}phonid CLI not found, building locally...${NC}"

    # Check if goreleaser is installed
    if ! command -v goreleaser &> /dev/null; then
        echo "Installing goreleaser..."
        go install github.com/goreleaser/goreleaser/v2@latest
    fi

    # Build phonid
    echo "Building phonid..."
    goreleaser build --snapshot --clean --single-target

    # Find the built binary
    PHONID_BIN=$(find dist -name "phonid" -type f | head -1)

    if [ -z "$PHONID_BIN" ]; then
        echo -e "${RED}Error: Failed to build phonid${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Built phonid locally: $PHONID_BIN"
else
    PHONID_BIN="phonid"
fi

# Get preset name
echo ""
echo "Enter a name for your preset (lowercase, hyphens allowed):"
echo "Example: minion-speak"
read -r preset_name

# Validate name
if [[ ! "$preset_name" =~ ^[a-z0-9-]+$ ]]; then
    echo -e "${RED}Error: Invalid name. Use only lowercase letters, numbers, and hyphens${NC}"
    exit 1
fi

filename=".${preset_name}.phonidrc.toml"
mkdir -p public_presets/temp/
filepath="public_presets/temp/$filename"
final_file="public_presets/$filename"

# Check if file already exists
if [ -f "$filepath" ]; then
    echo -e "${RED}Error: $filepath already exists${NC}"
    exit 1
fi

# Generate configuration
echo ""
echo -e "${GREEN}→${NC} Generating configuration..."
"$PHONID_BIN" preflight --suggest > "$filepath"

echo -e "${GREEN}✓${NC} Created: $filepath"

# Sanity check
echo ""
echo -e "${GREEN}→${NC} Running sanity check ..."
if "$PHONID_BIN" --config "$filepath" preflight --suggest; then
    echo -e "${GREEN}✓${NC} All preflight checks passed"
else
    echo -e "${RED}✗${NC} sanity check  failed"
    echo "Please review the errors above and try again"
    rm "$filepath"
    exit 1
fi

# Iterative editing loop
while true; do
    echo ""
    echo -e "${YELLOW}(i)${NC} Take your time to make changes at ${filepath}"
    echo ""
    echo "Press Enter when you're done editing to see the results..."
    read -r

    # Show results of the changes
    echo ""
    echo -e "${GREEN}→${NC} Running preflight check with your changes..."
    if "$PHONID_BIN" --config "$filepath" preflight --suggest; then
        echo -e "${GREEN}✓${NC} All preflight checks passed"

        # Ask if user wants to continue editing or proceed
        echo ""
        echo "Do you want to:"
        echo "  1) Continue editing"
        echo "  2) Proceed with this configuration"
        read -r -p "Choice [1/2]: " choice

        if [[ "$choice" == "2" ]]; then
            break
        fi
    else
        echo -e "${RED}✗${NC} Preflight checks failed"
        echo "You must fix the errors before proceeding."
    fi
done

# write to final_file
"$PHONID_BIN" --config "$filepath" preflight --suggest >| "$final_file"
rm "$filepath"

echo -e "${GREEN}✓${NC} Done"
// ...existing code...

echo -e "${GREEN}✓${NC} Done"
echo ""
echo "Next steps:"
echo "1. Review the generated config at: $final_file"
echo "2. Create a pull request with your preset"
echo ""

# Generate GitHub issue URL with prefilled parameters
title="Add ${preset_name} public preset"
# URL encode the title (replace spaces with %20)
encoded_title=$(echo "$title" | sed 's/ /%20/g')
issue_url="https://github.com/iilei/phonid/issues/new?template=public-preset-contribution.md&title=${encoded_title}&labels=preset-contribution"

echo "Then use this link to open a PR with pre-filled details:"
echo "$issue_url"
