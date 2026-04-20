# Iron Tree Protocol

> Version: 1.2
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
