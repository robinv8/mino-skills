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

1. **Discover tasks** — scan `.mino/briefs/issue-*.md` for the target issue(s). If the user specifies an issue number, load only that brief.
2. **Resolve DAG** — build the execution graph from `depends_on` fields.
3. **Pick next eligible node** — find the next task whose dependencies are all `done`.
4. **Executability gate** — confirm the task is `executable` and `approved`. Skip `container` tasks.
5. **Self-correction check** — if the brief contains a `Failure Context` from a prior `verify` failure, analyze it and adjust your approach. Do NOT blindly repeat the failed attempt.
6. **Lock target files** — check if target files are being modified by another process. If contention, wait or report.
7. **Execute** — perform the work:
   - Read target files
   - Make code changes
   - Run any necessary commands
   - Record what you did
8. **Report progress** — after each significant step, print a concise status line:
   > `[run] issue-8: Reading target files...`
   > `[run] issue-8: Applying changes to AuthService.swift...`
   > `[run] issue-8: Done. Changed: AuthService.swift, LoginView.swift`
9. **Hand off to verify** — when execution finishes, summarize:
   - Execution Summary
   - Changed Files (or "No File Changes")
   - Commands Run
10. **Stop conditions** — halt the DAG if any node reaches:
    - `blocked` (unrecoverable error)
    - `pending_acceptance` (needs human review)
    - `done`

## Constraints

- Do NOT execute unapproved tasks.
- Do NOT run sibling tasks in parallel.
- Do NOT bypass file lock contention.
- Do NOT ignore `Failure Context` — if retrying, attempt a different solution.
- Keep progress lines short and scannable.

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
