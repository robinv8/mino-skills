# Brief Contract

Defines the local brief format used by the skill set.

## Purpose

Keep `task`, `run`, `verify`, and `checkup` aligned on:

- what a local brief is
- which sections it should contain
- which sections are managed by skills
- which sections may remain human-authored

## Location

Default: `.mino/briefs/issue-<number>.md`

## Sections

A brief may contain:

- `Issue`
- `Type`
- `Status`
- `Delivery Metadata`
- `Objective`
- `Acceptance Criteria`
- `Verification`
- `Target Files`
- `Work Breakdown`
- `Execution Path`
- `Workflow State`
- `Completion Handoff`
- `Execution Summary`
- `Verification Summary`
- `Pass/Fail Outcome`
- `Open Questions / Warnings`
- `Source`

## Managed Sections

Managed by skills:

- `Issue`, `Type`, `Status`, `Delivery Metadata`
- `Objective`, `Acceptance Criteria`, `Verification`
- `Target Files`, `Work Breakdown`, `Execution Path`
- `Workflow State`, `Completion Handoff`
- `Execution Summary`, `Verification Summary`, `Pass/Fail Outcome`
- `Source`

`Open Questions / Warnings` is human-addable and must be preserved.

## Section Intent

| Section | Purpose |
|---------|---------|
| Issue | Source identity and stable URL |
| Type | `feature` or `bug` |
| Status | Local brief status: `open`, `synced` |
| Delivery Metadata | Priority, risk |
| Objective | Compressed execution-oriented summary |
| Acceptance Criteria | Observable outcomes to satisfy |
| Verification | Repository-native or manual checks |
| Target Files | Implementation/review scope |
| Work Breakdown | Required for composite work |
| Execution Path | Expected workflow: `run → verify → checkup` |
| Workflow State | Current stage, next stage, entry state |
| Completion Handoff | Execution handoff for verify/checkup |
| Execution Summary | Concrete execution results |
| Verification Summary | Concrete verification results |
| Pass/Fail Outcome | Verification result |
| Source | Source system facts |

## Safety Rules

- Do not overwrite meaningful human-authored narrative automatically
- Do not fabricate execution or verification evidence
- Do not guess source identifiers or URLs
- Do not mark as done without recorded evidence
