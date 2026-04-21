# Iron Tree Protocol

> Version: 1.3
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

### Phase 4: Execution (`run`)

1. Rebuild the DAG from brief metadata, using issue metadata as fallback
2. Select the next atomic task whose `depends_on` are `done`
3. Assert an advisory lock on target files
4. Increment `Attempt Count`
5. Modify the codebase
6. Handoff to `verify`

### Phase 5: Validation (`verify`)

1. Execute explicit repo-native checks, preferring `.mino/config.yml` overrides
2. Compare observed results to acceptance criteria
3. On recoverable failure, write `Failure Context` and hand control back to `run`
4. On unrecoverable failure, mark the task `blocked`
5. If automation cannot complete verification, stop at `pending_acceptance`
6. On success, publish code first if needed, then transition to `checkup`
7. If publication fails after checks pass, stay in `verify`, preserve the local code state, record publication failure context, and retry publication instead of falsely recording success; this does not consume retry budget

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
