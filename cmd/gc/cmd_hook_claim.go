package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gastownhall/gascity/internal/beads"
)

const hookClaimCommandName = "hook"

var hookClaimMutationTimeout = 10 * time.Second

type hookClaimOptions struct {
	Assignee           string
	IdentityCandidates []string
	RouteTargets       []string
	Env                []string
	DrainAck           bool
	JSON               bool
}

type hookClaimOps struct {
	Runner             WorkQueryRunner
	Claim              hookClaimFunc
	ListContinuation   hookListContinuationFunc
	AssignContinuation hookAssignContinuationFunc
	DrainAck           hookDrainAckFunc
	Now                func() time.Time
}

type (
	hookClaimFunc              func(context.Context, string, []string, string, string) (beads.Bead, bool, error)
	hookListContinuationFunc   func(context.Context, string, []string, string, string) ([]beads.Bead, error)
	hookAssignContinuationFunc func(context.Context, string, []string, string, string) error
	hookDrainAckFunc           func(io.Writer) error
)

type hookClaimJSONResult struct {
	SchemaVersion        string   `json:"schema_version"`
	OK                   bool     `json:"ok"`
	Command              string   `json:"command"`
	Action               string   `json:"action"`
	Reason               string   `json:"reason,omitempty"`
	BeadID               string   `json:"bead_id,omitempty"`
	Assignee             string   `json:"assignee,omitempty"`
	Route                string   `json:"route,omitempty"`
	ContinuationAssigned []string `json:"continuation_assigned,omitempty"`
	DrainAcknowledged    bool     `json:"drain_acknowledged,omitempty"`
}

func doHookClaim(workQuery, dir string, opts hookClaimOptions, ops hookClaimOps, stdout, stderr io.Writer) int {
	opts.Assignee = strings.TrimSpace(opts.Assignee)
	opts.IdentityCandidates = hookClaimIdentityCandidates(append([]string{opts.Assignee}, opts.IdentityCandidates...)...)
	opts.RouteTargets = hookClaimRouteTargets(opts.RouteTargets...)
	if opts.Assignee == "" {
		fmt.Fprintln(stderr, "gc hook --claim: assignee not specified (set $GC_SESSION_NAME or $GC_SESSION_ID)") //nolint:errcheck
		return 1
	}
	if ops.Runner == nil {
		fmt.Fprintln(stderr, "gc hook --claim: missing work query runner") //nolint:errcheck
		return 1
	}
	if ops.Claim == nil {
		ops.Claim = hookClaimWithCLI
	}
	if ops.ListContinuation == nil {
		ops.ListContinuation = hookListContinuationWithBdStore
	}
	if ops.AssignContinuation == nil {
		ops.AssignContinuation = hookAssignContinuationWithBdStore
	}
	if ops.DrainAck == nil {
		ops.DrainAck = hookRuntimeDrainAck
	}
	now := time.Now
	if ops.Now != nil {
		now = ops.Now
	}

	output, err := ops.Runner(workQuery, dir)
	if err != nil {
		fmt.Fprintf(stderr, "gc hook --claim: %v\n", err) //nolint:errcheck
		return 1
	}

	normalized := normalizeWorkQueryOutput(strings.TrimSpace(output))
	normalized = filterUnreadyHookCandidates(normalized, now())
	if !workQueryHasReadyWork(normalized) {
		return writeHookClaimNoWork(opts, ops, stdout, stderr)
	}
	candidates, err := decodeHookClaimBeads(normalized)
	if err != nil {
		fmt.Fprintf(stderr, "gc hook --claim: requires JSON work_query output to identify claim candidates: %v\n", err) //nolint:errcheck
		return 1
	}
	if len(candidates) == 0 {
		return writeHookClaimNoWork(opts, ops, stdout, stderr)
	}

	if result, bead, ok := hookClaimExistingOrAssigned(candidates, opts); ok {
		return writeHookClaimWorkResultForBead(result, bead, opts, ops, dir, stdout, stderr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), hookClaimMutationTimeout)
	defer cancel()
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.ID) == "" ||
			strings.TrimSpace(candidate.Assignee) != "" ||
			!hookClaimMatchesRoute(candidate, opts.RouteTargets) {
			continue
		}
		claimed, ok, err := ops.Claim(ctx, dir, opts.Env, candidate.ID, opts.Assignee)
		if err != nil {
			fmt.Fprintf(stderr, "gc hook --claim: claiming %s: %v\n", candidate.ID, err) //nolint:errcheck
			return 1
		}
		if !ok {
			continue
		}
		if claimed.Metadata == nil {
			claimed.Metadata = candidate.Metadata
		}
		result := hookClaimJSONResult{
			SchemaVersion: "1",
			OK:            true,
			Command:       hookClaimCommandName,
			Action:        "work",
			Reason:        "claimed",
			BeadID:        claimed.ID,
			Assignee:      claimed.Assignee,
			Route:         hookClaimRoute(claimed),
		}
		if result.BeadID == "" {
			result.BeadID = candidate.ID
		}
		if result.Assignee == "" {
			result.Assignee = opts.Assignee
		}
		return writeHookClaimWorkResultForBead(result, claimed, opts, ops, dir, stdout, stderr)
	}

	return writeHookClaimNoWork(opts, ops, stdout, stderr)
}

func hookClaimWithCLI(_ context.Context, dir string, env []string, beadID, assignee string) (beads.Bead, bool, error) {
	claimed, err := hookClaimRunClaimCommand(dir, env, beadID, assignee)
	if err == nil {
		if claimed.ID == "" {
			claimed.ID = beadID
		}
		if claimed.Assignee == "" {
			claimed.Assignee = assignee
		}
		if hookClaimHasIdentity(claimed.Assignee, []string{assignee}) {
			return claimed, true, nil
		}
	}

	if current, getErr := hookClaimBdStore(dir, env, assignee).Get(beadID); getErr == nil {
		if hookClaimHasIdentity(current.Assignee, []string{assignee}) && strings.EqualFold(strings.TrimSpace(current.Status), "in_progress") {
			return current, true, nil
		}
		return beads.Bead{}, false, nil
	}
	if err != nil {
		return beads.Bead{}, false, err
	}
	return beads.Bead{}, false, nil
}

func hookClaimExistingOrAssigned(candidates []beads.Bead, opts hookClaimOptions) (hookClaimJSONResult, beads.Bead, bool) {
	for _, candidate := range candidates {
		if strings.EqualFold(strings.TrimSpace(candidate.Status), "in_progress") &&
			hookClaimHasIdentity(candidate.Assignee, opts.IdentityCandidates) {
			result := hookClaimJSONResult{
				SchemaVersion: "1",
				OK:            true,
				Command:       hookClaimCommandName,
				Action:        "work",
				Reason:        "existing_assignment",
				BeadID:        candidate.ID,
				Assignee:      candidate.Assignee,
				Route:         hookClaimRoute(candidate),
			}
			return result, candidate, true
		}
	}
	for _, candidate := range candidates {
		if strings.EqualFold(strings.TrimSpace(candidate.Status), "open") &&
			hookClaimHasIdentity(candidate.Assignee, opts.IdentityCandidates) {
			result := hookClaimJSONResult{
				SchemaVersion: "1",
				OK:            true,
				Command:       hookClaimCommandName,
				Action:        "work",
				Reason:        "ready_assignment",
				BeadID:        candidate.ID,
				Assignee:      candidate.Assignee,
				Route:         hookClaimRoute(candidate),
			}
			return result, candidate, true
		}
	}
	return hookClaimJSONResult{}, beads.Bead{}, false
}

func writeHookClaimWorkResultForBead(result hookClaimJSONResult, bead beads.Bead, opts hookClaimOptions, ops hookClaimOps, dir string, stdout, stderr io.Writer) int {
	assigned, err := preassignHookContinuationGroup(bead, opts, ops, dir)
	if err != nil {
		fmt.Fprintf(stderr, "gc hook --claim: preassigning continuation group for %s: %v\n", bead.ID, err) //nolint:errcheck
		return 1
	}
	result.ContinuationAssigned = assigned
	if opts.JSON {
		if err := writeCLIJSONLine(stdout, result); err != nil {
			fmt.Fprintf(stderr, "gc hook --claim: writing JSON: %v\n", err) //nolint:errcheck
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, result.BeadID) //nolint:errcheck
	return 0
}

func writeHookClaimNoWork(opts hookClaimOptions, ops hookClaimOps, stdout, stderr io.Writer) int {
	result := hookClaimJSONResult{
		SchemaVersion: "1",
		OK:            true,
		Command:       hookClaimCommandName,
		Action:        "drain",
		Reason:        "no_work",
	}
	if opts.DrainAck {
		if err := ops.DrainAck(stderr); err != nil {
			fmt.Fprintf(stderr, "gc hook --claim: drain-ack failed: %v\n", err) //nolint:errcheck
			return 1
		}
		result.DrainAcknowledged = true
	}
	if opts.JSON {
		if err := writeCLIJSONLine(stdout, result); err != nil {
			fmt.Fprintf(stderr, "gc hook --claim: writing JSON: %v\n", err) //nolint:errcheck
			return 1
		}
	}
	if opts.DrainAck {
		return 0
	}
	return 1
}

func preassignHookContinuationGroup(bead beads.Bead, opts hookClaimOptions, ops hookClaimOps, dir string) ([]string, error) {
	rootID := strings.TrimSpace(bead.Metadata["gc.root_bead_id"])
	group := strings.TrimSpace(bead.Metadata["gc.continuation_group"])
	if rootID == "" || group == "" {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), hookClaimMutationTimeout)
	defer cancel()
	siblings, err := ops.ListContinuation(ctx, dir, opts.Env, rootID, group)
	if err != nil {
		return nil, err
	}
	assigned := make([]string, 0, len(siblings))
	for _, sibling := range siblings {
		if strings.TrimSpace(sibling.ID) == "" ||
			sibling.ID == bead.ID ||
			strings.TrimSpace(sibling.Assignee) != "" ||
			!strings.EqualFold(strings.TrimSpace(sibling.Status), "open") ||
			!hookClaimMatchesRoute(sibling, opts.RouteTargets) {
			continue
		}
		if err := ops.AssignContinuation(ctx, dir, opts.Env, sibling.ID, opts.Assignee); err != nil {
			return assigned, fmt.Errorf("assigning %s: %w", sibling.ID, err)
		}
		assigned = append(assigned, sibling.ID)
	}
	return assigned, nil
}

func hookClaimRunClaimCommand(dir string, env []string, beadID, assignee string) (beads.Bead, error) {
	runner := beads.ExecCommandRunnerWithEnv(hookClaimEnvMap(env, dir, assignee))
	out, err := runner(dir, "bd", "update", beadID, "--claim", "--json")
	if err != nil {
		return beads.Bead{}, fmt.Errorf("bd update --claim %s: %w", beadID, err)
	}
	return decodeHookClaimBead(string(out))
}

func hookListContinuationWithBdStore(_ context.Context, dir string, env []string, rootID, group string) ([]beads.Bead, error) {
	store := hookClaimBdStore(dir, env, "")
	return store.List(beads.ListQuery{
		Status: "open",
		Metadata: map[string]string{
			"gc.root_bead_id":       rootID,
			"gc.continuation_group": group,
		},
		TierMode: beads.TierBoth,
	})
}

func hookAssignContinuationWithBdStore(_ context.Context, dir string, env []string, beadID, assignee string) error {
	store := hookClaimBdStore(dir, env, assignee)
	return store.Update(beadID, beads.UpdateOpts{Assignee: &assignee})
}

func hookRuntimeDrainAck(stderr io.Writer) error {
	if code := cmdRuntimeDrainAck(nil, false, io.Discard, stderr); code != 0 {
		return errors.New("runtime drain-ack returned non-zero")
	}
	return nil
}

func hookClaimBdStore(dir string, env []string, actor string) *beads.BdStore {
	return beads.NewBdStore(dir, beads.ExecCommandRunnerWithEnv(hookClaimEnvMap(env, dir, actor)))
}

func hookClaimEnvMap(env []string, dir string, actor string) map[string]string {
	env = workQueryEnvForDir(env, dir)
	out := make(map[string]string, len(env)+1)
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		out[key] = value
	}
	if strings.TrimSpace(actor) != "" {
		out["BEADS_ACTOR"] = actor
	}
	return out
}

func decodeHookClaimBeads(output string) ([]beads.Bead, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}
	if !json.Valid([]byte(output)) {
		extracted, ok := firstHookJSONValue(output)
		if !ok {
			return nil, errors.New("output is not JSON")
		}
		output = extracted
	}
	output = normalizeWorkQueryOutput(output)
	var candidates []beads.Bead
	if err := json.Unmarshal([]byte(output), &candidates); err != nil {
		return nil, err
	}
	return candidates, nil
}

func decodeHookClaimBead(output string) (beads.Bead, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return beads.Bead{}, errors.New("output is empty")
	}
	if !json.Valid([]byte(output)) {
		extracted, ok := firstHookJSONValue(output)
		if !ok {
			return beads.Bead{}, errors.New("output is not JSON")
		}
		output = extracted
	}
	output = normalizeWorkQueryOutput(output)
	if len(output) > 0 && output[0] == '[' {
		var beadsOut []beads.Bead
		if err := json.Unmarshal([]byte(output), &beadsOut); err != nil {
			return beads.Bead{}, err
		}
		if len(beadsOut) == 0 {
			return beads.Bead{}, errors.New("empty bead array")
		}
		return beadsOut[0], nil
	}
	var bead beads.Bead
	if err := json.Unmarshal([]byte(output), &bead); err != nil {
		return beads.Bead{}, err
	}
	return bead, nil
}

func firstHookJSONValue(output string) (string, bool) {
	for idx, r := range output {
		if r != '[' && r != '{' {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(output[idx:]))
		var raw json.RawMessage
		if err := dec.Decode(&raw); err == nil {
			return string(raw), true
		}
	}
	return "", false
}

func hookClaimHasIdentity(assignee string, identities []string) bool {
	assignee = strings.TrimSpace(assignee)
	if assignee == "" {
		return false
	}
	for _, identity := range identities {
		if assignee == strings.TrimSpace(identity) {
			return true
		}
	}
	return false
}

func hookClaimMatchesRoute(candidate beads.Bead, routeTargets []string) bool {
	if len(routeTargets) == 0 {
		return false
	}
	effectiveRoute := hookClaimRoute(candidate)
	if effectiveRoute == "" {
		return false
	}
	for _, target := range routeTargets {
		if effectiveRoute == strings.TrimSpace(target) {
			return true
		}
	}
	return false
}

func hookClaimRoute(candidate beads.Bead) string {
	if runTarget := strings.TrimSpace(candidate.Metadata["gc.run_target"]); runTarget != "" {
		return runTarget
	}
	return strings.TrimSpace(candidate.Metadata["gc.routed_to"])
}

func hookClaimIdentityCandidates(values ...string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
		if legacy := hookLegacyWorkflowControlName(value); legacy != "" && !seen[legacy] {
			seen[legacy] = true
			out = append(out, legacy)
		}
	}
	return out
}

func hookClaimRouteTargets(values ...string) []string {
	return hookClaimIdentityCandidates(values...)
}

func hookLegacyWorkflowControlName(value string) string {
	value = strings.TrimSpace(value)
	const suffix = "control-dispatcher"
	if !strings.HasSuffix(value, suffix) {
		return ""
	}
	return strings.TrimSuffix(value, suffix) + "workflow-control"
}
