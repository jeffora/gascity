# Petra Novak - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The requirements name `internal/packs/core` as the end-state Core source authority and require Go embed declarations, builtin registry entries, materialization commands, generated hashes, and tests to close over that authority.
- Maintenance retirement is explicit: it must not be bundled, public-source recognized, auto-included, materialized as an active system pack, selected by lock refresh, or treated as an implicit dependency.
- The source-consumer closure requirement covers the important downstream-reference risk, including retired-Maintenance consumers, Dolt/provider scripts, hooks, fixtures, docs, tests, and the concrete `port_resolve.sh` to `dolt-target.sh` helper shape.

**Critical risks:**
- [Minor] The requirements depend heavily on `source-consumer-closure.yaml` to prevent stale embed/import/materialization paths. The implementation plan must make that validator parse actual Go embed declarations, registry entries, generation commands, and script source calls rather than relying on prose inventory.
- [Minor] The legacy `internal/bootstrap/packs/core` path may be deleted, emptied, or isolated as a fixture/compatibility shim. That flexibility is acceptable for requirements, but the later design must make the selected disposition unambiguous so stale bootstrap paths cannot keep satisfying runtime Core.

**Missing evidence:**
- No open product decision blocks this lane. The missing evidence is expected downstream: source-consumer closure rows, embed/registry/materialization tests, Maintenance-absent representative checks, and active test-execution evidence for replacement coverage.

**Required changes:**
- None before requirements approval.

**Questions:**
- None.
