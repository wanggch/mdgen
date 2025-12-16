#!/usr/bin/env bash
set -euo pipefail

LABEL=com.example.mdgen
PLIST=~/Library/LaunchAgents/${LABEL}.plist
TARGET_BIN=~/bin/mdgen
LOG_PATH=~/Library/Logs/mdgen.log

launchctl bootout gui/"$(id -u)" "$PLIST" 2>/dev/null || true
rm -f "$PLIST" "$TARGET_BIN"
rm -f "$LOG_PATH"

echo "Uninstalled $LABEL"
