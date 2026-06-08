package main

import (
	"strings"

	"github.com/gastownhall/gascity/internal/beads"
)

func routedToOrLegacyWorkflowTarget(b beads.Bead) string {
	if runTarget := strings.TrimSpace(b.Metadata["gc.run_target"]); runTarget != "" {
		return runTarget
	}
	return strings.TrimSpace(b.Metadata["gc.routed_to"])
}

func routedToAndLegacyWorkflowCandidates(b beads.Bead) []string {
	runTarget := strings.TrimSpace(b.Metadata["gc.run_target"])
	routedTo := strings.TrimSpace(b.Metadata["gc.routed_to"])
	if runTarget == "" {
		if routedTo == "" {
			return nil
		}
		return []string{routedTo}
	}
	return []string{runTarget}
}

// WorkRoutingModel returns the advisory per-dispatch model choice carried by a
// work bead in its "gc.model" metadata, or "" when none is set. This is the
// key the model-advisor pack and the mol-review-quorum formula already write;
// it is consumed at session spawn so an advised model applies per task/shape
// rather than only per agent. The value is a provider OptionsSchema "model"
// choice value (e.g. "opus", "sonnet"); validation against the resolved
// provider's schema happens at the spawn site.
func WorkRoutingModel(b beads.Bead) string {
	return strings.TrimSpace(b.Metadata["gc.model"])
}
