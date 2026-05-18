package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gastownhall/gascity/internal/doctor"
)

func TestEventEmitBinaryDoctorCheckWarnsWhenUnavailable(t *testing.T) {
	check := newEventEmitBinaryDoctorCheck(func(name string) (string, error) {
		if name != "gc-event-emit" {
			t.Fatalf("lookPath name = %q, want gc-event-emit", name)
		}
		return "", errors.New("not found")
	})

	result := check.Run(&doctor.CheckContext{})
	if result.Status != doctor.StatusWarning {
		t.Fatalf("status = %v, want warning; result=%+v", result.Status, result)
	}
}

func TestEventEmitBinaryDoctorCheckOKWhenAvailable(t *testing.T) {
	check := newEventEmitBinaryDoctorCheck(func(name string) (string, error) {
		if name != "gc-event-emit" {
			t.Fatalf("lookPath name = %q, want gc-event-emit", name)
		}
		return "/usr/local/bin/gc-event-emit", nil
	})

	result := check.Run(&doctor.CheckContext{})
	if result.Status != doctor.StatusOK {
		t.Fatalf("status = %v, want OK; result=%+v", result.Status, result)
	}
}

func TestEventEmitHookPathDoctorCheckWarnsOnStaleCityPath(t *testing.T) {
	cityPath := t.TempDir()
	hooksDir := filepath.Join(cityPath, ".beads", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	staleHook := `#!/bin/sh
GC_EVENTS_SOCK="${GC_EVENTS_SOCK:-/old-city/.gc/events.sock}"
GC_EVENTS_PENDING="${GC_EVENTS_PENDING:-/old-city/.gc/events-pending.jsonl}"
gc-event-emit bead.updated
`
	if err := os.WriteFile(filepath.Join(hooksDir, "on_update"), []byte(staleHook), 0o755); err != nil {
		t.Fatal(err)
	}

	result := newEventEmitHookPathDoctorCheck(cityPath, []string{cityPath}).Run(&doctor.CheckContext{})
	if result.Status != doctor.StatusWarning {
		t.Fatalf("status = %v, want warning; result=%+v", result.Status, result)
	}
	if len(result.Details) != 1 || !strings.Contains(result.Details[0], "on_update") {
		t.Fatalf("details = %#v, want stale on_update detail", result.Details)
	}
}

func TestEventEmitHookPathDoctorCheckOKWhenHooksMatchCityPath(t *testing.T) {
	cityPath := t.TempDir()
	if err := installBeadHooks(cityPath, cityPath); err != nil {
		t.Fatalf("installBeadHooks: %v", err)
	}

	result := newEventEmitHookPathDoctorCheck(cityPath, []string{cityPath}).Run(&doctor.CheckContext{})
	if result.Status != doctor.StatusOK {
		t.Fatalf("status = %v, want OK; result=%+v", result.Status, result)
	}
}
