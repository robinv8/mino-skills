# Brief Contract

Defines the local brief format used by the skill set.

## Purpose

Keep `task`, `run`, `verify`, and `checkup` aligned on:

- what a local brief is
- which sections it should contain
- which sections are managed by skills
- which sections may remain human-authored

## Location

Default after publication: `.mino/briefs/issue-<issue-number>.md`

The file path is a locator convenience, not the canonical identity. `Task Key` remains the authoritative local identity.

## Sections

A compliant brief should contain these sections:

- `Issue`
- `Classification`
- `Dependencies`
- `Acceptance Criteria`
- `Verification`
- `Target Files`
- `Work Breakdown`
- `Workflow State`
- `Manual Acceptance`
- `Failure Context`
- `External Event`
- `Completion Handoff`
- `Execution Summary`
- `Verification Report`
- `Verification Summary`
- `Pass/Fail Outcome`
- `Open Questions / Warnings`
- `Source`

`Work Breakdown` is required for composite tasks. `Manual Acceptance`, `Failure Context`, `External Event`, `Execution Summary`, `Verification Report`, `Verification Summary`, and `Pass/Fail Outcome` may be absent until the workflow reaches those phases.

## Required Fields By Section

### Issue

- `Task Key`
- `Issue Number`
- `GitHub`
- `Parent Issue` when applicable

### Classification

- `Type`
- `Shape`
- `Executability`
- `Approval State`

### Dependencies

- `Depends On` as a stable list of task keys

### Workflow State

- `Spec Revision`
- `Approved Revision`
- `Current Stage`
- `Next Stage`
- `Workflow Entry State`
- `Attempt Count`
- `Max Retry Count`
- `Code Publication State`

### Manual Acceptance

Required when `Workflow Entry State: pending_acceptance`:

- `Reason`
- `Checklist`
- `Action`

### External Event

Optional. Populated by `checkup reconcile` when out-of-band activity contradicts workflow state (e.g., the linked issue is closed externally). Required fields when present:

- `Event` (e.g., `issue_closed`)
- `Detected At`
- `Source` (e.g., `github`)
- `Action` (the human follow-up the protocol requires)

### Completion Handoff

- `Completion Basis` when known
- `Code Ref` when code publication is relevant

## Managed Sections

Managed by skills:

- `Issue`, `Classification`, `Dependencies`
- `Acceptance Criteria`, `Verification`, `Target Files`
- `Work Breakdown`, `Workflow State`, `Manual Acceptance`, `Failure Context`, `External Event`
- `Completion Handoff`, `Execution Summary`, `Verification Summary`, `Pass/Fail Outcome`
- `Source`

`Open Questions / Warnings` is human-addable and must be preserved.

## State Storage Rules

Brief state is a local cache for fast DAG scheduling and CLI inspection.

- Skills update brief state during workflow transitions
- State updates are local-only: do NOT stage or commit `.mino/briefs/` files
- The linked GitHub issue body plus structured workflow events remain the authoritative record
- If brief and issue state drift, `checkup reconcile` repairs the brief from issue metadata and the highest valid event sequence for the active approved revision

## Section Intent

| Section | Purpose |
|---------|---------|
| Issue | Source identity, stable URL, task key, and issue locator |
| Classification | Type, decomposition shape, executability, approval |
| Dependencies | Machine-readable DAG edges used by `run` |
| Acceptance Criteria | Observable outcomes to satisfy |
| Verification | Repository-native or manual checks |
| Target Files | Implementation and review scope |
| Work Breakdown | Child tasks and order for composite work |
| Workflow State | Revision, stage, gate state, attempt budget, and publication state |
| Manual Acceptance | Detailed local checklist for pending human verification |
| Failure Context | Exact failure payload for self-correction |
| External Event | Out-of-band activity detected during reconciliation that contradicts workflow state |
| Verification Report | Pointer to locally-authored evidence report and optional promoted doc link |
| Completion Handoff | How the task became completable: verified, accepted, or aggregated |
| Execution Summary | Concrete execution results |
| Verification Summary | Concrete verification or acceptance or aggregation results |
| Pass/Fail Outcome | Completion result when known |
| Source | Source document facts and publish metadata |

## Safety Rules

- Do not overwrite meaningful human-authored narrative automatically
- Do not fabricate execution or verification evidence
- Do not guess source identifiers or URLs
- Do not mark as `done` without recorded completion evidence
- Do not commit brief state changes to git
