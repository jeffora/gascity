# Oleg Marchetti - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The plan explicitly blocks Gas City source deletion, Core role-generalization, Maintenance removal, and public synthetic alias retirement until immutable public Gastown commits and proof artifacts exist (`design-before.md:105`-`design-before.md:128`).
- The Behavior Evidence manifest has the right traceability shape: stable row id, old owner/path/kind, trigger/requester/detector/mail/nudge/script/runtime-state details, old and new witnesses, new owner/path, exact public commit, consumed `PublicGastownPackVersion`, and approved semantic delta/removal records (`design-before.md:139`-`design-before.md:153`).
- Packcompat is tied to normal public-pack resolution from an exact pin, not copied assets: it installs public Gastown through the ordinary remote-pack path or validated ordinary cache and checks one assertion per manifest row (`design-before.md:162`-`design-before.md:166`).

**Critical risks:**
- [Major] The witness contract still allows weak old-side evidence. The manifest row may use "source assertion" as an old witness (`design-before.md:145`), which can prove that code existed but not that the old trigger, requester/detector path, notification, mail/nudge, prompt, formula, order, or script branch actually fired. For behavior-preservation rows, the plan should require executable, golden, transcript, or fixture-backed old evidence unless the row is explicitly classified as an approved delta/removal with owner and operator impact.
- [Major] CI failure conditions say rows must have old and new witnesses, but they do not state that the witness type must match the behavior kind. A prompt fragment, order trigger, notification target, and runtime-state migration need different proof shapes. Without a witness-type matrix, a path-level or source-only witness can satisfy the manifest while missing execution-level behavior.
- [Minor] The plan says generator, schemas, and tests move into implementation-owned paths selected by the first task slice (`design-before.md:133`-`design-before.md:137`). That is acceptable for sequencing, but task decomposition should not start source moves before the generator path and manifest schema are concrete enough for beads to enforce.

**Missing evidence:**
- A witness-type matrix mapping each behavior kind to the minimum old and new proof type.
- A concrete example manifest row for one formula/order trigger and one notification or mail/nudge path.
- A CI rule stating that source assertions alone are allowed only for historical inventory rows or approved deltas/removals, not for behavior-preservation equivalence rows.
- The exact checked-in home for the manifest schema and generator before source-moving beads are created.

**Required changes:**
- Strengthen the manifest contract so equivalence rows require behavior-level witnesses, not source-existence witnesses. Keep source assertions only as supplemental evidence or for explicitly approved deltas/removals.
- Add a witness-kind matrix for formulas, orders, scripts, prompts, provider overlays, notifications, requester/detector relationships, mail/nudge targets, and runtime-state helpers.
- Require packcompat to fail when a moved public Gastown row lacks a public-pin execution or composition witness through ordinary remote-pack resolution.
- Before bead decomposition for source moves, name the generator path, schema path, and at least one sample row fixture that downstream tasks must update.

**Questions:**
- For a behavior that never had an old test, will the migration generate an old command transcript/golden before moving it, or will it classify the change as an approved delta/removal?
- Which manifest field distinguishes behavior-equivalence rows from docs-only, inventory-only, or approved-retirement rows?
- Will every public Gastown-owned row be verified from the pinned public checkout in CI, or are any rows allowed to rely on local working-tree evidence?
