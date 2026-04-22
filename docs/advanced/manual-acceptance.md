---
title: Manual Acceptance
---

# Manual Acceptance

Not every task can be fully automated. When `verify` cannot prove correctness through build/test/lint alone, it routes the verdict to `pending_acceptance` and stops for human review.

## When It Happens

`verify` enters `pending_acceptance` when:

- The brief has no `Verification` section with explicit commands, and auto-detection finds nothing
- Acceptance criteria exist that have no automated coverage (e.g., "UI feels responsive")
- The operator explicitly wants human sign-off

## Flow

```
verify ──▶ pending_acceptance ──▶ /checkup accept issue-N ──▶ finalize ──▶ done
```

## Recording Acceptance

```
/checkup accept issue-8
```

`accept` performs the following:

1. Confirm `Workflow Entry State: pending_acceptance`
2. **Publish any local-only code first** — stage, commit, and push if there are uncommitted changes
3. Render an `Acceptance Summary` with reviewer name, timestamp, code ref, and notes
4. Update brief metadata:
   - `Current Stage: checkup`
   - `Next Stage: done`
   - `Pass/Fail Outcome: pass`
   - `Completion Basis: accepted`
5. Emit `checkup_accept_recorded` event (silent)
6. Remove the `pending-acceptance` label
7. Proceed to **Finalize**

## Optional Note

Since v0.6.3, you can append free-form text after the issue ref:

```
/mino-checkup accept #1842 verified on staging with 1000 concurrent users
```

This note is captured verbatim in the brief's `Manual Acceptance` section under `### Accept Note` and used by the agent as primary decision input for reply dispatch.

## Pending Acceptance List

`/checkup reconcile` emits a `Pending Acceptance` subsection listing every task still awaiting human review, with the next manual action required. This gives you a centralized view of all blocked work.
