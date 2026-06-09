# Kwame Asante — Gemini (Independent Review, Attempt 6)

**Verdict:** approve-with-risks

**Persona:** minimal vocabulary, facade creep, event-log deferral, backlog scope control.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 6, matching `.gc/design-reviews/ga-unpr2y/attempt-6/design-before.md` with Attempt 6 changes), `internal/session/lifecycle_projection.go`, and prior design iterations.

---

## Top Strengths

- **Clean Decoupling of Side Effects (Repair and Materialization) from Candidate Classification**:
  The addition of the explicit paragraph under target classification stating *"Repair and materialization are not classifier behavior. The raw classifier collects candidate facts and may return `repair-needed` or `requires-materialization`; it must not call `RepairEmptyType`, create a session, retire an identity, or rewrite metadata..."* (`DESIGN.md:634-641`) is a major YAGNI victory. It strictly enforces a separation of read-only token analysis from write-heavy or side-effecting operations, preventing the raw classifier from becoming an active mutator or universal repair facade on day one.

- **Durable Scan-First Recovery and Event-Log Deferral**:
  The event contract has been updated to clarify that: *"Current events are thin hints unless an event-migration slice explicitly changes typed payloads. The fields below are required durable fact identities for scans; they become wire payload fields only when the slice updates typed payloads..."* (`DESIGN.md:764-776`). This perfectly aligns with the lane's core principle of event-log deferral, preventing the premature scaffolding of complex event-delivery and outbox mechanisms on top of the in-process event model.

- **Highly Complete and Dynamic Mutation Guards**:
  The inclusion of additive exceptions (W-022 to W-028) and the requirement for positive guard fixtures for raw store writes, package-level mutators (`WakeSession`, `RepairEmptyType`), and dynamic metadata batches represents robust backlog scope control. It ensures that the transition path is tightly bracketed by static analysis, making any bypass immediately detectable.

- **Provisional Bound Controls for Later Slices**:
  Adding the `provisional bounds` checkpoint field and labeling slices 3-7 contracts as provisional scaffolding that will be finalized when their respective slices start prevents the entire 1300+ line document from being treated as a frozen, Day-One delivery list.

---

## Critical Risks & Risks

### [Major] 1. `TargetResolutionPolicy` Struct and Construction Remain Undefined

While the design enumerates nine explicit policy flags (`surface`, `allow_closed`, `materialize_named`, `allow_path_alias`, `allow_historical_alias`, `read_only`, `allow_template_factory`, `allow_ordinary_config_target`, and `reject_config_missing_named`), the actual type representation, fields, and construction of this policy input remain completely undefined in the vocabulary.
- **The Vocabulary Gap**: The shared vocabulary checkpoint table (`DESIGN.md:439-446`) defines `TargetCandidate` and `TargetSelection`, but leaves `TargetResolutionPolicy` (or the equivalent policy input struct) entirely out of the checkpoint list.
- **Day-One Scope Inflation**: If the raw classifier in Slice 1 is forced to consume all nine policy flags in a single struct, Slice 1 will introduce a vocabulary surface that spans six different surfaces before those surfaces even begin their migration.
- *Required Change*: Define the policy input type (e.g., `TargetResolutionPolicy`), list the exact fields that Slice 1 requires, and clarify if the classifier accepts a single policy struct or if individual surface adapters post-filter the raw candidate results.

---

### [Major] 2. Contradiction in Raw Classification of Negative Kinds

The design specifies that repair and materialization are not classifier behavior (`DESIGN.md:634`), yet the classifier's typed result contract still includes policy-dependent negative kinds such as `forbidden-kind`, `requires-materialization`, and `closed-not-allowed` (`DESIGN.md:648-650`).
- **The Architectural Conflict**: Whether a configured reserved target is forbidden or requires materialization is an operational policy decision that belongs to the adapter (e.g., mail query does not materialize, while API creation does). If the classifier is a raw read-only candidate collector, it cannot produce these negative kinds without executing policy, which directly loops back to the undefined policy input.
- *Required Change*: Either remove `forbidden-kind` and `closed-not-allowed` from the classifier's raw `negative_kind` vocabulary (letting adapters apply policy to raw candidate outputs), or explicitly declare that the classifier takes a policy input parameter and list the required fields under Slice-1 vocabulary.

---

### [Major] 3. Diagnostic Operability Contract Risks Facade Creep in Read-Only Slices

The diagnostic operability contract (`DESIGN.md:941-951`) mandates that "Every decision, command, and conflict result must carry" sixteen categories of evidence, including:
- `generation, instance token, runtime session key when relevant`
- `event emission result and subscriber/recovery path used`
- `lifecycle projection, blocker, wake cause`
While "when relevant" was added to the token fields, there is no similar exemption for event emission, lifecycle projection, or blocker fields. For raw target classification (Slice 1), which is a read-only token lookup that emits no events and does not evaluate lifecycle state, requiring these fields would force the implementation to build a diagnostics facade with stubbed or empty values.
- *Required Change*: Explicitly state that the operability field list is an eventual union across all command clusters, and each result type carries only the diagnostic fields its adopted surfaces consume. Add a per-operation scoping note: *"Classification results carry identity, target kind, candidates, and conflict details; generation, instance token, event emission, and wake cause are command-cluster fields."*

---

### [Minor] 4. `SessionFactEvent` Lacks an Explicit Owner Slice

`SessionFactEvent` remains listed in the vocabulary checkpoints (`DESIGN.md:445`), but no slice in the backlog is explicitly tasked with introducing it. Since Slice 1 is read-only and Slices 2-4 are the earliest event-emitting candidates, leaving this type unanchored risks premature introduction in Slice 1.
- *Required Change*: Assign the introduction of `SessionFactEvent` to a specific event-emitting slice (e.g., Slice 2 or Slice 3), and explicitly document that Slice 1 introduces no event vocabulary.

---

### [Minor] 5. Surface-Specific `path-alias` Candidate Resolution Boundaries

The precedence table lists `path-alias` candidates at position 8 (`DESIGN.md:604`). For API path alias fallback, the API scans the session titles in memory. If the classifier is expected to perform this title-matching store scan internally, it violates the read-only classification budget.
- *Required Change*: Document that `path-alias` candidates are supplied to the classifier by the API-adapter's title-matching step, rather than discovered by the classifier via a general store scan.

---

### [Minor] 6. Missing `closeFailedCreateBead` Cross-Family Writer

`cmd/gc/session_beads.go:1736-1750` (`closeFailedCreateBead`) writes `pending_create_claim`, `pending_create_started_at` (create/start family, Slice 3) and `sleep_intent` (wake/hold/drain family, Slice 5) in a single call. However, this cross-family mutation is not flagged in the mutation landscape table under exception W-008.
- *Required Change*: Add a note to W-008 or create a separate inventory entry marking `closeFailedCreateBead` as a cross-family writer with owner Slices 3 and 5 noted.

---

## Answers to Persona Questions

1. **Does vocabulary introduced match what the first implementing slice actually needs?**
   - *Answer*: Mostly yes. The separation of candidate kinds from mutation and materialization keeps Slice 1 vocabulary focused. However, the undefined policy input struct and the presence of command-only fields in the operability contract pose a minor facade creep risk.
2. **Does TR-007's future-compat language pull more structure into today's contracts than the in-process event model requires?**
   - *Answer*: No. The updated event contract successfully limits current events to thin hints and defers wire payload structures to event-migration slices, resolving this risk completely.
3. **Is anything on a path to becoming a broad facade?**
   - *Answer*: Yes. The operability contract and the negative classifier kinds (`forbidden-kind`, `requires-materialization`) still risk pulling downstream policy and diagnostics into Slice 1.

---

## Missing Evidence

- A concrete definition of the `TargetResolutionPolicy` input struct and its fields for Slice 1.
- Explicit mapping or population rules for classifier results showing that `materializable` and `candidates[]` are empty/omitted for exact ID matches.
- Integration of `closeFailedCreateBead` into the mutation landscape table as a cross-family writer.

---

## Required Changes

1. **Define the policy input type**: Name the struct (e.g., `TargetResolutionPolicy`) and list the minimum fields Slice 1 requires, clarifying whether individual surfaces pass this struct directly or if adapters wrap the classifier.
2. **Reconcile raw classifier negative kinds**: Either remove `forbidden-kind` and `closed-not-allowed` from the raw classifier's output vocabulary (letting the adapter apply policy), or explicitly list the policy input as a classifier parameter in Slice 1.
3. **Scope the operability contract per operation**: Add an explicit note that read-only classification results are exempt from command-only diagnostic fields like generation, instance token, event emission result, and wake cause.
4. **Assign event vocabulary**: Anchor `SessionFactEvent` to a specific event-emitting slice (e.g., Slice 2) and state that Slice 1 introduces no event vocabulary.
5. **Classify path-alias candidate supply**: Document that `path-alias` candidates are adapter-supplied rather than classifier-scanned.
6. **Update Mutation Inventory (W-008)**: List `closeFailedCreateBead` explicitly as a cross-family writer touching create/start and wake/hold/drain families, with owner Slices 3 and 5 noted.

---
