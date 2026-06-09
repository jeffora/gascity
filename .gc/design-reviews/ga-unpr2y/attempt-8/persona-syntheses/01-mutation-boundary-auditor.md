# Elena Marchetti

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The proposed scan and inventory scope is not source-complete. Claude found production bead metadata writers outside the stated roots or `internal/session` import heuristic, including `internal/sling`, `internal/convergence`, `internal/molecule`, `internal/nudgequeue`, `internal/sourceworkflow`, and `internal/convoy`. DeepSeek separately found the mandatory scan floor omits `internal/api/huma_handlers_beads.go` and internal session writer files. Codex also flags broad inventory buckets such as W-011, W-026, W-029, W-030, and W-031 as planning placeholders rather than exact proof rows.
- [Blocker] The doctor, repair, and migration exclusion is self-labeling. All reviewers agree production repair/migration utilities must be scanned first and then explicitly allowed by row, evidence, expiry, and owning slice. A class-level path or filename exclusion would let new production code bypass the mutation guard by choosing a repair-looking location or name.
- [Blocker] The static guard feasibility is under-proven for generic store and interface receiver paths. DeepSeek calls the AST/type-narrowing requirement speculative; Claude notes that wrapper and interface receiver detection needs type-checked analysis, not the simple import-string precedent; Codex asks for fixtures proving external patch-map application and extension fail outside approved transition points.
- [Major] The shrink-only allowlist is currently policy text, not an enforceable ratchet. Claude and DeepSeek both point out that a test comparing only current code to the current allowlist cannot detect new rows. Codex also requires an allowlist artifact proving shrink-only behavior, expiry, and migration authority.
- [Major] Generic bridge and dynamic metadata paths need explicit runtime or build-time decisions. Claude highlights `cmd/gc/cmd_bd_store_bridge.go` as an arbitrary bead/key passthrough that cannot be made safe by one permanent static allowlist row. Codex and DeepSeek likewise call for dynamic metadata, generic API/Huma bead update/close routes, and subprocess `bd` mutation verbs to be classified or blocked.
- [Major] Broad W-rows and prose key buckets are not sufficient implementation inputs. Codex requires exact child rows before conversion; DeepSeek requires machine-readable owned-key entries instead of prose buckets; Claude requires the scan floor to be mechanically derived from every W-row path and every scanned writer path.
- [Minor] Patch-builder retirement should end in compiler enforcement. Claude notes that once external callers convert, exported patch builders such as wake/start/close/retire patch helpers should be unexported as part of retirement, not left indefinitely protected only by a guard.

**Disagreements:**
- Codex returns `approve-with-risks`, while Claude and DeepSeek return `block`. My assessment: the practical lane verdict is `block` for the current design because the blockers are about the validity of Slice 0's proof surface, not later implementation polish. Codex's approval is explicitly limited to implementing Slice 0 only and does not approve routing mutation-owning beads.
- DeepSeek treats AST/interface receiver feasibility as a blocker requiring a prototype. Claude frames the same gap as missing mechanics and a load-bearing unknown, while Codex is more optimistic about the guard contract. My assessment: require at least a skeletal type-aware guard or equivalent fail-closed strategy in Slice 0 before the design can claim enforceability.
- Claude emphasizes unscanned internal packages such as sling, molecule, nudgequeue, and convoy; DeepSeek emphasizes Huma generic bead handlers and internal session files. These are complementary, not conflicting. The scan scope should be derived from all production code that imports or receives `internal/beads` store handles, plus exact inventory path coverage, rather than a hand-maintained floor.
- DeepSeek proposes git-aware comparison against `origin/main` for shrink-only enforcement, while Claude proposes a frozen baseline row-ID artifact. Either mechanism is acceptable if CI can prove that new allowlist rows fail without an explicit reviewed baseline update.

**Missing evidence:**
- A generated, source-complete writer inventory and command transcript proving every production `SetMetadata*`, `Update`, `Close`, `Create`, `RepairEmptyType`, `WakeSession`, patch-map application, package mutator, manager lifecycle method, generic API/Huma bridge, and subprocess `bd` mutation path has been classified.
- A concrete guard implementation approach showing whether analysis is type-checked, SSA/dataflow-backed, syntactic fail-closed, or deliberately broader with explicit non-session discrimination rows.
- Store topology proof explaining which bead stores can carry session beads, and whether generic store handles in sling, convergence, molecule, nudgequeue, sourceworkflow, convoy, API, and CLI bridge code can reach session beads.
- A machine-readable owned-key taxonomy for all session-owned and intentionally non-session-owned metadata keys written on session beads.
- A shrink-only enforcement artifact or CI check proving additions to the allowlist fail unless the baseline is explicitly changed, reviewed, and tied to an expiring row.
- Exact-path treatment for production doctor, repair, and migration writers, plus negative fixtures proving self-labeled repair-looking files do not escape the guard.

**Required changes:**
- Redefine Slice 0 scan scope to cover every production package under `cmd/` and `internal/` that imports `internal/beads` or receives a bead store handle, with only mechanically-defined exclusions for tests, testdata, and generated code.
- Regenerate the inventory under that scope and split broad W-rows into exact child rows with path, symbol, key family, target-bead proof or non-session discrimination, guard rule, owner slice, and retirement condition.
- Include the missing writer surfaces called out by the reviews, at minimum `internal/sling`, `internal/convergence`, `internal/molecule`, `internal/nudgequeue`, `internal/sourceworkflow`, `internal/convoy`, `internal/api/huma_handlers_beads.go`, the listed internal session writer files, and the cmd files omitted from the floor.
- Remove blanket doctor, repair, and migration exclusions. Scan those production files and allow intentional writes only through audited session-internal helpers or exact expiring allowlist rows.
- Mechanize shrink-only behavior with a frozen baseline or git-aware CI check, and fail on allowlist growth, expired rows, dangling rows, or rows that no longer match current source.
- Decide the generic bridge/API/Huma/subprocess mutation policy in Slice 0: reject session-owned keys on session-bead targets, build-tag unsafe utilities out of production, or create expiring exception rows with tests and retirement owners.
- Add a guard feasibility proof in Slice 0 that demonstrates how generic stores, wrapper receivers, interface values, dynamic metadata batches, patch-map application, and external patch-map extension are detected or conservatively blocked.
- Convert prose owned-key buckets into exact keys or regex/prefix entries, and add explicit classification for metadata written on session beads but intentionally owned outside the session lifecycle boundary.
- Add unexport-as-retirement language for exported patch builders after their last external caller converts.
