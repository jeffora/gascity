# Hugo Bautista

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, Gemini (DeepSeek V4 Flash perspective)

**Consensus findings:**
- [Info] The requirements use the right structural approach for asset ownership: file-by-file classification is moved out of the requirements body and into a validated asset migration ledger, preserving schema compliance while keeping ownership completeness as an implementation-approval gate.
- [Info] AC6 is a strong ownership-completeness contract. It requires deterministic source snapshots, stable asset and behavior IDs, current paths, provenance, target owners, output paths or retirement actions, split boundaries, rationales, proof commands, and fail-closed checks for unrepresented source files, unresolved `review` rows, phantom rows, stale snapshots, basename collisions, and orphaned split behavior.
- [Info] Review-marked assets are correctly gated. AC6 fails unresolved `review` rows, and downstream implementation is blocked until AC6 and AC7 exist and pass.
- [Major] AC6 still needs a normative, testable classification rule, not only ledger fields and rationale text. The rule should bind the W6H intent to the ledger: role-neutral SDK infrastructure and mechanisms stay Core, role-specific definitions and behavior move to Gastown, mixed assets split by behavior, and retired assets require explicit alternatives.
- [Major] The mechanism-versus-instance split must be stated explicitly. Formulas, molecules, orders, dispatch/sling, and `following-mol`-style assets can contain both Core mechanism and Gastown-specific workflow content; reviewers need a deterministic rule for assigning each half.
- [Major] Generic-maintenance assets need a product-level destination policy. AC9 addresses maintenance executors and work, but the retired Maintenance pack's prompts, formulas, docs, helper flows, diagnostics, and recovery content still need a clear Core-generalized, Gastown-specific, split, or retired outcome.
- [Major] The ledger or implementation-plan schema must be able to represent multi-output split assets. A single `new owner` / `new path` shape cannot prove both a Core-neutral successor and a public Gastown successor unless the design defines linked per-behavior rows and validates that all halves exist.
- [Minor] Active source files must be interpreted broadly. Nested scripts, doctor checks, `assets/scripts/checks`, template and prompt fragments, namepools, pack manifests, embed files, command docs, formulas, orders, provider overlays, static docs, and executable bits all need ledger coverage when active.
- [Minor] Some files need multiple behavior rows rather than one coarse file row. Architecture fragments, dispatch skills, maintenance docs, status-line scripts, and large patrol formulas may mix routing, diagnostics, notifications, recovery, and role-specific behavior.
- [Minor] The closed classification vocabulary and `fallback classification` semantics are still underdefined. The allowed owner/classification labels and proof obligations need to be enumerated.

**Disagreements:**
- Claude and Codex choose `approve-with-risks`; Gemini's DeepSeek-perspective review chooses `approve`. Assessment: `approve-with-risks` is the right persona verdict because the requirements direction is sound, but several classification and split-schema gaps remain material before implementation can safely proceed.
- Claude treats the missing testable classification rule as a requirements-level major risk. Codex sees no requirements-level blocker and frames its findings as implementation-plan validation work. Gemini says the requirements are acceptable but flags major implementation-plan schema and rollout risks. Assessment: the classification rule and generic-maintenance placement should be tightened in the requirements or AC contract; the split row shape, collision checks, and multi-repo publish gate can be finalized in the implementation plan before implementation approval.
- Gemini recommends dual destination fields for split rows, while Codex leaves the schema shape open. Assessment: dual fields are one valid design, but the required property is stronger than the field names: every split or core-renamed row must prove all Core, Gastown, and retired outcomes with stable identities and validation evidence.
- Claude notes AC6 already names basename collisions, while Gemini emphasizes collision scanning for static/template assets. Assessment: keep the basename collision gate, but add explicit reconciliation for same-concept divergent static or template assets where basename checks alone do not prove behavior preservation.
- Claude recommends adding representative hard split cases to requirements; Codex is satisfied if the validator covers them. Assessment: include a small number of product-level examples for high-risk categories without reintroducing a file-by-file migration table into requirements.

**Missing evidence:**
- The actual AC6 asset migration ledger and validator output.
- The exact deterministic source snapshot command and proof that it covers nested docs, scripts, fragments, manifests, embed files, provider overlays, and static assets.
- A finalized ledger or manifest schema that supports split and core-renamed assets with stable behavior/sub-asset IDs and multiple outputs or retirement actions.
- Product placement policy for generic-maintenance prompts, formulas, docs, helper flows, diagnostics, and recovery content.
- Explicit mechanism-versus-instance placement rule for formulas, molecules, orders, dispatch/sling, and `following-mol`-style assets.
- Enumerated classification vocabulary, `fallback classification` semantics, and `review` row resolution authority.
- Reconciliation policy for same-named or same-concept static/template assets across legacy roots.
- Clean-halves publish and verification evidence for split assets spanning Gas City and the public Gastown pack repository.

**Required changes:**
- Extend AC6, or add a companion AC, to require the ledger to apply an explicit classification rule: role-neutral SDK infrastructure/mechanisms -> Core; role-specific roles, workflows, prompts, commands, overlays, and recovery behavior -> Gastown; mixed assets -> split with per-behavior assignment; retired assets -> explicit retirement action and alternative.
- State the mechanism-versus-instance rule for formulas, molecules, orders, dispatch/sling, and `following-mol`-style assets so reviewers can verify the classification rather than trust row rationale.
- Define the product destination policy for the generic-maintenance asset corpus after the Maintenance pack is retired.
- Enumerate the closed classification vocabulary and define `fallback classification` and `review` resolution semantics.
- Finalize the ledger or manifest schema so split and core-renamed rows carry stable behavior/sub-asset IDs plus every Core, Gastown, and retired output or retirement action; validators must reject coarse rows that cannot prove all halves.
- Require the source snapshot and ledger validator to cover nested files, docs, fragments, scripts, executable bits, pack manifests, generated/embed references, provider overlays, and static assets.
- Add explicit reconciliation for basename collisions and duplicate conceptual static/template assets, including equivalence, divergence, split, canonical-source, merge, or retirement decisions.
- Add a clean-halves multi-repo publish and verification gate so split behavior is not stripped from Gas City before its Gastown successor is published and version-resolvable.
- Add at least one representative hard split example covering behavior, Core output, Gastown output, any retirement action, and proof evidence.
