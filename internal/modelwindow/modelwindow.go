// Package modelwindow resolves an LLM model ID to its context-window size in
// tokens. It is the single source of truth shared by the session-log context
// reader (internal/sessionlog) and the CLI context-pressure injector
// (cmd/gc/context_inject.go) so the two cannot resolve the same model ID to
// different windows.
package modelwindow

import "strings"

const (
	// Million is the context window, in tokens, for 1M-token model variants.
	Million = 1_000_000
	// Default is the conservative fallback window for a recognized Claude
	// family that is not a 1M variant (e.g. Haiku, Opus 4.5 and earlier).
	Default = 200_000
)

// millionMarkers force a 1M window when any is a substring of the model ID.
//
// Do NOT re-derive this list from the /v1/models reference (max_input_tokens).
// That oracle answers "what can the model accept", not "what does a
// subscription session actually get", and the two differ whenever the larger
// window is separately billed. Bare claude-sonnet-5 is 200K for us; the 1M
// variant is sonnet-5[1m], which is credit-billed and already caught by the
// "[1m]" marker. Marking a 200K model as 1M fails in the dangerous direction —
// the session under-reports usage and can overrun silently — so entries here
// must be justified against observed subscription behaviour, not the API
// reference. opus-4-5, opus-4-1 and haiku-4-5 are 200K and deliberately absent.
var millionMarkers = []string{
	"[1m]", "fable", "mythos",
	"opus-4-6", "opus-4-7", "opus-4-8", "opus-5",
	"sonnet-4-6",
}

// familyWindows pairs a model-family keyword with its context-window size, in
// longest-match-first order so a longer keyword wins over a shorter one it
// contains (e.g. "gpt-4o" before "gpt-4"). Claude families resolve to Default
// here; their 1M variants are caught earlier by millionMarkers.
var familyWindows = []struct {
	keyword string
	window  int
}{
	{"gpt-4o", 128_000},
	{"gpt-5", 258_000},
	{"gpt-4", 128_000},
	{"opus", Default},
	{"sonnet", Default},
	{"haiku", Default},
	{"gemini", Million},
	{"codex", 258_000},
}

// Window returns the context-window size, in tokens, for a model ID. Claude
// variants (Opus 4.6/4.7/4.8/5, Sonnet 4.6/5, Fable, Mythos) and any model
// carrying the explicit "[1m]" launch suffix resolve to the 1M window; older or
// unrecognized Claude variants use the 200K Default. Returns 0 when the model
// family is unrecognized, so callers can apply their own unknown-model policy
// (the session-log/API path treats 0 as "window unknown"; the injector floors
// it to Default).
func Window(model string) int {
	lower := strings.ToLower(model)
	for _, marker := range millionMarkers {
		if strings.Contains(lower, marker) {
			return Million
		}
	}
	for _, f := range familyWindows {
		if strings.Contains(lower, f.keyword) {
			return f.window
		}
	}
	return 0
}
