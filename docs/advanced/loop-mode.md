---
title: Loop Mode
---

# Loop Mode

Loop Mode is the default execution model for `/mino-task` since v0.6.0. After you approve a task set, the orchestrator autonomously drives `run` → `verify` → `checkup finalize` for every in-scope task without further human invocation, until a halt condition fires.

## Entry Points

`/mino-task` accepts free-form text and resolves it into a frozen task set:

| Pattern | Action |
|---|---|
| `PRD.md` (file path) | Native PRD flow → publish → Loop with `goal_kind: set_done` |
| `#123` | Adopt single issue → Loop with `goal_kind: task_done` |
| `#45 #47` | Adopt multiple issues → Loop with `set_done` |
| `前 N 条 issue` / `first N` / `top N` | Canonical query → Loop with `set_done` |
| `all open` / `所有 OPEN` | All open adopted issues → Loop with `set_done` |
| `resume <loop_id>` | Resume a halted Loop |

## Resolved Plan & Approval

Before entering Loop Mode, `/mino-task` prints a **Resolved Plan** and requires explicit `yes`:

```
You are authorizing Loop Mode to autonomously execute the following plan.

Loop ID:        {loop_id}
Goal:           {task_done | set_done}
Intent:         {verbatim user input}
Resolved query: {one-line summary or "n/a (file path)"}
Tasks ({len}):  budget = {budget_max_transitions} transitions
  1. #<N>  <title>  <task_key>
  2. ...
Excluded ({len}, see notes):
  - #<N>  <reason: composite / closed / not adopted / etc>

Halts on: approval-required, pending_acceptance, fail_terminal, blocked,
          reapproval_required, loop_budget_exhausted.
Stepwise opt-out: invoke /mino-run, /mino-verify, /mino-checkup directly.

Approve and start Loop? (yes / edit / cancel)
```

`yes` is the **explicit Loop Mode opt-in** required by protocol § Invariants.

## Halt Conditions

The Loop stops on the first of these conditions (evaluated in protocol order):

1. `approval-required` — DAG approval needed
2. `pending_acceptance` — verify cannot auto-prove correctness
3. `fail_terminal` — retry budget exhausted
4. `blocked` — pre-flight or external event blocks progress
5. `reapproval_required` — spec revision drifted from approved revision
6. `loop_budget_exhausted` — max transitions reached (safety rail)
7. `protocol_gap` — unrecoverable state inconsistency

Halts stop the **entire** Loop. Loop Mode never auto-skips an offending task. Skipping is a human act via `/mino-task resume <loop_id> skip <task_key>`.

## Resume Mode

```
/mino-task resume <loop_id> [continue | skip <task_key> | cancel]
```

| Sub-command | Effect |
|---|---|
| `continue` | Re-acquire lease, emit `loop_resumed`, re-enter Driver Iteration |
| `skip <task_key>` | Mark task cancelled; cascade to in-loop dependents that depend on it |
| `cancel` | Set status `cancelled`, release lease, exit |

## Loop Entity

Each Loop writes an authoritative entity to `.mino/loops/{loop_id}.yml`:

- `goal`, `frozen_task_set`, `budget`, `status`, `halt_reason`
- `transitions` array: `{iso, task_key, skill, outcome}`

A repo-level lease `.mino/loops/active.lock` prevents concurrent Loops. Stale leases (PID gone or heartbeat > 6h) are auto-detected and cleaned on takeover.

## Stepwise Opt-Out

Direct invocation of `/mino-run`, `/mino-verify`, `/mino-checkup` continues to work exactly as before. These skills detect orchestrator mode by the presence of `.mino/loops/active.lock` and switch to silent return only when an orchestrator holds the lease.
