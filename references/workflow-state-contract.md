# Workflow State Contract

Shared vocabulary for `task`, `run`, `verify`, and `checkup`.

## Core Fields

Every workflow-driven work item carries these fields:

- `Task Shape`
- `Executability`
- `Approval State`
- `Execution Path`
- `Current Stage`
- `Next Stage`
- `Workflow Entry State`
- `Retry Count` (default: 3)

Related local briefs may carry:

- `Workflow State`
- `Work Breakdown` when the task is composite
- `Pass/Fail Outcome` after verification
- `Failure Context` payload for self-correction

## Task Shape
- `atomic`: smallest executable unit
- `composite`: container for child work items; must be decomposed

## Executability
- `executable`: can enter `run`
- `container`: must stay in `definition` or advance to `decompose`

## Approval State
- `draft`, `approval_ready`, `approved`

## Stage Vocabulary
- `definition`: item defined, execution not started
- `decompose`: item needs breakdown into child tasks
- `run`: active execution or implementation
- `verify`: validation of results
- `checkup`: reconciliation and final alignment
- `done`: terminal state

## Workflow Entry State
- `ready_to_start`: may enter next stage immediately
- `needs_breakdown`: composite item needs fission
- `pending_acceptance`: automated checks passed, human acceptance required
- `blocked`: terminal failure or missing prerequisite

## Pass/Fail Outcome
- `pass`: verification successful
- `fail_retryable`: verification failed, self-correction possible
- `fail_terminal`: verification failed, max retries reached

## Advancement Rules

### run
- Normal: `definition` → `run` (if approved and executable)
- Execution Complete: `run` → `verify`
- Self-Correction: if `Failure Context` present, adjust strategy and increment retry counter

### verify
- Success: `verify` → `checkup`
- Retryable Failure: `verify` → `run` (if retries < max)
- Terminal Failure: mark `blocked` (if retries >= max)
- Manual Acceptance: mark `pending_acceptance`

### checkup
- Alignment: `checkup` → `done`

## Interpretation
- `done` means execution, verification, and reconciliation are complete
- `fail_retryable` is an internal loop; it triggers a new execution attempt
- Execution is not proof of correctness; only `pass` enables `done`
