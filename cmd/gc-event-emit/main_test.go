package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGcEventEmitWritesSingleJSONLineToSocket(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "events.sock")
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close() //nolint:errcheck
	if ul, ok := ln.(*net.UnixListener); ok {
		if err := ul.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Fatalf("SetDeadline: %v", err)
		}
	}

	lines := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			lines <- "accept error: " + err.Error()
			return
		}
		defer conn.Close() //nolint:errcheck
		line, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			lines <- "read error: " + err.Error()
			return
		}
		lines <- line
	}()

	var stderr bytes.Buffer
	code := run([]string{
		"bead.updated",
		"--subject", "gc-1",
		"--message", "updated title",
		"--payload", `{"bead":{"id":"gc-1"}}`,
	}, func(key string) string {
		switch key {
		case "GC_EVENTS_SOCK":
			return sockPath
		case "GC_EVENTS_PENDING":
			return filepath.Join(dir, "events-pending.jsonl")
		default:
			return ""
		}
	}, &stderr)
	if code != 0 {
		t.Fatalf("run = %d; stderr=%s", code, stderr.String())
	}

	var line string
	select {
	case line = <-lines:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for socket line")
	}
	if strings.HasPrefix(line, "accept error:") || strings.HasPrefix(line, "read error:") {
		t.Fatal(line)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("socket line is not JSON: %v\n%s", err, line)
	}
	if got["type"] != "bead.updated" || got["actor"] != "bd-hook" || got["subject"] != "gc-1" || got["message"] != "updated title" {
		t.Fatalf("unexpected event: %#v", got)
	}
	payload, ok := got["payload"].(map[string]any)
	if !ok {
		t.Fatalf("payload = %#v, want object", got["payload"])
	}
	if _, ok := payload["bead"].(map[string]any); !ok {
		t.Fatalf("payload missing bead object: %#v", payload)
	}
}

func TestGcEventEmitFallsBackToPendingFileWhenSocketUnavailable(t *testing.T) {
	dir := t.TempDir()
	pendingPath := filepath.Join(dir, "events-pending.jsonl")

	var stderr bytes.Buffer
	code := run([]string{
		"bead.created",
		"--subject", "gc-2",
		"--message", "created title",
	}, func(key string) string {
		switch key {
		case "GC_EVENTS_SOCK":
			return filepath.Join(dir, "missing.sock")
		case "GC_EVENTS_PENDING":
			return pendingPath
		default:
			return ""
		}
	}, &stderr)
	if code != 0 {
		t.Fatalf("run = %d; stderr=%s", code, stderr.String())
	}

	data, err := os.ReadFile(pendingPath)
	if err != nil {
		t.Fatalf("reading pending file: %v; stderr=%s", err, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("pending lines = %d, want 1:\n%s", len(lines), string(data))
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("pending line is not JSON: %v\n%s", err, lines[0])
	}
	if got["type"] != "bead.created" || got["actor"] != "bd-hook" || got["subject"] != "gc-2" || got["message"] != "created title" {
		t.Fatalf("unexpected pending event: %#v", got)
	}
}
