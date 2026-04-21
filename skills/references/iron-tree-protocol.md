# Iron Tree Protocol

> Version: 1.9
> Purpose: Define the recursive, low-touch execution engine.

## Concept

A human-approved requirement document unfolds into a fully implemented, verified, accepted when necessary, aggregated when composite, and reconciled feature without further human steering unless the workflow is blocked.

The protocol shifts work from ad hoc prompting to a state-machine driven loop with explicit gates, durable metadata, revision-aware approval, and deterministic recovery.

## Core Loop

### Phase 1: Document Intake (`task`)

1. Read the full Markdown document
2. Extract a DAG draft with explicit `depends_on`
3. Classify each node
4. Compute a deterministic `Task Key` per node
5. Compute a `Spec Revision` for the proposed DAG
6. Wait for human approval before publishing

### Phase 2: Publish (`task`)

1. Create or refresh source tasks idempotently using stable `Task Key`s
2. If the approved revision changed, require re-approval before refresh
3. Write initial issue metadata and local briefs
4. Initialize each task at `Current Stage: definition`
5. Record `Approved Revision = Spec Revision`

### Phase 3: Readiness Gate (`checkup pre-flight`)

1. Validate environment readiness before execution
2. Check task-specific prerequisites
3. If the runway is broken, mark the task `blocked`
4. Never transition a task to `done` during pre-flight

## Execution Lock

The protocol guarantees serial execution of `run` at the repository level. Only one `run` may be active at any time.

- Before entering Phase 4, `run` checks for `.mino/run.lock`
- If the lock exists, `run` refuses execution and reports the active task key and start time
- On successful entry, `run` writes the lock with its task key and ISO timestamp
- On normal completion or unrecoverable failure, `run` removes the lock
- A stale lock (older than a configurable threshold, default 2 hours) may be overridden with explicit confirmation

V1 executes DAG nodes serially even when they have no mutual dependencies. Parallel execution is deferred to v3 after conflict detection and file-lock advisory semantics are fully specified.

### Phase 4: Execution (`run`)

1. Check `.mino/run.lock`; if present, refuse execution
2. Acquire lock with task key and ISO timestamp
3. Rebuild the DAG from brief metadata, using issue metadata as fallback
4. Select the next atomic task whose `depends_on` are `done`
5. Assert an advisory lock on target files
6. Increment `Attempt Count`
7. Modify the codebase
8. Commit changes locally so `verify` has a stable SHA to anchor against
9. Handoff to `verify`

If the commit step fails (e.g., a pre-commit hook rejects, identity not configured), `run` does not consume retry budget: it persists the error in `Failure Context`, emits `run_commit_failed`, decrements `Attempt Count` back to the pre-run value, and halts. This mirrors `verify_publication_failed` for symmetry between the two skills' publication boundaries.

### Phase 5: Validation (`verify`)

1. Record `Verify Anchor SHA` = current HEAD commit
2. Execute explicit repo-native checks, preferring `.mino/config.yml` overrides
3. Compare observed results to acceptance criteria
4. On recoverable failure, write `Failure Context` and hand control back to `run`
5. On unrecoverable failure, mark the task `blocked`
6. If automation cannot complete verification, stop at `pending_acceptance`
7. On success, publish code first if needed, then transition to `checkup`
8. If publication fails after checks pass, stay in `verify`, preserve the local code state, record publication failure context, and retry publication instead of falsely recording success; this does not consume retry budget

## Verify Anchor

`verify` evaluates the codebase at a specific commit, not the working tree. This guarantees deterministic verification even if the user edits files during a long-running check.

- `run` must commit its changes before handoff to `verify`
- `verify` records `Verify Anchor SHA` = `HEAD` at the moment it starts
- The verification result is bound to that SHA; subsequent uncommitted changes do not invalidate the result
- If `verify` fails with `fail_retryable`, the next `run` starts from the same SHA or a newer committed state

This field is required in Loop Mode so the Decision Function can distinguish between a verify-pass at a stale SHA and a verify-pass at the current SHA.

### Phase 6: Acceptance Or Aggregation (`checkup`)

1. Record human acceptance when the task is in `pending_acceptance`
2. Or aggregate child completion for composite/container parents
3. Bind completion evidence to a published `code_ref` when code changed
4. If code publication fails during acceptance, do not record acceptance; preserve `pending_acceptance`, keep the local code state, and record publication failure context for retry
5. Reconcile the local brief with the authoritative issue record
6. Transition `checkup` → `done` only when completion evidence exists
7. Evaluate the next serially eligible DAG node

## Source Of Truth

- Stable task metadata lives in the GitHub issue body
- Workflow transitions live in structured workflow events posted as issue comments
- Local briefs are a cache for scheduling, inspection, and recovery
- `checkup reconcile` repairs local drift by replaying valid events for the active approved revision

## External Events

The source system (GitHub) is authoritative for issue existence and visibility, but workflow state is authoritative for completion. When an external event contradicts workflow state, the protocol must not silently overwrite workflow state.

### Issue Closed Externally

When `checkup reconcile` discovers that an issue is closed on GitHub but has no `checkup_done` or equivalent terminal event in its workflow history:

1. Do **not** sync the brief to `done`
2. Record `External Event: issue_closed` on the brief
3. Set `Workflow Entry State: blocked`
4. Post a `checkup_reconcile_external_close_detected` event to the issue
5. Require human confirmation before the task can advance

This preserves the invariant that `done` requires recorded completion evidence, not just a closed issue tracker entry.

## Multi-Agent Git Hygiene

The protocol assumes the remote (origin/main and any shared branches) is **append-only from each agent's perspective**. Multiple agents may run sequentially or concurrently against the same repository, and any one of them rewriting shared history silently destroys another agent's work.

Hard rules for every skill that touches git:

1. **Never `git push --force` or `--force-with-lease`** against a branch that has been pushed by anyone else, including a previous agent run.
2. **Never `git reset --hard` to a commit older than the current `origin/<branch>` tip** when intending to push afterwards. Use `git revert` to undo published commits instead.
3. **Never rebase or amend a commit that already exists on the remote.** Local-only commits may be rebased; once pushed, they are immutable.
4. **Never delete a remote branch** unless the user explicitly asks.
5. If reconciliation requires rewriting history (e.g., secret leak), stop and ask the user — this is outside the agent's authority.

Rationale: agent collaboration without these rules degenerates into a last-writer-wins race where revert commits, `.gitignore` rules, and protocol fixes vanish without trace. The cost of `git revert` (one extra commit) is always lower than the cost of recovering lost work from a forced push.

## Adopting Existing Issues

The protocol's native entry point is `/task <spec.md>`, which posts `task_published` at sequence 1. Repositories that pre-existed the protocol have issues without this event; `run` / `verify` / `checkup` therefore refuse to operate on them.

`/task adopt issue-N` is the standard on-ramp. It produces the **same shape of artifacts** as native publication so downstream skills cannot tell the difference:

- A local brief at `.mino/briefs/issue-{N}.md` (same template as native)
- An event yml at `.mino/events/issue-{N}/0001-task-adopted.yml` with `event: task_adopted` and `sequence: 1`
- An issue comment carrying the same yml block, so reconciliation can replay it
- A pair of GitHub labels marking workflow position: `iron-tree:adopted` (permanent) + `stage:task` (mutable)

### Eligibility

`/task adopt issue-N` accepts an issue iff:

1. The issue exists and is `OPEN` on the host repository
2. The issue body does not declare more than `COMPOSITE_THRESHOLD = 3` open checkboxes (`- [ ]`); composite issues must be broken into child issues by the human first
3. Any prior `iron-tree:adopted` label triggers **re-adopt** semantics, not refusal

Closed issues are refused with a clear error.

### Spec Identity for Adopted Issues

Adopted issues have no Markdown spec. The protocol synthesizes:

```
spec_path     = github://issue-{N}
spec_revision = sha256( normalize(issue_title) + "\n---\n" + normalize(issue_body) )[:8]
```

`normalize` is the same transformation used by `task` (strip trailing whitespace, CRLF→LF, collapse blank-line runs).

### Re-adoption

If `iron-tree:adopted` already exists on the issue:

1. Move the existing `.mino/briefs/issue-{N}.md` and `.mino/events/issue-{N}/` into `.mino/archive/issue-{N}-rev-{previous_spec_revision}/`
2. Compute fresh `spec_revision` from current title + body
3. Emit `task_re_adopted` (sequence 1 of the new chain) instead of `task_adopted`; the event carries `previous_revision` and `archive_path`
4. Remove any `stage:run` / `stage:verify` / `stage:done` label and apply `stage:task`
5. Approval must be re-requested from the user (same gate as a native re-approval)

### Stage Label Lifecycle

`stage:*` labels are **mutually exclusive** and mirror the active workflow stage:

| Transition | Performed by | Trigger |
|---|---|---|
| (none) → `stage:task` | `task` | `/task adopt` succeeds, or native `task` publishes |
| `stage:task` → `stage:run` | `task` | User approval recorded |
| `stage:run` → `stage:verify` | `run` | Step 7 commit succeeds (or `not_applicable` path) |
| `stage:verify` → `stage:done` | `verify` | `verify_passed` recorded |

Failure paths **do not move the label**: a stuck issue's GitHub label tells the human exactly where it halted. Label sync failures (gh CLI down, missing permission) are non-fatal: the skill records a `stage_label_sync_failed` warning in its report and continues; the local yml remains authoritative.

### PR Out of Scope

Pull Requests are not adopted. PRs continue to merge against an issue; merging does not by itself transition any label.

## DAG Rules

- A task cannot enter `run` until all `depends_on` tasks are `done`
- V1 executes DAG nodes serially, even sibling tasks
- Container tasks never enter execution directly
- Composite parents may complete through aggregation once all required children are `done`

## Required Capabilities

- **Predictive DAG (`task`)**: Extract full tree structure before execution
- **Revision-Aware Approval (`task`)**: Tie approval to a specific `Spec Revision`
- **Idempotent Publish (`task`)**: Refresh existing issues and briefs instead of duplicating them on rerun
- **Pre-flight (`checkup`)**: Ensure the runway is clear before takeoff
- **Serial Scheduling (`run`)**: Respect `depends_on`, order, attempt budget, and advisory file locks
- **Structured Failure (`verify`)**: Feed actionable errors back to `run`
- **Publish-Before-Pass (`verify`)**: Publish accepted code before recording a successful terminal state
- **Manual Acceptance (`verify` + `checkup`)**: Emit explicit human steps, then record acceptance against a published code ref
- **Aggregate Completion (`checkup`)**: Finalize composite parents from child completion evidence
- **Structured State Events (all skills)**: Post machine-readable workflow events so `checkup reconcile` can recover state deterministically

## Execution Modes

The Core Loop above defines **what** transitions are legal. Execution Modes define **who drives** the transitions and **when control returns to a human**.

Two modes are defined. Both operate on the same state machine and produce identical state events. Only the driver differs.

### Stepwise Mode (default)

- A human invokes one skill at a time (`/task`, `/run`, `/verify`, `/checkup`)
- Each skill performs exactly one transition and returns control to the human
- The human decides what to invoke next based on the resulting state
- This is the safe baseline; every skill must remain usable in stepwise mode

### Loop Mode

- The agent invokes skills autonomously in sequence until a `Halt Condition` is reached
- The agent does not require a human invocation between transitions
- The agent must continuously observe authoritative state (issue body and events, repaired by `checkup reconcile` when needed) before each next action; it must not act on stale local cache
- All transitions remain governed by the Core Loop and `workflow-state-contract.md`; Loop Mode adds no new transitions and weakens no gates
- Loop Mode is opt-in per goal; the default remains stepwise

### Goal

Loop Mode requires an explicit, observable goal. Two goal shapes are recognized:

- `task_done`: a specific `Task Key` (and, for composites, all required children) reaches `done`
- `set_done`: every `Task Key` in a named set reaches `done`

A goal is **observable** when its completion can be determined from issue state plus structured workflow events alone. Goals that require subjective judgement are not valid Loop Mode goals.

### Halt Conditions

The agent **must immediately halt** Loop Mode and return control to a human when any of the following are true. Halting is not failure; halting is the protocol explicitly handing the next decision to a human.

Halt conditions are evaluated **in the order listed**. The first matching condition wins; this prevents more general gates from masking more specific outcomes.

| # | Halt Reason | Trigger | Resolution |
|---|---|---|---|
| 1 | `goal_reached` | Goal predicate satisfied | None; report success |
| 2 | `fail_terminal` | Any in-scope task has `Pass/Fail Outcome: fail_terminal` | Human revises spec, code, or tooling |
| 3 | `pending_acceptance` | Any in-scope task has `Workflow Entry State: pending_acceptance` | Human runs `/checkup accept` |
| 4 | `reapproval_required` | Any in-scope task has `Spec Revision != Approved Revision` | Human re-approves via `/task` |
| 5 | `blocked` | Any in-scope task has `Workflow Entry State: blocked` **and** `Pass/Fail Outcome != fail_terminal` | Human investigates and resolves |
| 6 | `protocol_gap` | The agent encounters a state not covered by the Core Loop or contract | Human extends the protocol |
| 7 | `loop_budget_exhausted` | The configured maximum number of consecutive transitions is reached | Human inspects and decides whether to resume |

`fail_terminal` is evaluated before `blocked` because a terminal verification failure always also sets `Workflow Entry State: blocked`; without ordered evaluation, terminal failures would be reported under the more general `blocked` reason and lose diagnostic specificity.

When halting, the agent must:

1. Post a structured event with `event: loop_halted` and a `halt_reason` field matching one of the values above
2. Set `Halt Reason` on the affected task's brief (see `workflow-state-contract.md`)
3. Stop issuing further skill invocations until a human re-engages

### Decision Function

Within a Loop Mode iteration, the agent selects the next action by evaluating in order:

1. If any halt condition is true → halt
2. If a task is in `Current Stage: run` with `Next Stage: verify` → invoke `/verify`
3. If a task is in `Current Stage: verify` with `Pass/Fail Outcome: fail_retryable` and `Attempt Count <= Max Retry Count` → invoke `/run` (which performs its own pre-flight before scheduling)
4. If a task is in `Current Stage: checkup` with `Next Stage: done` and `Completion Basis` set → invoke `/checkup finalize <issue>` to bind completion evidence and transition to `done`
5. If all required children of an in-scope composite parent are `done` and the parent is not yet `done` → invoke `/checkup aggregate` on the parent
6. If a DAG node has all `depends_on` `done` and `Workflow Entry State: ready_to_start` → invoke `/run` on the eligible node with the lowest `sequence` of last activity (deterministic tie-break)
7. Otherwise → halt with `protocol_gap`

Notes:

- `pre-flight` is not a Decision Function step. The contract requires `run` to perform `checkup pre-flight` internally before scheduling work; Loop Mode relies on that internal call rather than scheduling pre-flight separately.
- Step 4 requires a `finalize` mode on the `checkup` skill. This mode performs only the finalization sub-step (bind `code_ref`, write the terminal event, transition to `done`) and never performs `accept` or `aggregate` work; those remain explicit, separately invoked modes. Skills that have not yet implemented `finalize` are not Loop-Mode-ready for verify-pass paths and must remain in stepwise mode for those tasks.

The decision function is intentionally narrow. New decision branches must be added to the protocol before any skill or agent uses them in Loop Mode.

### Loop Budget

- A Loop Mode session must declare a `Max Consecutive Transitions` upper bound (default `50`)
- Each invoked skill counts as one transition regardless of outcome
- Reaching the budget halts with `loop_budget_exhausted`; this is a safety valve against runaway loops, not a verdict on the work

### Invariants

- Loop Mode never bypasses approval gates. Initial DAG approval still requires a human; only post-approval execution is automated.
- Loop Mode never weakens `pending_acceptance`. Human acceptance always remains a required state transition for the affected task.
- Loop Mode emits the same workflow events as stepwise mode. A reconciliation run cannot tell the two modes apart from event history alone, except for `loop_halted` and `loop_resumed` markers.
- A skill that is not safe in stepwise mode is not safe in Loop Mode. There is no "Loop Mode only" capability.
