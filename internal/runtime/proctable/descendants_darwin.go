//go:build darwin

package proctable

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// snapshotProcesses shells out to `ps` for a host-wide pid/ppid/comm table.
func snapshotProcesses() ([]ProcessRecord, error) {
	out, err := exec.Command("ps", "-ax", "-o", "pid=,ppid=,comm=").Output()
	if err != nil {
		return nil, fmt.Errorf("running ps: %w", err)
	}
	var records []ProcessRecord
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		records = append(records, ProcessRecord{PID: pid, PPID: ppid, Name: filepath.Base(fields[2])})
	}
	return records, nil
}
