---
name: checkup
description: |
  Diagnose and repair workflow health: environment readiness, skill wiring,
  brief freshness, and source-task reconciliation. Use before running tasks
  or when things feel off. Supports check, repair, reconcile, pre-flight,
  accept, and aggregate modes.
---

# Workflow Health Mechanic

Diagnose environment readiness, skill wiring, brief freshness, source-task alignment, manual-acceptance handoff, and composite-task aggregation. Repair what you can safely, report what you cannot.

## Modes

The user may specify a mode. Default is `check`.

- `pre-flight` — validate environment for a specific task before execution
- `check` — inspect health, report gaps, do not mutate
- `repair` — fix missing wiring and auto-repairable issues
- `reconcile` — compare local briefs against source tasks, detect drift
- `accept` — record human acceptance for a task currently in `pending_acceptance`
- `aggregate` — finalize a `composite` / `container` task from completed child tasks

## Workflow

1. **Detect mode** — from user input or default to `check`.
2. **Core checks** (all modes):
   - `.mino/` and `.mino/briefs/` exist
   - Source adapter (`gh`) is authenticated
   - The current host agent can access project or global skills
3. **Skill ecosystem scan** (all modes):
   - Discover all installed skills across scopes:
     - Project: `.claude/skills/`, `.agents/skills/`
     - Global: `~/.claude/skills/`, `~/.agents/skills/`
   - Verify core skills: `task`, `run`, `verify`, `checkup`
   - Check complementary skills by role
   - Report which roles are covered and which are gaps
4. **Pre-flight** (if requested with a task):
   - Check task-specific dependencies such as `node_modules`, Xcode projects, toolchains, and local secrets
   - Auto-repair minor issues only when safe
   - If the environment is broken, set `Workflow Entry State: blocked`
   - Post a structured `checkup_preflight_blocked` event when blocking
   - Do NOT mark the task `done` during pre-flight
5. **Reconcile** (if `reconcile` or `repair`):
   - List local briefs: `ls .mino/briefs/issue-*.md`
   - List source tasks: `gh issue list --state all`
   - Detect gaps: missing briefs, orphan briefs, stale metadata
   - Refresh stale brief metadata without overwriting human content
   - Rebuild brief state by replaying valid workflow events for the active approved revision in ascending `sequence` order
   - Collect every task whose `Workflow Entry State` is `pending_acceptance`
   - For each pending task, read the brief `Manual Acceptance` section as the primary source for:
     - reason
     - checklist summary
     - next action
   - Cross-check the shared issue state:
     - the issue should carry label `pending-acceptance`
     - the issue timeline should already have the pending-acceptance summary comment
6. **Accept** (if `accept` with a task):
   - Confirm the task is in `Workflow Entry State: pending_acceptance`
   - If relevant code changes remain unpublished, publish them first:
     - Stage all changes **except** workflow-local files:
       ```bash
       git add -A -- ':!.mino/briefs/' ':!.mino/locks/'
       ```
     - Commit with `[run] issue-{N}: {concise change summary}`
     - `git push`
     - Capture the resulting `HEAD` SHA as `Code Ref`
     - If commit or push fails, do NOT record acceptance. Keep `Current Stage: verify`, `Next Stage: checkup`, `Workflow Entry State: pending_acceptance`, `Code Publication State: local_only`, leave `Pass/Fail Outcome` and `Completion Basis` unchanged, persist the publication error in `Failure Context`, post a short human-readable issue summary plus a structured `checkup_accept_publication_failed` event, and stop without advancing to `done`
   - Record concise human acceptance evidence in `Verification Summary`
   - Post a shared issue comment including:
     - `Reviewer`
     - `Result: accepted`
     - `Timestamp`
     - `Code Ref`
     - `Notes`
   - Remove the `pending-acceptance` label
   - Update brief: `Current Stage: checkup`, `Next Stage: done`, `Workflow Entry State: ready_to_start`, `Code Publication State: published|not_applicable`, `Pass/Fail Outcome: pass`
   - Record `Completion Basis: accepted` and `Code Ref`
   - Post a structured `checkup_accept_recorded` event
7. **Aggregate** (if `aggregate` with a task):
   - Confirm the task is `composite` or `container`
   - Confirm all required child task keys are `done`
   - Record aggregate completion evidence in `Verification Summary` or `Completion Handoff`, including the child task keys and issue numbers used as evidence
   - Update brief: `Current Stage: checkup`, `Next Stage: done`, `Workflow Entry State: ready_to_start`, `Code Publication State: not_applicable`, `Pass/Fail Outcome: pass`
   - Record `Completion Basis: aggregated`
   - Post a structured `checkup_aggregate_recorded` event
8. **Report** — print a concise health report.
   - The report should include a dedicated `Pending Acceptance` subsection whenever any tasks are waiting on human verification
   - Example:

   ```
   [checkup] Health Report
   ───────────────────────
   ...
   Briefs: 7 active
   ─ issue-1  ✅ done (auto-closed)
   ─ issue-2  ⏸️ pending acceptance — verify passed, awaiting manual check
   ─ issue-3  ⏸️ pending acceptance — no tooling detected
   ─ issue-4  ❌ blocked — terminal failure

   Pending Acceptance: 2
   ─ issue-2: Settings toggle
     Action: Verify in app, then run `/checkup accept issue-2`
   ─ issue-3: Dark mode fallback
     Action: Review manually, then run `/checkup accept issue-3`
   ```
9. **Finalize state** — after `reconcile`, `accept`, or `aggregate` completes for a specific task:
   - Only mark the task `done` if `Completion Basis` is `verified`, `accepted`, or `aggregated`
   - Only mark the task `done` if `Code Publication State` is `published` or `not_applicable`
   - Update local brief: `Current Stage: done`, `Next Stage: none`
   - Post a structured `checkup_done` event
   - If `Close On Done: auto`, close the issue:
     - `gh issue close {N} --reason completed`
     - Do NOT close if the issue was already closed by the user
   - If `Close On Done: manual`, post a friendly reminder and leave the issue open:
     > [checkup] issue-{N}: Task done — awaiting manual verification
     >
     > This task has been completed but requires manual verification before the issue can be closed.
     >
     > To close this issue after verification, run:
     > ```
     > gh issue close {N} --reason completed
     > ```

## Constraints

- Do NOT create issues or open closed issues.
- Do NOT overwrite meaningful human-authored content automatically.
- Do NOT bypass pre-flight failures — if the environment is broken, block execution.
- Do NOT mark tasks `done` during `pre-flight` or plain `check`.
- Do NOT record manual acceptance against an unpublished code state.
- Do NOT close issues when `Close On Done: manual`.
- Do NOT stage or commit `.mino/briefs/` or `.mino/locks/` during code publication — briefs are local workflow cache and must not enter the git history.

## References

- [../references/brief-contract.md](../references/brief-contract.md)
- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
