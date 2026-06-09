# Session Refactor Design

| Field | Value |
|---|---|
| Status | Draft |
| Behavior source | `REQUIREMENTS.md` |
| Scope | Small refactors that move session decisions into `internal/session` |
| Archive | Prior dense review history is in `DESIGN_REVIEW_NOTES.md` |

This is not a new architecture program. It is a simple refactor plan for making
session behavior easier to find, test, and change without changing product
behavior.

## Goal

Move session-specific decisions into `internal/session` one operation at a
time.

Today some callers in API, CLI, worker, and reconciler code know too much about
session state, target resolution, wake rules, and lifecycle metadata. The
refactor should make those callers thinner:

```text
caller gathers facts -> internal/session decides -> caller executes or renders
```

The caller may still gather external facts, call runtime providers, write
non-session domain state, or render API/CLI output. The session module should
own the session rule.

## Product Rule

`REQUIREMENTS.md` is the behavior source of truth. A refactor must preserve the
scenario rows it touches.

If current code and `REQUIREMENTS.md` disagree:

- update code when the requirement is right
- update `REQUIREMENTS.md` when current behavior is the intended product rule
- ask for a decision when neither is clear

## First Refactor

Start with session target resolution/classification.

Why this first:

- it is session-specific
- several surfaces already re-derive it
- it can be tested without provider runtimes
- it should not require a broad new facade

The first cut should preserve the existing resolver behavior:

- direct session bead ID
- open exact `session_name`
- open exact current `alias`
- allow-closed lookup only where the current caller already allows it
- no fallback from ordinary config names or `template:<name>` to live sessions

Do not make one universal target policy. Different callers may still have
different allowed behaviors. `internal/session` should identify what the target
is; the caller or adapter decides what that operation is allowed to do with it.

## Shape

Prefer small operation-specific APIs over a broad `SessionService`.

Good shape:

```text
Target facts -> target classifier -> caller-specific adapter
Wake facts -> wake decider -> wake command
Close facts -> close decider -> close command
```

Avoid:

- one large `SessionFacts` struct
- a generic command bus
- event sourcing as the first step
- moving work, mail, extmsg, provider, or pool policy into `internal/session`

## Boundaries

`internal/session` should own:

- lifecycle projection and transition rules
- session identity and target classification rules
- session-owned lifecycle and wake metadata mutations
- pure decisions that can be unit-tested without stores or providers

Callers should own:

- API and CLI rendering
- work assignment and release policy outside session facts
- mail and external-message delivery policy
- runtime provider execution
- reconciler scaling, budget, progress, and alert policy

## Refactor Rules

For each operation:

1. Pick one current behavior cluster.
2. Read the matching `REQUIREMENTS.md` scenario rows.
3. Add or keep characterization tests for the current caller behavior.
4. Add a small session-owned decider or command.
5. Move one caller to it.
6. Keep the old behavior unless the requirements row changes.
7. Delete or shrink duplicated caller logic after parity is proven.

The test should prove the behavior the user sees, not every internal branch.

## Backlog

1. Target classification: centralize session target identity while preserving
   caller-specific policy.
2. Explicit wake: move wake eligibility/conflict decisions behind a
   session-owned operation.
3. Close and identity retirement: keep close semantics and work-release recovery
   clear without scattering lifecycle metadata writes.
4. Runtime start: fold prepare/commit/rollback metadata ownership into one
   command once the smaller slices have proven the pattern.
5. Reconciler facts: only extract narrow lifecycle eligibility facts; keep pool
   scaling and provider-health policy in the reconciler.

## Non-Goals

- Do not rewrite the reconciler wholesale.
- Do not introduce a large facade before one small operation proves value.
- Do not move work, mail, extmsg, or provider-specific runtime policy into
  `internal/session`.
- Do not require a large preflight artifact system for the first small refactor.
- Do not use design review feedback as a substitute for readable requirements
  and tests.
