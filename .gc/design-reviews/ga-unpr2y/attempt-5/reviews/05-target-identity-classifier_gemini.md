# Amara Diallo — DeepSeek V4 Flash (Independent Review, Attempt 5)

**Verdict:** block

**Persona:** target taxonomy, alias precedence, conflict cases, caller behavior preservation, cross-document consistency.

**Reviewed against:** `internal/session/DESIGN.md` (attempt 5, `.gc/design-reviews/ga-unpr2y/attempt-5/design-before.md`), `internal/session/resolve.go`, `internal/session/named_config.go`, `internal/session/lifecycle_projection.go`, `internal/api/session_resolution.go`, `internal/api/handler_beads.go`, `internal/api/handler_mail.go`, `internal/mail/resolve.go`, `internal/dispatch/control.go`, attempt-1 through attempt-4 reviews (all target-identity-classifier lanes), and attempt-5 cross-persona reviews (mutation-boundary-auditor, event-delivery-contract, api-cli-worker-boundary, yagni-contract-scope).

---

## Top Strengths

- **Right First Extraction Slice**: The design continues to correctly isolate target classification as the first extraction slice (`DESIGN.md:362`, `DESIGN.md:982`). Separating token identification from mutating operational policies is the cleanest approach to defining an enforceable session boundary.
- **Accurate Candidate Kind Taxonomy**: The candidate kinds enumerated at `DESIGN.md:507-520` cleanly capture complex real-world situations. In particular, representing `rejected-by-config` and `configured-named-conflict` as native classification result kinds maps beautifully to active API logic, such as `apiSessionTargetRejectedByConfig` at `session_resolution.go:451`.
- **Fidelity of Huma API Compatibility Chains**: The per-surface compatibility resolver chains (`DESIGN.md:522-537`) map Huma API live and allow-closed routes with high precision. Specifying that named session config lookups may materialize only when the caller passes explicit policy inputs is a solid design invariant.
- **Decoupling of Policy Inputs**: Defining explicit policy flags (`DESIGN.md:538-542`) ensures the classifier does not silently reorder surface-specific precedence or inject mutating side-effects.

---

## Critical Risks (Blockers)

### [Blocker] 1. Precedence Table Mismatch (Candidate Kinds vs. API Compatibility Chain)

The `Candidate kinds` table (`DESIGN.md:507-520`) lists `live-session-name` and `live-alias` as high-ranking kinds (positions 2 and 3), while `configured-named-canonical` appears as a separate kind later in the list.

However, both the actual API live chain (`DESIGN.md:528`) and the source implementation (`session_resolution.go:441-445`) specify:
```go
if id, matched, err := s.resolveConfiguredNamedSessionIDWithContext(ctx, store, identifier, opts); err == nil {
    return id, nil
}
```
Which runs *before* falling through to `ResolveSessionID` (which handles live session_name/alias). Thus, `configured-named` has a *higher* precedence than generic package-level live names for API targets. Stating the candidate kinds in an apparently ordered list where live names appear before configured-named identities continues to create misleading assumptions for implementers.

* **Required Change**: Either (a) reorder the Candidate Kinds table to match the canonical precedence of the live API target resolver, or (b) explicitly document in the table introduction that the Candidate Kinds list is an *unordered pool*, and that precedence rules are defined purely by each surface's compatibility resolver chain.

---

### [Blocker] 2. Direct Exact-ID Closed-Bead Resolution Inaccuracy

The package-resolver compatibility chain (`DESIGN.md:526`) states that `ResolveSessionID` handles exact ID matches first, and that `ResolveSessionIDAllowClosed` handles closed session_name/alias fallbacks.

However, looking at the source code, `ResolveSessionBeadByExactID` (`resolve.go:50-63`) loads a bead by exact ID via `store.Get(identifier)` and returns it immediately if `IsSessionBeadOrRepairable(b)` is true, **regardless of whether the bead is open or closed**. This means that closed exact IDs successfully resolve on *both* `ResolveSessionID` (live-only) and `ResolveSessionIDAllowClosed` (allow-closed) paths. The `allowClosed` parameter only gates the session_name/alias fallbacks.

* **Required Change**: Document this behavior clearly in the compatibility chains table (`DESIGN.md:526-527`). State that exact-ID lookup returns closed beads on all surfaces, and that `allowClosed` parameter only restricts the session_name and alias fallback matching.

---

### [Blocker] 3. `RepairEmptyType` Store Mutation inside "Read-Only" Target Classification

The design characterizes target classification as having "none; read-only snapshot plus policy" side effects under the Command Atomicity Contract (`DESIGN.md:623`).

However, `ResolveSessionBeadByExactID` (`resolve.go:50-63`) calls `RepairEmptyType(store, &b)` at line 56. This helper writes directly to the store via `store.Update(b.ID, beads.UpdateOpts{Type: &t})` (line 227) if the bead's type field is empty. This is a durable write operation.

* **Required Change**: To keep target classification genuinely read-only, either (a) document the exact-ID path as "read-only with bounded repair side-effect" in the classifier contract and atomicity tables, or (b) move `RepairEmptyType` to a pre-classification adapter step that runs before target resolution and marks the bead as repaired in-memory so the classifier itself makes no store writes.

---

### [Blocker] 4. Mail Resolution Chains are Not Fully Separated from Session Resolution

The compatibility chain for "Mail send" (`DESIGN.md:531`) lists:
`human -> live configured-named mailbox basename -> API live target without materialization -> configured named mailbox address -> not found`

However, the source at `handler_mail.go:72-109` shows that "Mail send" uses its own unique lookup flow:
1. `human` passes through (`handler_mail.go:76`).
2. `resolveLiveConfiguredNamedMailTarget` (`handler_mail.go:82-84`) queries `NamedSessionIdentityMetadata` directly via `configuredNamedMailIdentities`, completely bypassing the general session classifier. It enforces its own state filtering (`:269-278`) and ambiguity rules (`:300-303`).
3. Only on a miss does it fall back to the general `resolveSessionTargetIDWithContext` (`handler_mail.go:85-86`).

Similarly, "Mail query" (`DESIGN.md:532`) bypasses the classifier in step 2 by querying identity metadata directly via `mailRecipientsForNamedSession` (`handler_mail.go:119-125`).

* **Required Change**: Add distinct, precise rows for mail send and mail query in the compatibility chains table. Document that `resolveLiveConfiguredNamedMailTarget` and `mailRecipientsForNamedSession` use `NamedSessionIdentityMetadata` directly instead of the general session resolver, and clearly state whether these mail-specific taxonomy lookups are in-scope for Slice 1 centralization or are excluded as mail-specific helpers.

---

### [Blocker] 5. Assignee Normalization Compatibility Chain is Too Vague

The compatibility chain for "Assignee normalization" (`DESIGN.md:536`) states: "bead ID/session-name/alias/configured identity resolver used by the current call site — must preserve ambiguity and duplicate-key handling before centralization." This is too vague to be an implementer's gate.

The source at `handler_beads.go:63-88` shows `normalizeRawBeadAssignee` executing a two-call retry with a materialization fallback:
1. It tries `resolveSessionTargetIDWithContext` without materialization (`:72`).
2. On `ErrSessionNotFound`, it retries `resolveSessionTargetIDWithContext` with `{materialize: true}` (`:74`).
3. It then validates that the resolved bead is open (`:82-87`).

* **Required Change**: Expand the assignee normalization row in the compatibility chains table to clearly detail this two-call materialization retry, the open-bead validation check, and the `RepairEmptyType` call. Specify whether the classifier should return a `configured-named-reserved` kind and let the assignee adapter handle the retry, or if the classifier natively supports a retry-materialization policy.

---

## Missing Evidence & Major Risks

- **Extmsg Chain Specificity**: The extmsg compatibility chains (`DESIGN.md:533-534`) lack specific function names and fallback behaviors. The design must cite specific source functions and document whether extmsg materialization failures abort the entire notification or are gracefully logged.
- **Diagnostic `candidates[]` Structure**: The result contract (`DESIGN.md:544-552`) does not specify what the `candidates[]` array contains. It must explicitly state whether the list includes all evaluated candidates (including demoted or non-matching ones) or only those that matched, and whether dual alias/session_name demotion is exposed as a diagnostic candidate fact.
- **De-scoping Future Extensions**: The Candidate Kinds table includes `historical-alias` (`DESIGN.md:517`) and `ordinary-config-target` (`DESIGN.md:519`), but no current production surface resolves using these kinds alone. They should be marked as "future extensions, not implemented in Slice 1" to prevent scope-creep.
- **Surface Adoption Sequence and Revert Criteria**: The coexistence table (`DESIGN.md:764`) lists "API target resolver adapter first; then mail/extmsg/CLI/assignee helpers one surface at a time" but fails to specify the adoption sequence order, the details of the oracle-to-delegate migration period, or the exact criteria for triggering a revert.

---

## Questions

1. When the classifier returns `direct-session-id` for a closed bead, should the classifier itself enforce closed rejection, or should the calling operation check the status independently? If the classifier enforces it, it violates the separation of classification and policy. If the caller checks, the caller is partially re-derived status.
2. Should `filterOutAliasMatches` dual demotion be exposed as a `demoted_alias` diagnostic candidate, or hidden as an internal implementation detail of the classifier? Exposing it provides stronger traceability for parity tests.
3. Is `resolveLiveConfiguredNamedMailTarget` in scope for the unified target classifier, or should it remain a mail-specific lookup? It queries `NamedSessionIdentityMetadata` directly and returns display addresses rather than raw session IDs.
