# Hugo Bautista - Claude

**Verdict:** block

> Lane note (verify-don't-copy + path handling).
> 1. *Re-grounding.* A re-grounded prior review of this same 16KB doc exists at
>    `attempt-1/reviews/` (mis-placed there by the `gc.attempt=1` path bug). The
>    current `requirements.md` is byte-identical (md5 `4f1ef50…`) to
>    `attempt-2/` and `attempt-3/` `design-before.md`, and `attempt-2` produced
>    no edit (it blocked on a workflow defect), so the prior round's
>    asset-classification findings persist *unaddressed* in this artifact. I did
>    not inherit conclusions: I re-read AC6 and the Example Mapping and verified
>    the split/duplicate hazards against the live source tree.
> 2. *Output path.* Written to `attempt-3/` (the live `iteration.3` dir; its
>    `reviews/` was empty), not the literal `attempt-${gc.attempt}=attempt-1`,
>    which `attempt-2/synthesis.md` documents as the known defect that blocks
>    synthesis and would overwrite a historical review set. Flagging for the
>    operator.
> 3. *Schema.* Removing the file-by-file table into an external ledger (AC6) is
>    schema-*correct* (`gc.mayor.requirements.v1` forbids implementation file
>    assignments in requirements). My block is about the ledger *contract's*
>    completeness, not its absence from the body.

**Top strengths:**
- **The file-by-file map was correctly extracted into an external validated ledger contract (AC6) while keeping file-by-file ownership a hard prerequisite.** This resolves the schema conflict and preserves my lane's guarantee as a gate: the doc honestly stays `status: questions` until the ledger validates, rather than claiming readiness.
- **AC6 covers most of the right per-row dimensions and bans wholesale moves at the contract level.** It records provenance, target owner, target output/retirement, split boundary, fallback classification, rationale, and proof command, and *fails on unrepresented active source files* — so a large bucket cannot be moved without behavior-based per-file rows (my first red flag is addressed in principle).
- **Review-marked resolution is enforced as a gate, not a hope (lane Q3, fully satisfied).** AC6 *fails on unresolved `review` rows* and the ledger must exist *before implementation approval*, so no downstream implementation can depend on an unresolved classification. This is the strongest part of the contract.

**Critical risks:**
- **[Major] AC6's field schema cannot mechanically enforce its own "fails on orphaned split behavior" rule for the split assets that actually exist.** A row records a *single* "target output path or retirement action" plus a prose "split boundary." But the current tree has concrete same-named, dual-pack assets that must split: `following-mol.template.md`, `architecture.template.md`, and `propulsion.template.md` each exist in **both** `examples/gastown/packs/gastown/template-fragments` and `examples/gastown/packs/maintenance/template-fragments` (two of these — `following-mol`, `architecture` — are named in this lane's mandate). A single-output row physically cannot record the Core output *and* the Gastown output *and* which content lands on each side, so the validator has no structured basis to detect an orphaned half. The anti-orphan validation rule is asserted but unbacked by the data model — the precise failure this lane exists to prevent.
- **[Major] Ledger validation is one-directional; phantom/stale rows are not explicitly failed.** AC6 fails on "missing current paths" and "unrepresented active source files" (forward direction: every active file has a row). It does **not** clearly fail on the reverse — a row whose `current path` no longer resolves to a tracked file at a named snapshot. "missing current paths" reads most naturally as "blank field," not "path that no longer exists." With 107 tracked files moving across three roots (`core` 28, `gastown` 47, `maintenance` 32) and the prior 392-line draft having carried 9 verified phantom rows, an external ledger can reproduce that exact defect and drive implementation off stale rows.
- **[Minor] No closed classification vocabulary, so ambiguous owners can be invented.** AC6 implies values via separate fields but never pins an allowed set (e.g. `core` / `core-renamed` / `gastown` / `split` / `retire` / `review`). It cannot distinguish a plain Core keep from a "keep in Core but strip role references" (`core-renamed`) decision — which is a live question for the role-named Core formulas `internal/bootstrap/packs/core/formulas/mol-polecat-base.toml` and `mol-polecat-commit.toml`. Without closed semantics, a conditional decision can be silently recorded as definitive.

**Missing evidence:**
- The validated asset migration ledger does not yet exist (acceptable for `status: questions`, but it means file-by-file ownership can only be assessed as a *contract*, not as content).
- The single git snapshot (commit/tree) the ledger validates against — AC6 requires per-row "provenance" but no one named basis, which is what makes a "phantom row" distinguishable from an "intended-but-not-yet-created target."
- The canonical-source / merge policy for the concrete divergent duplicates that exist today (`following-mol.template.md`, `architecture.template.md`, `propulsion.template.md` present in both the gastown and maintenance packs).
- The classification decision for role-named Core assets (`mol-polecat-base.toml`, `mol-polecat-commit.toml`): `core-renamed` (strip and keep) vs `gastown` (move).

**Required changes:**
- Amend AC6 so each `split` row (or equivalent child-output records) names **both** the Core output path and the Gastown output path (or retirement), plus the content/behavior assigned to each, and tie the "fails on orphaned split behavior" rule to that dual-output schema so it is mechanically enforceable.
- Require **bidirectional** ledger validation against one named snapshot: every active tracked source file under the legacy roots maps to exactly one row (present), AND every row's `current path` must resolve to an existing tracked file at that snapshot (add); failure on stale/phantom rows must be explicit. Name `git ls-files` under the three roots (or equivalent) as the authoritative inventory.
- Enumerate a closed classification vocabulary in AC6 (`core`, `core-renamed`, `gastown`, `split`, `retire`, `review`) with when each applies, distinguishing a plain Core keep from `core-renamed`.
- Add one Example Mapping row exercising a real split: a same-named divergent fragment (e.g. `following-mol.template.md` in both gastown and maintenance) or a role-named Core formula (`mol-polecat-base.toml`), showing its Core output, Gastown output/retirement, and proof command.

**Questions:**
- For `following-mol.template.md` / `architecture.template.md` / `propulsion.template.md` (same basename in both the gastown and maintenance packs today), does the ledger merge them, keep separate Core/Gastown outputs, or retire one — and which copy is canonical?
- For role-named Core formulas (`mol-polecat-base.toml`, `mol-polecat-commit.toml`), is the classification `core-renamed` (strip role references, keep neutral in Core) or `gastown` (move)? AC6's `fallback classification` needs a defined resolution rule.
- Who generates and validates the ledger, and against which commit/tree snapshot? (Open Question 1 raises owner but not the snapshot basis the row-level validation depends on.)
