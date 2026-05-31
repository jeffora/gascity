package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/gastownhall/gascity/internal/beads"
	"github.com/gastownhall/gascity/internal/config"
	"github.com/gastownhall/gascity/internal/runtime"
)

// Investigation artifact for ga-odfazq.
//
// The bead's stated root cause is: "The reconciler trusts a stored,
// pre-resolved provider string on the session/template bead instead of
// re-resolving the provider from the template + current config at start
// time." These two tests bracket the actual behavior and show that claim is
// MISDIAGNOSED for the buildDesiredState skip path:
//
//   - StaleStoredProviderIsIgnored: a session bead carrying a stale
//     provider="claude-sonnet" in its metadata does NOT cause a skip, because
//     buildDesiredState re-resolves the provider from live cfg every tick
//     (findAgentByTemplate / findNamedSessionSpec -> spec.Agent). The stored
//     string is never consulted by the desired-state resolution. This
//     DISPROVES the "trusts a stored provider" hypothesis.
//
//   - UndefinedLiveProviderSilentlySkips: when LIVE cfg names a provider that
//     is not defined ([providers]) and not a builtin, the configured named
//     session is silently dropped from desired state with only an stderr
//     line. This REPRODUCES the genuine 05/23 failure mode — which was a live
//     config error (workspace/agent provider naming an undefined provider),
//     since healed — and shows the real remaining defect: silent skip, no
//     surfaced/observable error beyond stderr.

// TestBuildDesiredState_StaleStoredProviderIsIgnored_LiveConfigWins proves the
// reconciler does NOT trust a stored provider string on a session bead. The
// agent resolves from live config (StartCommand escape hatch); a session bead
// carrying provider="claude-sonnet" (the value the bead worries about) is
// present in the store but never enters provider resolution, so the agent is
// not skipped.
func TestBuildDesiredState_StaleStoredProviderIsIgnored_LiveConfigWins(t *testing.T) {
	cityPath := t.TempDir()
	store := beads.NewMemStore()

	// A persisted session bead for "mayor" carrying a STALE stored provider.
	// This is exactly the "stale stored provider string captured into a
	// session/template desired-state bead" the bead describes.
	if _, err := store.Create(beads.Bead{
		Title:  "mayor",
		Type:   sessionBeadType,
		Labels: []string{sessionBeadLabel, "agent:mayor", "template:mayor"},
		Metadata: map[string]string{
			"template":     "mayor",
			"agent_name":   "mayor",
			"alias":        "mayor",
			"session_name": "s-mayor",
			"state":        "awake",
			"provider":     "claude-sonnet", // stale, no longer in config
		},
	}); err != nil {
		t.Fatalf("create mayor session bead: %v", err)
	}

	// Live config is HEALED: the mayor agent resolves fine. StartCommand is the
	// provider escape hatch, so resolution is deterministic and PATH-independent
	// — and crucially it has nothing to do with the stored "claude-sonnet".
	cfg := &config.City{
		Workspace: config.Workspace{Name: "test-city"},
		Agents: []config.Agent{{
			Name:         "mayor",
			StartCommand: "true",
		}},
		NamedSessions: []config.NamedSession{{
			Name:     "mayor",
			Template: "mayor",
			Mode:     "always",
		}},
	}

	var stderr bytes.Buffer
	dsResult := buildDesiredState("test-city", cityPath, time.Now().UTC(), cfg, runtime.NewFake(), store, &stderr)

	if strings.Contains(stderr.String(), "(skipping)") {
		t.Fatalf("mayor was skipped despite healthy live config; the stored provider must not drive resolution.\nstderr = %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "claude-sonnet") {
		t.Fatalf("stale stored provider \"claude-sonnet\" leaked into desired-state resolution; it must never be consulted.\nstderr = %q", stderr.String())
	}
	if len(dsResult.State) == 0 {
		t.Fatalf("mayor missing from desired state; expected the always-mode named session to be present.\nstderr = %q", stderr.String())
	}
}

// TestBuildDesiredState_UndefinedLiveProviderSilentlySkipsNamedSession
// reproduces the genuine defect: when LIVE cfg pins the agent to a provider
// that is neither defined in [providers] nor a builtin, ResolveProvider returns
// ErrProviderNotFound and the configured named session is silently dropped from
// desired state. The ONLY observability is an stderr line — there is no
// surfaced error, event, or alert. This is the real bug behind ga-odfazq (the
// live config has since been healed, so it is no longer reproducible from the
// current city.toml — only from a deliberately-broken config like this one).
func TestBuildDesiredState_UndefinedLiveProviderSilentlySkipsNamedSession(t *testing.T) {
	cityPath := t.TempDir()
	store := beads.NewMemStore()

	// LIVE cfg names a provider that does not exist. No [providers] entry, not a
	// builtin. This is the 05/23 shape: workspace/agent provider -> undefined.
	cfg := &config.City{
		Workspace: config.Workspace{Name: "test-city"},
		Agents: []config.Agent{{
			Name:     "mayor",
			Provider: "claude-sonnet", // undefined -> ErrProviderNotFound
		}},
		NamedSessions: []config.NamedSession{{
			Name:     "mayor",
			Template: "mayor",
			Mode:     "always",
		}},
	}

	var stderr bytes.Buffer
	dsResult := buildDesiredState("test-city", cityPath, time.Now().UTC(), cfg, runtime.NewFake(), store, &stderr)

	// The named session is dropped from desired state...
	for key := range dsResult.State {
		if strings.Contains(key, "mayor") {
			t.Fatalf("expected mayor to be dropped from desired state on unresolvable provider, but found key %q", key)
		}
	}
	// ...with only an stderr line as evidence (the "silent skip" defect).
	s := stderr.String()
	for _, want := range []string{"named session", "unknown provider", "claude-sonnet", "(skipping)"} {
		if !strings.Contains(s, want) {
			t.Fatalf("stderr missing %q; got %q", want, s)
		}
	}
}
