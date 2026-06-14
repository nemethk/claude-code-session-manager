#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== ccsm Uninstaller ==="
echo

# Remove binary
BINARY_PATH="/usr/local/bin/ccsm"
if [ -f "$BINARY_PATH" ]; then
    sudo rm "$BINARY_PATH"
    echo -e "${GREEN}✓${NC} Removed $BINARY_PATH"
else
    echo -e "${YELLOW}✓${NC} Binary already removed ($BINARY_PATH not found)"
fi

echo

# Ask about removing the Claude Code skill
SKILL_PATH="$HOME/.claude/skills/sessions.md"
if [ -f "$SKILL_PATH" ]; then
    read -p "Remove Claude Code skill ($SKILL_PATH)? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm "$SKILL_PATH"
        echo -e "${GREEN}✓${NC} Removed $SKILL_PATH"
    else
        echo -e "${YELLOW}✓${NC} Skipped skill removal"
    fi
else
    echo -e "${YELLOW}✓${NC} Skill not installed ($SKILL_PATH not found)"
fi

echo

echo -e "${GREEN}✓${NC} Uninstall complete"
echo
echo "Note: your session data in ~/.claude/projects/ is untouched."
