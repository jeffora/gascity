package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gastownhall/gascity/internal/config"
	"github.com/gastownhall/gascity/internal/doctor"
)

type eventEmitBinaryDoctorCheck struct {
	lookPath doctor.LookPathFunc
}

func newEventEmitBinaryDoctorCheck(lookPath doctor.LookPathFunc) *eventEmitBinaryDoctorCheck {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	return &eventEmitBinaryDoctorCheck{lookPath: lookPath}
}

func (c *eventEmitBinaryDoctorCheck) Name() string { return "gc-event-emit-binary" }

func (c *eventEmitBinaryDoctorCheck) Run(_ *doctor.CheckContext) *doctor.CheckResult {
	r := &doctor.CheckResult{Name: c.Name()}
	path, err := c.lookPath("gc-event-emit")
	if err != nil {
		r.Status = doctor.StatusWarning
		r.Message = "gc-event-emit not found in PATH"
		r.FixHint = "install gc-event-emit alongside gc so bead hooks use the event socket fast path"
		return r
	}
	r.Status = doctor.StatusOK
	r.Message = fmt.Sprintf("found %s", path)
	return r
}

func (c *eventEmitBinaryDoctorCheck) CanFix() bool { return false }

func (c *eventEmitBinaryDoctorCheck) Fix(_ *doctor.CheckContext) error { return nil }

func (c *eventEmitBinaryDoctorCheck) WarmupEligible() bool { return false }

type eventEmitHookPathDoctorCheck struct {
	cityPath  string
	hookRoots []string
}

func newEventEmitHookPathDoctorCheck(cityPath string, hookRoots []string) *eventEmitHookPathDoctorCheck {
	return &eventEmitHookPathDoctorCheck{cityPath: cityPath, hookRoots: hookRoots}
}

func eventEmitHookRoots(cityPath string, cfg *config.City) []string {
	roots := []string{cityPath}
	if cfg == nil {
		return roots
	}
	for _, rig := range cfg.Rigs {
		if strings.TrimSpace(rig.Path) != "" {
			roots = append(roots, rig.Path)
		}
	}
	return roots
}

func (c *eventEmitHookPathDoctorCheck) Name() string { return "gc-event-emit-hooks" }

func (c *eventEmitHookPathDoctorCheck) Run(_ *doctor.CheckContext) *doctor.CheckResult {
	r := &doctor.CheckResult{Name: c.Name()}
	expectedSock := filepath.Join(c.cityPath, ".gc", "events.sock")
	expectedPending := filepath.Join(c.cityPath, ".gc", "events-pending.jsonl")
	checked := 0

	for _, root := range c.hookRoots {
		if strings.TrimSpace(root) == "" {
			continue
		}
		for filename := range beadHooks {
			path := filepath.Join(root, ".beads", "hooks", filename)
			data, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				r.Status = doctor.StatusError
				r.Message = fmt.Sprintf("reading bead hook %s: %v", path, err)
				return r
			}
			text := string(data)
			if !strings.Contains(text, "gc-event-emit") {
				continue
			}
			checked++
			if !strings.Contains(text, expectedSock) || !strings.Contains(text, expectedPending) {
				r.Details = append(r.Details, path)
			}
		}
	}

	if len(r.Details) > 0 {
		r.Status = doctor.StatusWarning
		r.Message = fmt.Sprintf("%d gc-event-emit hook(s) embed a stale city path", len(r.Details))
		r.FixHint = "reinstall bead hooks so they embed the current city event socket path"
		return r
	}
	r.Status = doctor.StatusOK
	if checked == 0 {
		r.Message = "no gc-event-emit bead hooks found"
	} else {
		r.Message = fmt.Sprintf("%d gc-event-emit hook(s) point at current city", checked)
	}
	return r
}

func (c *eventEmitHookPathDoctorCheck) CanFix() bool { return false }

func (c *eventEmitHookPathDoctorCheck) Fix(_ *doctor.CheckContext) error { return nil }

func (c *eventEmitHookPathDoctorCheck) WarmupEligible() bool { return false }
