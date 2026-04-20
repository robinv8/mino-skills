---
name: run
description: |
  Advance an approved task DAG through execution. Reads local briefs,
  respects dependency order, runs code changes, and self-corrects from
  prior verification failures. Stops on blocked or pending-acceptance states.
  Use when a task DAG has been approved and you want to execute it.
---

# Execution Engine

Read an approved task DAG from local briefs, schedule work in dependency order, execute code changes, and drive tasks toward completion.

## Workflow

1. **Resolve user target** — accept either an issue locator like `issue-8` or a `Task Key`, but resolve back to `Task Key` before scheduling.
2. **Pre-flight gate** — invoke or emulate `checkup` in `pre-flight` mode for the target task before scheduling any work. If the environment is broken, halt and report.
3. **Discover tasks** — scan `.mino/briefs/issue-*.md` for the target issue(s). If the user specifies an issue number, load only that brief.
4. **Resolve DAG** — build the execution graph from the brief `Dependencies` section. If the brief is stale, consult issue metadata or replay the latest valid workflow event sequence for the active approved revision.
5. **Pick next eligible node** — find the next task whose dependencies are all `done`.
6. **Executability gate** — confirm the task is `executable`, `approved`, `Approved Revision = Spec Revision`, and in `Workflow Entry State: ready_to_start`. Skip `container` tasks.
7. **Advisory file lock** — check whether target files are already claimed. Use a local advisory lock such as `.mino/locks/` when possible; if contention exists, wait or report.
8. **Update state** — before executing:
   - Increment `Attempt Count`
   - Update local brief: `Current Stage: run`, `Next Stage: verify`
   - If code changes are expected, set `Code Publication State: local_only`
   - Post a structured `run_started` event
9. **Execute** — perform the work:
   - Read target files
   - Make code changes
   - Run any necessary commands
   - Record what you did
10. **Report progress** — after each significant step, print a concise status line:
    > `[run] issue-8: Reading target files...`
    > `[run] issue-8: Applying changes to AuthService.swift...`
    > `[run] issue-8: Done. Changed: AuthService.swift, LoginView.swift`
11. **Update state** — after execution:
    - Update local brief: `Current Stage: verify`
    - Post a structured `run_completed` event indicating handoff to `verify`
12. **Hand off to verify** — when execution finishes, summarize:
    - Execution Summary
    - Changed Files (or `No File Changes`)
    - Commands Run
13. **Aggregate handoff** — when all children of a composite parent are `done`, stop execution and hand the parent to `checkup aggregate` instead of attempting `run`.
14. **Stop conditions** — halt the DAG if any node reaches:
    - `Workflow Entry State: blocked`
    - `Workflow Entry State: pending_acceptance`
    - `Current Stage: done`

## Constraints

- Do NOT execute unapproved tasks.
- Do NOT run sibling tasks in parallel.
- Do NOT bypass file lock contention.
- Do NOT ignore `Failure Context` — if retrying, attempt a different solution.
- Do NOT treat `blocked` or `pending_acceptance` as stage values.
- Do NOT increment attempts anywhere except when `run` starts.
- Keep progress lines short and scannable.

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
