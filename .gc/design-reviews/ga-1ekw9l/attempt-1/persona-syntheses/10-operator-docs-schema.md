# Claire Dubois

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The operator onboarding and tutorial story is not safe for the role-neutral default. Claude verified that `docs/tutorials/01-cities-and-rigs.md` still teaches the minimal default as having an always-on mayor, and the surrounding tutorial sequence continues Gastown-role framing outside the stated update scope. Codex and DeepSeek also require transcript/golden proof for the default and Gastown-template paths.
- [Blocker] The docs and generated-surface inventory is too narrow. The reviewers converge that a hand-picked list cannot satisfy AC12; the plan needs an audit-derived inventory covering tutorials, getting-started and migration guides, troubleshooting, generated docs/schema/reference artifacts, CLI help, doctor/import-state text, examples, prompts, scripts, and public Gastown docs.
- [Blocker] The support artifacts are not yet executable contracts. `terminology-matrix.yaml`, `docs-authority-audit.yaml`, generated-reference outputs, and `migration-diagnostics.schema.json` need concrete row schemas, fixture paths, validators, generated report paths, CI commands, and positive/negative controls before decomposition can safely split implementation.
- [Major] `docs/reference/system-packs.md` is contradictory. All reviewers note that it does not exist in the repo, while the Current System and docs-update language describe it as an existing legacy document to update. The plan must consistently mark it as a new authoritative page or retarget the authority to an existing file.
- [Major] Generated artifact freshness and wording checks are under-sequenced. OpenAPI, dashboard types, docs/schema files, docs/reference output, CLI help, tutorial transcripts, doctor output goldens, and import-state output must regenerate before wording lint, and those gates must run after each operator-facing behavior slice rather than only during final cleanup.
- [Major] The wording matrix needs explicit collision fixtures and allowlist policy. It must accept legitimate lowercase maintenance, Dolt/store-maintenance, `store.maintenance` events/schema terms, and Core maintenance-worker bindings while rejecting retired `packs/maintenance`, in-tree Gastown authority, stale generated examples, and Gastown role names presented as built-in/default Gas City concepts.
- [Major] The operator upgrade and recovery path is scattered. The design needs one canonical guide, transcript, or runbook from stale packs, missing Core, retired imports, moved orders/formulas, public pin/cache failures, version skew, or live-controller repair refusal through diagnosis, repair or manual remediation, and re-verification.
- [Major] Doctor/import-state machine contracts are underspecified. Stable condition codes and text/JSON goldens are needed for `gc doctor`, `gc doctor --json`, `gc doctor --fix --non-interactive`, and `gc import-state`, tied to `migration-diagnostics.schema.json`.
- [Major] Headless repair and rollback safety remain incomplete. DeepSeek specifically identifies missing `--non-interactive` semantics for mutating doctor repair and an undefined deterministic re-upgrade flow or one-way-boundary warning for old-binary writes after the migration marker.
- [Minor] Support-artifact ownership and safety checks need cleanup. `terminology-matrix.yaml` is called the vocabulary authority but is absent from the required support set; the asset-migration ledger path and validation hook should be named; tutorial 05/07 transcript disposition and an automated guard against default tmux-server cleanup should be required.

**Disagreements:**
- Claude blocks, while Codex and DeepSeek return `approve-with-risks`. My assessment: the persona verdict is `block` because Claude's tutorial finding is current-tree grounded, and the other reviews still require gate-critical docs/schema contracts before implementation can proceed safely.
- DeepSeek is more optimistic that the current plan names the correct tutorials and mitigates stale docs. Claude and Codex are more persuasive here because they identify concrete stale surfaces and the plan still lacks an audit artifact that would prevent omissions during decomposition.
- Codex treats executable support-artifact policy as the highest risk; Claude spreads the same issue across docs inventory, CI sequencing, matrix controls, and import-state goldens. My assessment: keep this as blocking for this lane because docs/schema correctness depends on those artifacts being precise before tasks are created.
- DeepSeek trusts the terminology-matrix classifier concept more than the others. My assessment: the classifier approach is acceptable only after real positive/negative fixtures, severity, expiry, and suppression rules are specified.
- DeepSeek uniquely emphasizes non-interactive repair, post-marker old-binary writes, ledger path, and tmux isolation. My assessment: retain these as required changes because they affect headless operator recovery, state-safety messaging, and local/CI safety.
- Codex and DeepSeek treat `docs/reference/system-packs.md` as readily fixable by creating the page. My assessment: it remains a major decomposition hazard until all Current System and docs-scope language consistently says create-not-update or names an existing page.

**Missing evidence:**
- Complete audit output or a checked-in inventory for every affected tutorial, operator doc, generated reference, schema, CLI help surface, script, prompt, and example.
- Golden first-run and tutorial transcripts proving role-neutral minimal/default `gc init` behavior and the separate `gc init --template gastown` path.
- The exact slice that creates and wires `terminology-matrix.yaml`, `docs-authority-audit.yaml`, wording scanner fixtures, allowlist policy, generated reports, and CI before the first operator-facing behavior change.
- The regeneration command list or make target for OpenAPI, dashboard TypeScript, docs/schema, docs/reference, CLI help, tutorial transcripts, doctor output, import-state output, and wording reports.
- The authoritative vocabulary destination for Core, provider host packs, retired Maintenance, public Gastown, and legitimate store/lowercase maintenance terminology.
- Stable `migration-diagnostics.schema.json` condition codes and text/JSON goldens for doctor and import-state flows.
- The exact non-interactive doctor repair semantics for headless use.
- The detection mechanism and recovery policy for old-binary writes after the migration marker.
- The asset-migration ledger path, validation hook, tutorial 05/07 disposition, and tmux isolated-socket cleanup guard.

**Required changes:**
- Replace the hand-picked docs update list with an audit-derived inventory covering every affected getting-started, migration, troubleshooting, tutorial, generated reference, schema, CLI help, script, prompt, and example surface.
- Reframe tutorial 01 so minimal/default `gc init` is role-neutral Core with no mayor/polecat, then require transcript goldens for both minimal and `--template gastown` paths across the affected tutorial sequence.
- Add wording rules and fixtures for "Gastown role name presented as a built-in/default Gas City concept" while allowing legitimate public Gastown-template documentation.
- Resolve `docs/reference/system-packs.md` consistently as a new authoritative page or retarget all references to an existing page.
- Define schemas, fixture locations, generated report paths, validators, allowlist owner/expiry rules, severity, suppression policy, and CI commands for `terminology-matrix.yaml`, `docs-authority-audit.yaml`, generated references, and `migration-diagnostics.schema.json`.
- Pin the docs/schema/wording gates before the first operator-facing behavior slice, and require regeneration plus wording checks after each operator-facing slice.
- Name one canonical upgrade/recovery guide or transcript and bind it to stable condition codes from first diagnostic through repair or refusal and re-verification.
- Add `gc import-state` to the wording/golden surface and bind doctor/import-state JSON and text output to `migration-diagnostics.schema.json`.
- Define the exact `gc doctor --fix --non-interactive` flag behavior for automated, headless environments.
- Specify either a deterministic re-upgrade/merge flow or an explicit one-way-boundary warning for post-marker old-binary writes.
- Add `terminology-matrix.yaml` to the required support set or explain its authority relationship to `docs-authority-audit.yaml`; name the asset-migration ledger path and validation hook.
- Require tutorial 05/07 transcript goldens and an automated tmux cleanup guard that rejects default-server cleanup.
