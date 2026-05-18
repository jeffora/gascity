package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gastownhall/gascity/internal/fsys"
)

// beadHooks maps bd hook filenames to the Gas City event types they emit.
var beadHooks = map[string]string{
	"on_create": "bead.created",
	"on_close":  "bead.closed",
	"on_update": "bead.updated",
}

// hookScript returns the shell script content for a bd hook that forwards
// events to the Gas City event log via gc-event-emit.
func hookScript(eventType, cityPath string) string {
	sockPath := shellDoubleQuoteDefault(filepath.Join(cityPath, ".gc", "events.sock"))
	pendingPath := shellDoubleQuoteDefault(filepath.Join(cityPath, ".gc", "events-pending.jsonl"))
	return fmt.Sprintf(`#!/bin/sh
# Installed by gc — forwards bd events to Gas City event log.
# Args: $1=issue_id  $2=event_type  stdin=issue JSON
GC_BIN="${GC_BIN:-gc}"
GC_EVENTS_SOCK="${GC_EVENTS_SOCK:-%s}"
GC_EVENTS_PENDING="${GC_EVENTS_PENDING:-%s}"
DATA=$(cat)
PAYLOAD=$(printf '{"bead":%%s}' "$DATA")
title=$(echo "$DATA" | grep -o '"title":"[^"]*"' | head -1 | cut -d'"' -f4)
(
  if command -v gc-event-emit >/dev/null 2>&1; then
    GC_EVENTS_SOCK="$GC_EVENTS_SOCK" GC_EVENTS_PENDING="$GC_EVENTS_PENDING" gc-event-emit %s --subject "$1" --message "$title" --payload "$PAYLOAD" >/dev/null 2>&1 || true
  else
    "$GC_BIN" event emit %s --subject "$1" --message "$title" --payload "$PAYLOAD" >/dev/null 2>&1 || true
  fi
) </dev/null >/dev/null 2>&1 &
`, sockPath, pendingPath, eventType, eventType)
}

// closeHookScript returns the on_close hook script. It forwards the
// bead.closed event, triggers convoy autoclose for the closed bead's
// parent convoy (if any), and auto-closes any open molecule/wisp
// children attached to the closed bead. Workflow-control watches the city
// event stream directly, so the close hook no longer sends a separate poke.
func closeHookScript(cityPath string) string {
	sockPath := shellDoubleQuoteDefault(filepath.Join(cityPath, ".gc", "events.sock"))
	pendingPath := shellDoubleQuoteDefault(filepath.Join(cityPath, ".gc", "events-pending.jsonl"))
	return fmt.Sprintf(`#!/bin/sh
# Installed by gc — forwards bd close events, auto-closes completed convoys,
# and auto-closes orphaned wisps.
# Args: $1=issue_id  $2=event_type  stdin=issue JSON
GC_BIN="${GC_BIN:-gc}"
GC_EVENTS_SOCK="${GC_EVENTS_SOCK:-%s}"
GC_EVENTS_PENDING="${GC_EVENTS_PENDING:-%s}"
DATA=$(cat)
PAYLOAD=$(printf '{"bead":%%s}' "$DATA")
title=$(echo "$DATA" | grep -o '"title":"[^"]*"' | head -1 | cut -d'"' -f4)
(
  if command -v gc-event-emit >/dev/null 2>&1; then
    GC_EVENTS_SOCK="$GC_EVENTS_SOCK" GC_EVENTS_PENDING="$GC_EVENTS_PENDING" gc-event-emit bead.closed --subject "$1" --message "$title" --payload "$PAYLOAD" >/dev/null 2>&1 || true
  else
    "$GC_BIN" event emit bead.closed --subject "$1" --message "$title" --payload "$PAYLOAD" >/dev/null 2>&1 || true
  fi
  # Auto-close parent convoy if all siblings are now closed.
  "$GC_BIN" convoy autoclose "$1" >/dev/null 2>&1 || true
  # Auto-close open molecule/wisp children so they don't outlive the parent.
  "$GC_BIN" wisp autoclose "$1" >/dev/null 2>&1 || true
) </dev/null >/dev/null 2>&1 &
`, sockPath, pendingPath)
}

// installBeadHooks writes bd hook scripts into dir/.beads/hooks/ so that
// bd mutations (create, close, update) emit events to the Gas City event
// log. Idempotent — leaves matching hooks in place. Returns nil on success.
func installBeadHooks(dir, cityPath string) error {
	hooksDir := filepath.Join(dir, ".beads", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	for filename, eventType := range beadHooks {
		path := filepath.Join(hooksDir, filename)
		content := hookScript(eventType, cityPath)
		if filename == "on_close" {
			content = closeHookScript(cityPath)
		}
		if err := fsys.WriteFileIfContentOrModeChangedAtomic(fsys.OSFS{}, path, []byte(content), 0o755); err != nil {
			return fmt.Errorf("writing hook %s: %w", filename, err)
		}
	}
	return nil
}

func shellDoubleQuoteDefault(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		`"`, `\"`,
		`$`, `\$`,
		"`", "\\`",
	)
	return replacer.Replace(s)
}
