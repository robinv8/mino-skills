---
title: Composite Tasks
---

# Composite Tasks

In the Iron Tree Protocol, tasks are classified into three types:

| Classification | Behavior |
|---|---|
| `atomic` | Executable leaf task. Has target files and acceptance criteria. |
| `composite` | Parent task that decomposes into child tasks. Not directly executable. |
| `container` | Similar to composite; acts as a grouping parent. Not directly executable. |

## DAG Structure

A composite parent and its children form a Directed Acyclic Graph (DAG) linked by `depends_on`:

```
┌─────────────────┐
│   composite-1   │
│  (container)    │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐ ┌───────┐
│ task-A│ │ task-B│
│atomic │ │atomic │
└───┬───┘ └───┬───┘
    │         │
    └────┬────┘
         ▼
    ┌─────────┐
    │  task-C │
    │ atomic  │
    └─────────┘
```

## Execution Rules

- `run` skips `container` and `composite` tasks; they decompose, not execute.
- Child tasks must all reach `done` before the parent can be aggregated.
- `checkup aggregate <issue>` finalizes a composite parent once all required children are `done`.

## Aggregation

`aggregate` performs the following:

1. Confirm `Classification` is `composite` or `container`
2. Resolve required child task keys from `Work Breakdown` and `Dependencies`
3. Verify every child has `Pass/Fail Outcome: pass` and `Current Stage: done`
4. Replace `Verification Summary` with an aggregate evidence list
5. Set `Completion Basis: aggregated`, `Code Publication State: not_applicable`
6. Emit `checkup_aggregate_recorded` event (silent)
7. Proceed to **Finalize**

## Adoption Guard

When adopting existing issues, composite issues (≥ 3 open checkboxes) are **refused** with the `iron-tree:needs-breakdown` label. You must split them into child issues first, then adopt each child individually.
