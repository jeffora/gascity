// Command gc-event-emit writes one event to the supervisor event socket.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gastownhall/gascity/internal/events"
)

const eventEmitWriteTimeout = 200 * time.Millisecond

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stderr))
}

func run(args []string, getenv func(string) string, stderr io.Writer) int {
	var subject, message, actor, payload string
	fs := flag.NewFlagSet("gc-event-emit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "gc-event-emit %s (%s, %s)\n", version, commit, date)                            //nolint:errcheck
		fmt.Fprintln(stderr, "usage: gc-event-emit <type> [--subject id] [--message text] [--payload json]") //nolint:errcheck
	}
	fs.StringVar(&subject, "subject", "", "Event subject")
	fs.StringVar(&message, "message", "", "Event message")
	fs.StringVar(&actor, "actor", "bd-hook", "Actor name")
	fs.StringVar(&payload, "payload", "", "JSON payload")
	eventType := ""
	parseArgs := args
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		eventType = args[0]
		parseArgs = args[1:]
	}
	if err := fs.Parse(parseArgs); err != nil {
		return 2
	}
	if eventType == "" && fs.NArg() == 1 {
		eventType = fs.Arg(0)
	}
	if eventType == "" {
		fmt.Fprintln(stderr, "gc-event-emit: missing event type") //nolint:errcheck
		return 2
	}

	event := events.Event{
		Type:    eventType,
		Actor:   actor,
		Subject: subject,
		Message: message,
	}
	if payload != "" {
		if !json.Valid([]byte(payload)) {
			fmt.Fprintln(stderr, "gc-event-emit: --payload is not valid JSON") //nolint:errcheck
			return 0
		}
		event.Payload = json.RawMessage(payload)
	}
	line, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(stderr, "gc-event-emit: marshal event: %v\n", err) //nolint:errcheck
		return 1
	}
	line = append(line, '\n')

	sockPath := getenv("GC_EVENTS_SOCK")
	if sockPath != "" {
		if err := writeEventSocket(sockPath, line); err == nil {
			return 0
		}
	}

	pendingPath := getenv("GC_EVENTS_PENDING")
	if pendingPath == "" {
		fmt.Fprintln(stderr, "gc-event-emit: GC_EVENTS_PENDING is not set") //nolint:errcheck
		return 1
	}
	if err := appendPendingEvent(pendingPath, line); err != nil {
		fmt.Fprintf(stderr, "gc-event-emit: append pending event: %v\n", err) //nolint:errcheck
		return 1
	}
	return 0
}

func writeEventSocket(sockPath string, line []byte) error {
	conn, err := net.DialTimeout("unix", sockPath, eventEmitWriteTimeout)
	if err != nil {
		return err
	}
	defer conn.Close() //nolint:errcheck
	if err := conn.SetWriteDeadline(time.Now().Add(eventEmitWriteTimeout)); err != nil {
		return err
	}
	_, err = conn.Write(line)
	return err
}

func appendPendingEvent(path string, line []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating pending event directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("locking pending event file: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN) //nolint:errcheck
	if _, err := f.Write(line); err != nil {
		return err
	}
	return nil
}
