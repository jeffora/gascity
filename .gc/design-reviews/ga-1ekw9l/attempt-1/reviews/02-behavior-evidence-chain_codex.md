# 02 Behavior Evidence Chain - Codex Review

Persona: Oleg Marchetti

Verdict: approve for prerequisite-producing decomposition only. Do not approve
Gas City source deletion, Maintenance removal, public activation-pin
consumption, or behavior-changing loader cutover until the AC6, AC7, AC14,
AC15, AC16, and AC17 evidence artifacts exist, validate, and cite immutable
public-pack evidence.

## Findings

### Required clarification: mark the draft maintenance classification as non-gating input

The implementation plan now defines the right generated evidence chain:
`pack-evidence generate` and `validate` create
`support/behavior-preservation-manifest.yaml`; each row carries trigger,
requester, detector, route metadata, mail/nudge target, prompt/script evidence,
old witness, new witness, semantic-equivalence or approved-delta records, and
public pin/digest data; `test/packcompat` then proves the exact public Gastown
pin through normal resolution with in-tree fallback disabled
(`implementation-plan.md:151`, `implementation-plan.md:170`,
`implementation-plan.md:192`, `implementation-plan.md:199`,
`implementation-plan.md:208`).

The remaining ambiguity is that the only current file under
`plans/core-gastown-pack-migration/support/` is
`maintenance-asset-classification.md`, and that file is explicitly `Status:
draft` while still listing unresolved "Working Decisions Needed" for JSONL
archive/export ownership, reaper cleanup ownership, and `mol-dog-*` compatibility
aliases (`maintenance-asset-classification.md:1`,
`maintenance-asset-classification.md:123`). The implementation plan correctly
lists the binding support set separately and does not list this draft as a gate
artifact (`implementation-plan.md:554`). It also correctly blocks
behavior-changing work until AC6/AC7/AC14-AC17 artifacts exist and validate
(`implementation-plan.md:692`).

Make that relationship explicit in the plan: `maintenance-asset-classification.md`
is a seed inventory only. It must not satisfy AC6, AC7, or AC17, unlock any
source move/deletion, or count as ownership/witness evidence. The generated
`asset-migration-ledger.yaml`, `behavior-preservation-manifest.yaml`, public
Gastown manifest/pin ledger, and packcompat transcripts are the first gating
evidence. Without that sentence, an implementer could mistake a draft,
path-level classification for the behavior denominator and move or retire rows
that still lack behavior IDs, call-site granularity, old witnesses, new public
pin witnesses, and explicit policy decisions.

## What Is Solid

- The plan answers the three evidence-chain questions this persona cares about:
  inventory scope includes triggers, requesters, detectors, notifications,
  prompts, orders, mail/nudge targets, scripts, route metadata, runtime state,
  and historical fixtures; moved/split/generalized/deleted rows need both old
  and new executable witnesses; public-pack rows cannot unblock Gas City
  deletion until commit, pack digest, behavior-manifest digest, and packcompat
  transcript are cited.
- The packcompat contract uses the ordinary remote-pack path or validated
  ordinary remote cache, not copied assets, and has separate compatibility-pin
  and activation-pin modes. That addresses the main risk that Gas City could
  consume a local in-tree copy while claiming to validate the public pack.
- The rollout section preserves sequencing discipline: only external
  prerequisites and proof-producing beads may be created before full
  implementation decomposition, and source deletion/Maintenance removal waits on
  the named evidence gates.

## Residual Risk

The evidence system is intentionally large. The first implementation beads must
land generator tests and negative controls early enough that "manifest exists"
cannot become a checkbox. The validator should fail on omitted behavior-bearing
files, path-only rows, existence-only witnesses, skipped/no-op tests, rows whose
new public witness is unavailable, and any Gas City deletion attempt that lacks
the exact public commit/digest/transcript chain.
