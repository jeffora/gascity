package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gastownhall/gascity/internal/citylayout"
	"github.com/gastownhall/gascity/internal/events"
)

const (
	eventSocketMaxLineBytes = 1 << 20
	eventSocketReadTimeout  = 2 * time.Second
	eventSocketAcceptLimit  = 64
)

type eventSocketServer struct {
	listener net.Listener
	path     string
	done     chan struct{}
	closed   atomic.Bool
	once     sync.Once
}

func cityEventSocketPath(cityPath string) string {
	return filepath.Join(cityPath, citylayout.RuntimeRoot, "events.sock")
}

func cityEventPendingPath(cityPath string) string {
	return filepath.Join(cityPath, citylayout.RuntimeRoot, "events-pending.jsonl")
}

func startEventSocket(cityPath string, rec events.Recorder, stderr io.Writer) (*eventSocketServer, error) {
	if rec == nil {
		return nil, errors.New("event recorder is nil")
	}
	path := cityEventSocketPath(cityPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("creating event socket directory: %w", err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("removing stale event socket: %w", err)
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listening on event socket: %w", err)
	}
	if ul, ok := listener.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(false)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		listener.Close() //nolint:errcheck
		os.Remove(path)  //nolint:errcheck
		return nil, fmt.Errorf("chmod event socket: %w", err)
	}
	server := &eventSocketServer{
		listener: listener,
		path:     path,
		done:     make(chan struct{}),
	}
	go server.serve(rec, stderr)
	return server, nil
}

func (s *eventSocketServer) serve(rec events.Recorder, stderr io.Writer) {
	defer close(s.done)
	sem := make(chan struct{}, eventSocketAcceptLimit)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			fmt.Fprintf(stderr, "gc events socket: accept: %v\n", err) //nolint:errcheck
			continue
		}
		select {
		case sem <- struct{}{}:
			go func() {
				defer func() { <-sem }()
				handleEventSocketConn(conn, rec, stderr)
			}()
		default:
			conn.Close()                                                       //nolint:errcheck
			fmt.Fprintln(stderr, "gc events socket: connection limit reached") //nolint:errcheck
		}
	}
}

func (s *eventSocketServer) Close() error {
	var err error
	s.once.Do(func() {
		s.closed.Store(true)
		err = s.listener.Close()
		<-s.done
		if removeErr := os.Remove(s.path); removeErr != nil && !os.IsNotExist(removeErr) && err == nil {
			err = removeErr
		}
	})
	return err
}

func handleEventSocketConn(conn net.Conn, rec events.Recorder, stderr io.Writer) {
	defer conn.Close() //nolint:errcheck
	if err := conn.SetReadDeadline(time.Now().Add(eventSocketReadTimeout)); err != nil {
		fmt.Fprintf(stderr, "gc events socket: deadline: %v\n", err) //nolint:errcheck
		return
	}
	limited := io.LimitReader(conn, eventSocketMaxLineBytes+1)
	line, err := bufio.NewReader(limited).ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintf(stderr, "gc events socket: read: %v\n", err) //nolint:errcheck
		return
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return
	}
	if len(line) > eventSocketMaxLineBytes {
		fmt.Fprintln(stderr, "gc events socket: event line too large") //nolint:errcheck
		return
	}
	var event events.Event
	if err := json.Unmarshal(line, &event); err != nil {
		fmt.Fprintf(stderr, "gc events socket: decode event: %v\n", err) //nolint:errcheck
		return
	}
	if event.Type == "" {
		fmt.Fprintln(stderr, "gc events socket: event type is required") //nolint:errcheck
		return
	}
	rec.Record(event)
}

func drainPendingEventFile(cityPath string, rec events.Recorder, stderr io.Writer) error {
	if rec == nil {
		return nil
	}
	pendingPath := cityEventPendingPath(cityPath)
	drainPath := filepath.Join(
		filepath.Dir(pendingPath),
		fmt.Sprintf("events-pending.%d.%d.drain", os.Getpid(), time.Now().UnixNano()),
	)
	if err := os.Rename(pendingPath, drainPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("renaming pending event file: %w", err)
	}
	defer os.Remove(drainPath) //nolint:errcheck

	f, err := os.Open(drainPath)
	if err != nil {
		return fmt.Errorf("opening pending event file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), eventSocketMaxLineBytes)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var event events.Event
		if err := json.Unmarshal(line, &event); err != nil {
			fmt.Fprintf(stderr, "gc events pending: decode event: %v\n", err) //nolint:errcheck
			continue
		}
		if event.Type == "" {
			fmt.Fprintln(stderr, "gc events pending: event type is required") //nolint:errcheck
			continue
		}
		rec.Record(event)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading pending event file: %w", err)
	}
	return nil
}

func drainRunningCityPendingEvents(cr *cityRegistry, stderr io.Writer) {
	type cityRecorder struct {
		path string
		rec  events.Recorder
	}
	var recorders []cityRecorder
	cr.ReadCallback(func(
		cities map[string]*managedCity,
		_ map[string]cityInitProgress,
		_ map[string]*initFailRecord,
		_ map[string]*panicRecord,
	) {
		for path, mc := range cities {
			if mc != nil && mc.eventRecorder != nil {
				recorders = append(recorders, cityRecorder{path: path, rec: mc.eventRecorder})
			}
		}
	})
	for _, r := range recorders {
		if err := drainPendingEventFile(r.path, r.rec, stderr); err != nil {
			fmt.Fprintf(stderr, "gc supervisor: city '%s': drain pending events: %v\n", filepath.Base(r.path), err) //nolint:errcheck
		}
	}
}
