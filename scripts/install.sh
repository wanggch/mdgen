#!/usr/bin/env bash
set -euo pipefail

BINARY_SRC=${BINARY_SRC:-"$(pwd)/mdgen"}
CONFIG_SRC=${CONFIG_SRC:-"$(pwd)/examples/config.yaml"}
LABEL=com.example.mdgen
PLIST=~/Library/LaunchAgents/${LABEL}.plist
TARGET_BIN=~/bin/mdgen
LOG_PATH=~/Library/Logs/mdgen.log
mkdir -p ~/bin ~/Library/Logs

if [ ! -f "$BINARY_SRC" ]; then
  echo "Binary not found at $BINARY_SRC" >&2
  exit 1
fi

cp "$BINARY_SRC" "$TARGET_BIN"
chmod +x "$TARGET_BIN"

CONFIG_TARGET=~/.config/mdgen/config.yaml
mkdir -p ~/.config/mdgen
cp "$CONFIG_SRC" "$CONFIG_TARGET"

cat deploy/com.example.mdgen.plist \
  | sed "s|{{BINARY_PATH}}|$TARGET_BIN|" \
  | sed "s|{{CONFIG_PATH}}|$CONFIG_TARGET|" \
  | sed "s|{{LOG_PATH}}|$LOG_PATH|" > "$PLIST"

launchctl bootout gui/"$(id -u)" "$PLIST" 2>/dev/null || true
launchctl load "$PLIST"
launchctl start "$LABEL"

echo "Installed $LABEL. Logs: $LOG_PATH"
