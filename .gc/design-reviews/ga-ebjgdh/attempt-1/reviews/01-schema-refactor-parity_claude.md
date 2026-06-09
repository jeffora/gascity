# Morgan Schema Refactor Parity Reviewer - Claude

**Verdict:** block

**Why this verdict (lane question 1, the schema gate):** I was handed a
non-empty `implementation-plan.schema.md` (`gc.mayor.implementation-plan.v1`)
and told to judge whether `internal/session/DESIGN.md` conforms to it. It does
not, and the schema's own rule for that case is to
`stop with blocked:wrong-artifact rather than iterating the document`. So the
schema-gate answer is unambiguous: **block as wrong-artifact**. This is a
process/artifact-routing block, *not* a judgment that the engineering content is
bad — as a module refactor design the document is largely sound (see strengths).
The maintainer/synthesis step must first resolve which artifact family this
review is supposed to gate before any of the content findings matter.

**Top strengths:**
- Accurate, code-grounded resolver contract. The "First Refactor" cases (direct
  bead ID -> open exact `session_name` -> open exact `alias`; allow-closed only
  where the caller already allows it; no fallthrough from ordinary config names
  or `template:<name>`) match `internal/session/resolve.go:65-122` and
  `REQUIREMENTS.md` invariants/SESSION-ID-003/004 exactly. The design describes
  real current behavior, not an aspiration.
- Strong anti-premature-abstraction posture. "Shape" and "Non-Goals" explicitly
  reject a broad `SessionService`, one large `SessionFacts` struct, a generic
  command bus, and event-sourcing-first, in favor of operation-specific deciders
  (`target classifier`, `wake decider`, `close decider`). This directly answers
  my red flag on broad facades.
- Layering intent matches the codebase as it stands. "Boundaries"/"Non-Goals"
  keep work assignment, mail, extmsg, provider execution, and pool/reconciler
  policy out of `internal/session` — and today the code honors that:
  `internal/session` imports neither `internal/mail` nor `internal/extmsg`, and
  carries no work-release logic.

**Critical risks:**
- [Blocker] Wrong artifact for the implementation-plan schema gate (lane Q1, red
  flag 1). `internal/session/DESIGN.md` has a markdown status table instead of
  the required YAML front matter (`plan_slug`, `phase: implementation-plan`,
  `rig`, `rig_root`, `artifact_root`, `requirements_file`, `status`,
  timestamps); it uses sections `Goal / Product Rule / First Refactor / Shape /
  Boundaries / Refactor Rules / Backlog / Non-Goals` instead of the mandated
  `Summary / Current System / Proposed Implementation / Data And State /
  Testing / Rollout And Recovery / Open Questions`; and it lives at a module
  path, not `<rig-root>/plans/<slug>/implementation-plan.md`. The schema names
  this exact case ("It is not a module-local reference document"; "turn the
  artifact into a module-local `DESIGN.md`" is forbidden) and prescribes
  `blocked:wrong-artifact`. Note the schema also forbids the remedy of
  retrofitting front matter onto this file — you cannot make `DESIGN.md` conform
  without violating the schema, which is itself the signal that the gate is
  pointed at the wrong file.
- [Major] Target-classification parity is asserted, not traceable (lane Q2/Q3).
  The first slice never cites the `REQUIREMENTS.md` rows it touches
  (SESSION-ID-003, SESSION-ID-004, SESSION-ID-008, SESSION-ID-009, plus the
  Global Invariants on resolution exactness and template/config-name exclusion),
  and never enumerates the caller surfaces it must preserve. There are ~8-11
  surfaces re-deriving resolution today: `cmd/gc/session_resolve.go`,
  `cmd/gc/cmd_sling.go`, `cmd/gc/session_template_start.go`,
  `cmd/gc/worker_handle.go`, `internal/worker/factory.go`,
  `internal/dispatch/control.go`, `internal/mail/beadmail/beadmail.go`,
  `internal/graphroute/graphroute.go`, `internal/api/session_resolution.go`,
  `internal/api/handler_extmsg.go`, `internal/api/handler_status.go`. They are
  *not* uniform: several repeat the `ByExactID -> ResolveSessionID ->
  AllowClosed` ladder, but `internal/api/session_resolution.go:369-470` layers
  an extra orphan-rejection step on top. "Preserved across every caller surface"
  cannot be claimed until each surface's wrapper behavior is named and pinned by
  a characterization test. The design states the right rule ("add or keep
  characterization tests") but identifies zero existing tests
  (`internal/session/resolve_test.go` covers the core resolver; the per-caller
  wrappers, especially the API orphan path, are unaddressed).
- [Major] Close vs work-release ownership is ambiguous and points the wrong way
  (layering lane, red flag 3). Backlog item 3 folds "work-release recovery" into
  close, but today work release on session close (SESSION-WORK-001..004) lives
  in the caller layer (`cmd/gc/session_beads.go`, `session_work_guard.go`,
  `session_reconciler.go`, ...), and "Boundaries" assigns "work assignment and
  release policy" to callers. Moving release into a session-owned close decider
  would pull caller-owned policy into `internal/session` and break the clean
  layering the design otherwise protects. The doc must state explicitly that the
  close decider emits close/lifecycle facts while the caller performs work
  release, or it risks authorizing the exact absorption my red flags forbid.
- [Minor] The anti-`SessionService` stance is prose only (lane Q4). The doc
  gives principles and a backlog ordering but no enforceable checkpoint. This
  module otherwise relies on CI-enforced invariants (e.g. the worker-boundary
  import test, the agent field-sync test); a comparable mechanical guard — or an
  explicit statement that review discipline is the only gate — would make the
  "no universal service" commitment durable.

**Missing evidence:**
- No `REQUIREMENTS.md` row IDs cited for any backlog slice, so parity coverage
  is unverifiable from the document alone.
- No caller-surface inventory; "every caller surface" has no enumerated set to
  check against (the ~8-11 surfaces above are absent from the doc).
- No named characterization tests beyond the implicit core resolver; the
  per-caller wrappers (API orphan rejection, worker factory, graphroute, sling)
  have no stated pinning test.
- No concrete checkpoint preventing the classifier from growing into a universal
  `SessionService`.

**Required changes:**
1. Resolve the artifact-type mismatch before iterating. Decide whether this
   review gates a Mayor implementation plan or a module design doc. If the
   former, author `implementation-plan.md` at `<rig-root>/plans/<slug>/` with the
   required front matter and the seven mandated sections and point the review at
   it — do **not** mutate `internal/session/DESIGN.md` into that shape (the
   schema forbids it). If the latter, the implementation-plan schema must not be
   this bead's conformance gate.
2. For the target-classification slice, name the exact `REQUIREMENTS.md` rows
   (SESSION-ID-003, -004, -008, -009 and the relevant Global Invariants) and the
   caller surfaces listed above, and confirm a characterization test pins each
   surface's current behavior — including the API orphan-rejection wrapper —
   before any caller is moved.
3. Disambiguate close vs work-release ownership in backlog item 3: the
   session-owned close decider returns close/lifecycle facts; work release
   (SESSION-WORK-001..004) stays caller-owned in `cmd/gc`. Make this explicit so
   the slice cannot drift into moving release policy into `internal/session`.
4. State the anti-`SessionService` checkpoint concretely (a test/lint/review gate
   analogous to the existing worker-boundary import guard), or acknowledge that
   none exists and reviewer discipline is the control.

**Questions:**
- Is the intended output of this design-review track a Mayor
  `implementation-plan.md`, or a module-local design doc? This single answer
  decides whether the schema gate applies at all and therefore whether the block
  stands.
- For target classification, does the centralized classifier return only an
  identity/type (callers keep their own allow-closed and orphan-rejection
  policy), or does it absorb allow-closed semantics? The API orphan-rejection
  wrinkle suggests classify-only is the safe contract; please confirm.
- What enforceable checkpoint will catch the session-owned decider growing into
  a universal service or absorbing assignment/mail/provider/pool policy?
