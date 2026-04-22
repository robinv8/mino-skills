# Workflow State Contract

Shared vocabulary for `task`, `run`, `verify`, and `checkup`.

## Core Fields

Every workflow-driven work item must carry these fields:

- `Task Key` — stable identifier for the task, ideally a deterministic slug derived from the spec path, parent key, and title. Published issues may also reference their GitHub issue number, but the task key itself must remain stable across reruns.
- `Issue Number` — source-system locator after publication
- `Spec Revision` — normalized hash of the source spec and approved DAG shape
- `Approved Revision` — the exact `Spec Revision` last approved by a human
- `Task Shape`
- `Executability`
- `Approval State`
- `Depends On`
- `Current Stage`
- `Next Stage`
- `Workflow Entry State`
- `Attempt Count` (starts at `0`, increments when `run` begins)
- `Max Retry Count` (default `3`, meaning one initial attempt plus up to three retries)
- `Code Publication State`
- `Pass/Fail Outcome` (optional until completion evidence exists)
- `Completion Basis` (optional until the task is ready to finalize)
- `Halt Reason` (optional; set only when Loop Mode halts on this task; see `iron-tree-protocol.md` § Execution Modes)
- `Verify Anchor SHA` (optional; set when `verify` starts; records the commit SHA being verified)

Local briefs also carry workflow cache fields for scheduling and recovery:

- `Workflow State`
- `Dependencies`
- `Work Breakdown` when the task is composite
- `Manual Acceptance`
- `Failure Context`
- `Execution Summary`
- `Verification Summary`
- `External Event` (optional; populated by `checkup reconcile` when an out-of-band event is detected, e.g., the linked issue is closed without a recorded `checkup_done`)

The linked GitHub issue body plus structured workflow events are the authoritative record. Brief state is a local cache for fast DAG scheduling and may drift; `checkup reconcile` repairs the brief from issue metadata and the latest valid workflow event sequence.

Brief state updates are local-only. Do NOT stage or commit `.mino/briefs/` files.

## Task Shape

- `atomic`: smallest executable unit
- `composite`: container for child work items; must be decomposed

## Executability

- `executable`: can enter `run`
- `container`: must stay in `definition` or advance to `decompose`

## Approval State

- `draft`
- `approval_ready`
- `approved`

If `Spec Revision` changes and no longer matches `Approved Revision`, the task must return to `Approval State: approval_ready` before any refresh or execution.

## Stage Vocabulary

- `definition`: item exists and is fully described, but execution has not started
- `decompose`: item needs breakdown into child tasks
- `run`: active execution or implementation
- `verify`: validation of results
- `checkup`: reconciliation and final alignment
- `done`: terminal state

## Next Stage

`Next Stage` should be one of `decompose`, `run`, `verify`, `checkup`, `done`, or `none`.

Use `none` only when the task is terminally blocked or already `done`.

## Workflow Entry State

- `ready_to_start`: may enter `Next Stage` immediately
- `needs_breakdown`: composite item needs fission
- `pending_acceptance`: automated validation cannot finish the task; human review is required
- `blocked`: terminal failure or missing prerequisite

`pending_acceptance` and `blocked` are gates, not stages.

## Attempt And Retry Semantics

- `Attempt Count` counts execution attempts and increments exactly once when `run` starts
- `Max Retry Count` defaults to `3`
- Allowed total attempts = `1 + Max Retry Count`
- When a verification failure happens on attempt `N`, it is retryable if `N <= Max Retry Count`
- In practice, with `Max Retry Count: 3`, failures on attempts `1`, `2`, and `3` may still yield `fail_retryable`; a failure on attempt `4` must end in `fail_terminal`

This makes retry budget deterministic without relying on ambiguous off-by-one interpretation.

## Code Publication State

- `not_applicable`: no code artifact needs publication for this transition
- `local_only`: relevant code changes exist locally but have not been published
- `published`: the accepted code state has been committed and pushed, or otherwise durably published

`checkup` may only finalize `done` when `Code Publication State` is `published` or `not_applicable`.

## Pass/Fail Outcome

- `pass`: completion evidence exists
- `fail_retryable`: verification failed and control returns to `run`
- `fail_terminal`: verification failed and should not be retried

When validation has not yet completed, `Pass/Fail Outcome` may be omitted.

## Completion Basis

- `verified`: automated checks passed and the published code ref is recorded
- `accepted`: a human accepted the published code ref
- `aggregated`: a composite task was satisfied by recorded completion of its child tasks

## Structured Workflow Events

Every state-changing event has two sinks:

1. **Local event file** — `.mino/events/issue-{N}/{sequence:04d}-{event-kebab}.yml`. Authoritative. Must be written before any comment is attempted.
2. **GitHub comment** — optional mirror; see § Event Categories.

The event schema inside both sinks is identical:

```yaml
iron_tree:
  version: 1
  task_key: dark-mode-toggle
  issue_number: 8
  spec_revision: a1b2c3d4
  approved_revision: a1b2c3d4
  sequence: 4
  event: verify_failed_retryable
  current_stage: run
  next_stage: verify
  workflow_entry_state: ready_to_start
  approval_state: approved
  attempt_count: 2
  max_retry_count: 3
  code_publication_state: local_only
  pass_fail_outcome: fail_retryable
  completion_basis: null
  code_ref: null
```

Rules:

- `sequence` must increase monotonically by `1` for each state-changing event on the same `Task Key` and `Approved Revision`, counted across local events regardless of comment visibility.
- Recovery must ignore malformed events and any event whose `Approved Revision` does not match the task's current approved revision.
- Human free-form comments without a valid event block are informational only and must not drive reconciliation.
- When a comment is posted, it must exactly reproduce the local file's yml block; skills must not modify, re-serialize, or add fields on the way out.

## Event Categories

Events fall into exactly one category:

### Silent (local only)

These events do not produce GitHub comments. They are recorded solely in the local log.

- `task_published`
- `task_adopted`
- `task_re_adopted`
- `run_started`
- `run_completed`
- `verify_passed`
- `checkup_accept_recorded`
- `checkup_aggregate_recorded`
- `loop_resumed`

### Audible (post immediately)

These events demand human attention and post a comment with a human-readable narrative and the full yml block, at the time of emission.

- `task_reapproval_required`
- `run_commit_failed`
- `verify_failed_retryable`
- `verify_failed_terminal`
- `verify_publication_failed`
- `verify_pending_acceptance`
- `loop_halted`
- `checkup_preflight_blocked`
- `checkup_accept_publication_failed`
- `checkup_reconcile_external_close_detected`
- `checkup_reconcile_sequence_gap_detected`

Audible comments are human notifications only: a one-line heading (`⚠️` / `⏸️` + event description + `#{N}`), an optional `Reason:` line, and an `Action:` line telling the operator what to do next. Do NOT inline the rendered YAML and do NOT include a `Local events:` pointer — the local file at `.mino/events/issue-{N}/{seq:04d}-{event-kebab}.yml` is the sole authoritative record. The GitHub comment is a notification channel; if it is lost, the local log is unaffected.

### Terminal (consolidated summary)

Only one event is terminal:

- `checkup_done`

Its comment is a short, fixed-format completion notice: heading `🏁 Issue #{N} done — {task_key}`, plus three bullets (`Completion Basis`, `Code Ref`, `Code Publication State`). See `skills/mino-checkup/templates/comment-checkup-summary.md.tmpl`. The local event chain at `.mino/events/issue-{N}/*.yml` is the sole authoritative log; GitHub comments are a notification channel and are not used for replay.

## Loop Events

Loop-level events use a `loop:` top-level block (instead of `iron_tree:`) and live at `.mino/loops/{loop_id}/events/{seq:04d}-{event-kebab}.yml`.

- `loop_started` -- written when `/mino-task` enters Loop Mode after approval. Fields: `loop_id`, `goal_kind`, `task_keys`, `budget_max_transitions`, `intent_hash` (sha256 of normalized intent text).
- `loop_halted` -- written when a halt condition fires. Fields: `loop_id`, `halt_reason` (one of the 7 protocol values), `halt_at_task_key`, `transitions_used`.
- `loop_resumed` -- written when `/mino-task resume <loop_id> continue` re-enters the driver. Fields: `loop_id`, `previous_halt_reason`, `transitions_used`.
- `loop_completed` -- written when every `task_key` reaches `done`. Fields: `loop_id`, `completed_at`, `transitions_used`.
- `loop_cancelled` -- written on `cancel` (or on `skip <task_key>` for the cancelled task only). Fields: `loop_id`, `cancelled_at`, `cancellation_scope` (`loop` or `task`), `task_key` (when scope=task).

All loop events share a top-level structure analogous to issue events but with `loop:` instead of `iron_tree:`:

```yaml
loop:
  version: 1
  loop_id: <id>
  sequence: <int>          # monotonic within .mino/loops/{loop_id}/events/
  event: loop_started
  # event-specific fields below
```

Storage: `.mino/loops/{loop_id}/events/{seq:04d}-{event-kebab}.yml`.

## Back-Compatibility with Legacy Chains

Issues completed under protocol ≤ v1.9 have per-event comments. `checkup reconcile` continues to accept them as a secondary source. v1.10 skills never emit new silent-category comments, so the corpus of per-event silent comments cannot grow after upgrade. Legacy chains must not be rewritten or squashed.

## Halt Reason

`Halt Reason` is set on a task's brief only when Loop Mode halts on that task. It mirrors the `halt_reason` field on the `loop_halted` event and is cleared on the next successful transition out of the halting state.

Allowed values:

- `goal_reached`
- `pending_acceptance`
- `blocked`
- `fail_terminal`
- `reapproval_required`
- `protocol_gap`
- `loop_budget_exhausted`

`Halt Reason` is informational. It does not by itself permit any state transition; it records why automated execution stopped so a human can resume from a known point.

When set on a brief, `Halt Reason` is a **diagnostic mirror** of the most recent `loop_halted` event for the affected task. The Loop Entity at `.mino/loops/{loop_id}.yml` is the authoritative record. Tools that need to discover whether a task is halted in a Loop must consult the Loop Entity, not the brief. The brief mirror exists so a human reading a single brief can see "this task is currently blocking loop X" without cross-referencing the loop registry.

## Pending Acceptance Coordination

When a task enters `Workflow Entry State: pending_acceptance`:

- The detailed manual verification checklist lives in the local brief under `Manual Acceptance`
- The linked issue should receive the label `pending-acceptance` for easy multi-user discovery
- The linked issue should receive a short summary comment containing the reason and the action to run `/mino-checkup accept issue-{N}`
- The summary comment is for shared visibility; the detailed checklist remains local brief data

When a task leaves `pending_acceptance` through `checkup accept`:

- `checkup` records the human result in an issue comment with reviewer, result, timestamp, `code_ref`, and notes
- `checkup` removes the `pending-acceptance` label

The label is an index for querying pending tasks. It is not the authoritative workflow state by itself.

## Identity Model

- `Task Key` is the canonical logical identity
- `Issue Number` is the published source-system locator
- Local brief paths default to `.mino/briefs/issue-<Issue Number>.md` after publication
- Skills may accept either a task key or an issue locator from the user, but they must resolve back to `Task Key` before scheduling or reconciliation

## Advancement Rules

### task publish

- After explicit approval, `task` creates or refreshes issues and briefs
- `task` must compute `Spec Revision` from the normalized source doc and DAG
- If an existing task with the same `Task Key` has a different `Approved Revision`, `task` must require fresh approval for the current `Spec Revision` before publish continues
- Initial `Current Stage` is `definition`
- `container` / `composite` tasks start with `Next Stage: decompose` and `Workflow Entry State: needs_breakdown`
- `executable` / `atomic` tasks start with `Next Stage: run` and `Workflow Entry State: ready_to_start`
- Published tasks set `Approved Revision = Spec Revision`, `Attempt Count = 0`, `Max Retry Count = 3`, and `Code Publication State = not_applicable`

### run

- Precondition: `Approval State: approved`, `Approved Revision = Spec Revision`, `Executability: executable`, `Workflow Entry State: ready_to_start`
- `run` must perform `checkup pre-flight` before scheduling work
- `Attempt Count` increments once when `run` starts
- Execution start: `Current Stage: run`, `Next Stage: verify`
- If code files changed, `Code Publication State` becomes `local_only`
- Execution complete: `Current Stage: verify`
- Commit failure after execution: keep `Current Stage: run`, `Next Stage: verify`, `Workflow Entry State: ready_to_start`, `Code Publication State: local_only`, leave `Pass/Fail Outcome` and `Completion Basis` unset, persist the commit error in `Failure Context`, emit `run_commit_failed`, and decrement `Attempt Count` back to its pre-run value (commit failures must not consume retry budget, mirroring `verify_publication_failed`)

### verify

- Success: publish code first if needed, then record `Current Stage: checkup`, `Next Stage: done`, `Workflow Entry State: ready_to_start`, `Code Publication State: published|not_applicable`, `Pass/Fail Outcome: pass`, `Completion Basis: verified`
- Publication failure after checks pass: keep `Current Stage: verify`, `Next Stage: verify`, `Workflow Entry State: ready_to_start`, `Code Publication State: local_only`, leave `Pass/Fail Outcome` and `Completion Basis` unset, persist the publication error in `Failure Context`, emit `verify_publication_failed`, and do not increment `Attempt Count` or consume retry budget
- Retryable failure: `Current Stage: run`, `Next Stage: verify`, `Workflow Entry State: ready_to_start`, `Pass/Fail Outcome: fail_retryable`
- Terminal failure: `Current Stage: verify`, `Next Stage: none`, `Workflow Entry State: blocked`, `Pass/Fail Outcome: fail_terminal`
- Manual acceptance required: `Current Stage: verify`, `Next Stage: checkup`, `Workflow Entry State: pending_acceptance`
- No-tooling verification also routes to `pending_acceptance`; it does not auto-pass the task

### checkup

- `pre-flight` validates readiness and may mark `Workflow Entry State: blocked`, but must never write `done`
- `accept` records human acceptance for a task in `pending_acceptance`, publishes any remaining code with the standard commit style `[run] #{N}: {concise change summary}` if needed, binds the acceptance to the published `code_ref`, posts the shared acceptance record to the issue, removes the `pending-acceptance` label, then transitions it to `Current Stage: checkup`, `Next Stage: done`, `Workflow Entry State: ready_to_start`, `Pass/Fail Outcome: pass`, `Completion Basis: accepted`
- If publication fails during `accept`, do NOT record acceptance. Keep `Current Stage: verify`, `Next Stage: checkup`, `Workflow Entry State: pending_acceptance`, `Code Publication State: local_only`, leave `Pass/Fail Outcome` and `Completion Basis` unchanged, persist the publication error in `Failure Context`, post a short human-readable failure summary, emit `checkup_accept_publication_failed`, and do not advance to `done`
- `aggregate` records aggregate completion for a `composite` / `container` task after all required children are `done`; it transitions the parent to `Current Stage: checkup`, `Next Stage: done`, `Workflow Entry State: ready_to_start`, `Pass/Fail Outcome: pass`, `Completion Basis: aggregated`
- `reconcile` aligns the local brief with authoritative issue state by replaying the highest valid `sequence` for the active approved revision
- Final alignment: `checkup` → `done` only when `Completion Basis` is `verified`, `accepted`, or `aggregated`, and `Code Publication State` is not `local_only`

## Interpretation

- `done` means execution, verification or acceptance or aggregation, publication if needed, and reconciliation are complete
- `fail_retryable` is an internal loop; it hands control back to `run`
- Execution is not proof of correctness; only recorded completion evidence enables `done`
- A repository with no build/test/lint tooling must still go through manual acceptance before `done`
