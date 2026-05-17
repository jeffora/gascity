package main

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gastownhall/gascity/internal/events"
)

func TestDrainPendingEventFileRecordsAndClears(t *testing.T) {
	cityPath := t.TempDir()
	runtimeDir := filepath.Join(cityPath, ".gc")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pendingPath := cityEventPendingPath(cityPath)
	payload := json.RawMessage(`{"bead":{"id":"gc-1"}}`)
	line, err := json.Marshal(events.Event{
		Type:    events.BeadUpdated,
		Actor:   "bd-hook",
		Subject: "gc-1",
		Message: "update title",
		Payload: payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pendingPath, append(line, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	rec, err := events.NewFileRecorder(filepath.Join(runtimeDir, "events.jsonl"), &stderr)
	if err != nil {
		t.Fatalf("NewFileRecorder: %v", err)
	}
	defer rec.Close() //nolint:errcheck

	if err := drainPendingEventFile(cityPath, rec, &stderr); err != nil {
		t.Fatalf("drainPendingEventFile: %v; stderr=%s", err, stderr.String())
	}
	if _, err := os.Stat(pendingPath); !os.IsNotExist(err) {
		t.Fatalf("pending file still exists after drain: %v", err)
	}

	got, err := events.ReadFiltered(filepath.Join(runtimeDir, "events.jsonl"), events.Filter{})
	if err != nil {
		t.Fatalf("ReadFiltered: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("drained events = %d, want 1: %#v", len(got), got)
	}
	if got[0].Type != events.BeadUpdated || got[0].Actor != "bd-hook" || got[0].Subject != "gc-1" {
		t.Fatalf("drained event = %+v", got[0])
	}
	if got[0].Seq == 0 || got[0].Ts.IsZero() {
		t.Fatalf("drained event missing recorder seq/ts: %+v", got[0])
	}
}

func TestStartEventSocketRecordsJSONLine(t *testing.T) {
	cityPath := t.TempDir()
	runtimeDir := filepath.Join(cityPath, ".gc")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	rec, err := events.NewFileRecorder(filepath.Join(runtimeDir, "events.jsonl"), &stderr)
	if err != nil {
		t.Fatalf("NewFileRecorder: %v", err)
	}
	defer rec.Close() //nolint:errcheck

	server, err := startEventSocket(cityPath, rec, &stderr)
	if err != nil {
		t.Fatalf("startEventSocket: %v; stderr=%s", err, stderr.String())
	}
	defer server.Close() //nolint:errcheck

	conn, err := net.DialTimeout("unix", cityEventSocketPath(cityPath), time.Second)
	if err != nil {
		t.Fatalf("dial event socket: %v", err)
	}
	if err := json.NewEncoder(conn).Encode(events.Event{
		Type:    events.BeadCreated,
		Actor:   "bd-hook",
		Subject: "gc-2",
		Message: "created title",
	}); err != nil {
		conn.Close() //nolint:errcheck
		t.Fatalf("write event socket: %v", err)
	}
	conn.Close() //nolint:errcheck

	deadline := time.Now().Add(2 * time.Second)
	for {
		got, err := events.ReadFiltered(filepath.Join(runtimeDir, "events.jsonl"), events.Filter{})
		if err != nil {
			t.Fatalf("ReadFiltered: %v", err)
		}
		if len(got) == 1 {
			if got[0].Type != events.BeadCreated || got[0].Subject != "gc-2" {
				t.Fatalf("socket event = %+v", got[0])
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("event socket did not record event; stderr=%s", stderr.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
