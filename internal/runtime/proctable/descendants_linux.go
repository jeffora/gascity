//go:build linux

package proctable

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// snapshotProcesses walks /proc for a host-wide pid/ppid/comm table.
func snapshotProcesses() ([]ProcessRecord, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var records []ProcessRecord
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		ppid, ok, err := readParentPID(filepath.Join("/proc", e.Name(), "stat"))
		if err != nil || !ok {
			continue
		}
		comm, err := os.ReadFile(filepath.Join("/proc", e.Name(), "comm"))
		if err != nil {
			continue
		}
		records = append(records, ProcessRecord{PID: pid, PPID: ppid, Name: strings.TrimSpace(string(comm))})
	}
	return records, nil
}
