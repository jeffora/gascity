# Leah Okafor — Gemini (Doctor Mutation Safety Reviewer, Attempt 2, Independent Review)

**Verdict:** approve

**Lane:** doctor fix coordinator atomicity, byte-preserving TOML, concurrency with live controllers, advisory locks, idempotent recovery.

Reviewed against the Iteration 2 / Attempt 2 draft of `plans/core-gastown-pack-migration/implementation-plan.md` and `plans/core-gastown-pack-migration/requirements.md` in the active repository workspace.

---

## Executive Summary

As Leah Okafor, the **Doctor Mutation Safety Reviewer**, I have conducted an independent, evidence-backed safety and risk audit of the Iteration 2 (Attempt 2) design for the Core and Gastown Pack Split. My verdict is **Verdict: approve**.

In this iteration, the design team has directly resolved all of the critical transaction safety, TOCTOU race conditions, and POSIX failure-atomicity gaps identified during the Iteration 1 review. The introduction of the `FixIntent` mutation coordinator, a strict lock-before-preflight protocol, and a durable write-ahead journal/recovery ledger transforms the doctor's mutation path from a series of unsafe sequential file operations into a robust, crash-resilient convergence transaction.

The proposed design is outstanding, first-principles systems engineering and is fully ready for safe implementation.

---

## Top Strengths

- **Enforceable Fix Coordination (§246–251)**: Replacing legacy `Check.Fix(ctx)` direct writes with a unified `FixIntent` plus mutation coordinator API guarantees that all modifications to city manifests, lockfiles, installed pack directories, and runtime-state are mediated by a single, secure boundary.
- **Strict Lock-Before-Preflight Protocol (§253–257)**: The coordinator now acquires the city's directory advisory lock *before* performing preflight checks and digest validation. If a reporting phase reads data before lock acquisition, the digests are re-validated post-lock, closing any TOCTOU concurrency windows.
- **Live Controller Detection (§256–257)**: The doctor now discovers running controllers via live runtime state instead of stale-prone status files, refusing automatic fixes if a controller is active.
- **Durable Write-Ahead Journal & Safe Recovery (§259–264, §372–376)**: Multi-file fixes must write a durable recovery state (mutation journal) documenting staged paths, preflight digests, publish order, and rollback instructions before any publication occurs. This ensures that any process death before commit can be rolled back or rolled forward deterministically on subsequent runs.
- **CST-Based TOML Comment and Format Preservation (§432–434)**: The plan explicitly specifies that scoped TOML edits must preserve comments, unknown tables, unknown fields, array order, and formatting, or else the automatic fixer refuses and guides the operator manually.
- **Version-Skew Safety via Migration Markers (§269–273)**: Recording a detailed migration marker that detects old-binary post-marker writes and requiring manual reconciliation prevents silent data corruption under version-skewed operation.

---

## Technical Audit & Verification of Changes

Compared to the Iteration 1 draft, the design now has concrete answers and explicit specifications for every critical risk:

| Legacy Risk (Iteration 1) | Iteration 2 Resolution | Evidence in Plan |
| --- | --- | --- |
| POSIX Sequential Rename Crash Hazard | Durable Recovery State written before publish, specifying a single commit point and a deterministic rollback/re-run loop on startup. | §259–262, §372–376 |
| Concurrency Window with Controller | Coordinator discovers live running controllers directly from runtime state and refuses automatic fix; holding directory advisory lock blocks mutations. | §253–257 |
| Unenforced TOML formatting/comments | The automatic fix refuses if scoped TOML edits cannot preserve comments, unknown tables/fields, or array order. | §432–434 |
| Weak Provenance Verification | `internal/packsource` classifier returns typed states (retired custom/fork vs retired generated/example) to prevent overwriting user edits. | §222–228, §238–242 |

---

## Operational Notes & Continuous Monitoring

While the design is fully approved, the following operational nuances should be verified in code reviews and test assertions:

1. **Advisory Lock Leases on Networked Filesystems (NFS/Gluster)**: If the city directory is located on a networked mount, verify how the OS directory advisory lock behaves on connection loss. In extremely rare cases of lock lease drops, the coordinator should ensure that file descriptor validation remains active before executing renames.
2. **Quarantine of Unclassifiable Legacy States**: Stale system packs are ignored by active discovery and reported as legacy state rather than deleted. Maintainers should ensure that the doctor's manual cleanup messaging provides precise, non-interactive copy-paste terminal commands to minimize operator overhead.

---

## Questions

### Detailed Responses to Lane-Specific Questions

#### Q1: Do all doctor --fix paths stage FixIntent objects before mutation and hold a city advisory lock across stage, validate, compare-before-rename, and publish?
**Answer**: Yes. The design mandates that direct `Check.Fix(ctx)` mutations are replaced with `FixIntent` staging. The mutation coordinator is the sole path allowed to write, and it acquires a crash-released city advisory lock *before* preflight digest checks. By holding this lock continuously across staging, target-digest verification, and temp-file renaming, the coordinator successfully isolates the mutation transaction from concurrent processes.

#### Q2: When scoped TOML edits cannot preserve comments or unknown fields byte-for-byte outside intended changes, does the fixer refuse rather than rewrite whole files?
**Answer**: Yes. Under the revised testing and validation criteria, scoped TOML edits must preserve comments, unknown tables, unknown fields, array order, formatting, and unrelated lock entries. If this byte-preserving round-trip cannot be guaranteed, the automatic fix is refused, and the operator is provided with manual steps, preventing the risk of standard TOML encoders stripping custom configurations.

#### Q3: What recovery is specified for crashes or concurrent old and new binaries between per-file renames so cities cannot remain half-migrated?
**Answer**: The coordinator writes a durable recovery state record containing a mutation intent id, locked city, preflight file digests, staged paths, publish order, commit point, completed steps, rollback instructions, and final validation result before executing any publish step. If a crash occurs before the commit point, the recovery loop reruns deterministically or rolls back based on these instructions. If a crash occurs after the commit point, the system converges by re-validating required Core participation and marker state. Concurrent old binaries are handled via post-marker write detection, raising version-skew diagnostics rather than silently corrupting data.
