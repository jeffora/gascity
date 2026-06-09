# Tomas Park - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The rollout is now sliced into green intermediate states instead of a single risky migration. The plan separates public Gastown preparation, public-pin adoption, Core extraction, Core loading/doctor, Maintenance folding, registry/cache cleanup, and final source/docs deletion.
- The gates are behavior-oriented. The design repeatedly rejects path/count-only assertions and requires formula composition, prompt rendering, script execution, runtime-state mutation, hook installation/execution, doctor goldens, duplicate-active matrices, old/new binary checks, and packcompat against the exact public pin.
- `slice-gates.generated.yaml`, `test-migration.generated.yaml`, `behavior-manifest.generated.yaml`, and `public-gastown-pins.yaml` give implementation beads a generated source of truth for commands, replacement witnesses, pin evidence, and forbidden intermediate states.

**Critical risks:**
- [Minor] Several generated artifacts still use placeholder command wording. Lines 2249-2259 say `slice-gates.generated.yaml` is the binding source and each row names exact commands, but lines 2586-2595 still describe required command/test as "generator command", "scanner test", "gate validator", and "pin validator" without naming the entrypoints. The design is good enough directionally, but the implementation beads need concrete commands before they can execute this safely.
- [Minor] The external `gascity-packs` gates are behavior-rich but not as executable as the Gas City gates. Lines 3377-3388 require the `gascity-packs` test suite, manifest completeness, formula/order composition, prompt-template resolution, script execution, retired-path scans, and old-test mapping, but they do not name the exact package commands or pack-local gate script.
- [Minor] The broad "focused suites first" section is slightly less authoritative than the generated gate artifact. Lines 3353-3367 list useful local commands, but implementers should treat `slice-gates.generated.yaml` as binding whenever it names broader process/integration shards for Core loading, Maintenance folding, or registry/cache cleanup.

**Missing evidence:**
- Concrete generator/validator entrypoints for `slice-gates.generated.yaml`, `behavior-manifest.generated.yaml`, `test-migration.generated.yaml`, `public-gastown-pins.yaml`, and the scanner artifacts.
- Exact `gascity-packs` commands for the candidate public Gastown slice, including formula/order composition, prompt-template resolution, script execution, and retired Maintenance path scans.
- A sample `slice-gates.generated.yaml` row showing one slice with command, package, `-run` filter, artifact inputs, public-pin phase, forbidden states, and required sharded targets.

**Required changes:**
- Add the concrete command/test names for the generated gate artifacts, either in the design text or by requiring an initial sample row checked into the plan directory before implementation beads begin.
- Make the public Gastown slice gates as executable as the Gas City gates by naming the pack-local commands and expected artifacts.
- State explicitly that prose command lists are advisory when `slice-gates.generated.yaml` contains a stricter command set.

**Questions:**
- Will `slice-gates.generated.yaml` be generated in the Gas City repo only, or mirrored into `gascity-packs` for cross-repo slices?
- What is the fallback if the `gascity-packs` repo cannot provide a single suite command with the same completeness guarantees as the Gas City gates?
