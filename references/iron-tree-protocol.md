# Iron Tree Protocol

> Version: 1.0
> Purpose: Define the recursive, low-touch execution engine.

## Concept

A human-approved requirement document unfolds into a fully implemented, verified, and reconciled feature without further human intervention (unless blocked).

Shifts from "step-by-step human prompting" to "state-machine driven autonomous orchestration".

## Core Loop

### Phase 1: Document Intake (`task`)
1. Read the full Markdown document
2. Extract a DAG draft with explicit `depends_on`
3. Wait for human approval before publishing

### Phase 2: Publish & Approval
1. Write tasks to the source system (GitHub issues)
2. Create or refresh local briefs
3. Automation begins only after explicit approval

### Phase 3: Execution (`run`)
1. **Pre-flight**: `checkup` validates environment readiness
2. **Serial Scheduling**: select next atomic task whose `depends_on` are `done`
3. **Locking**: assert lock on target files
4. **Implementation**: modify codebase
5. **Handoff**: transition active node to `verify`

### Phase 4: Validation (`verify`)
1. Execute native repository checks (build, test, lint)
2. **Self-Correction Loop**: if failed:
   - Capture exact error output (Failure Context)
   - Increment retry counter (limit: 3)
   - Transition back to `run` with Failure Context
3. **Manual Acceptance**: if automation insufficient, emit explicit human steps and stop at `pending_acceptance`
4. **Terminal Failure**: if retries exhausted, mark `blocked`
5. **Success**: transition to `checkup`

### Phase 5: Reconciliation (`checkup`)
1. Align local brief with final source truth
2. Transition task to `done`
3. Evaluate next serially eligible DAG node

## DAG Rules

- A task cannot enter `run` until all `depends_on` tasks are `done`
- V1 executes DAG nodes serially, even sibling tasks

## Required Capabilities

- **Predictive DAG (`task`)**: Extract full tree structure before execution
- **Serial Scheduling (`run`)**: Respect `depends_on`, order, file locks
- **Structured Failure (`verify`)**: Feed actionable errors back to `run`
- **Manual Acceptance (`verify`)**: Emit explicit human steps when automation insufficient
- **Pre-flight (`checkup`)**: Ensure runway is clear before takeoff
