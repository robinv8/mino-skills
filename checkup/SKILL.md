---
name: checkup
description: |
  Diagnose and repair workflow health: environment readiness, skill wiring,
  brief freshness, and source-task reconciliation. Use before running tasks
  or when things feel off. Supports check, repair, reconcile, and pre-flight modes.
---

# Workflow Health Mechanic

Diagnose environment readiness, skill wiring, brief freshness, and source-task alignment. Repair what you can safely, report what you cannot.

## Modes

The user may specify a mode. Default is `check`.

- `pre-flight` — validate environment for a specific task before execution
- `check` — inspect health, report gaps, do not mutate
- `repair` — fix missing wiring and auto-repairable issues
- `reconcile` — compare local briefs against source tasks, detect drift

## Workflow

1. **Detect mode** — from user input or default to `check`.
2. **Core checks** (all modes):
   - `.agents/skills/` exists and contains `task`, `run`, `verify`, `checkup`
   - `.agents/third-party-skills.json` is valid JSON
   - `.mino/` and `.mino/briefs/` exist
   - Source adapter (GitHub CLI `gh`) is authenticated
   - Claude Code is available
3. **Pre-flight** (if requested with task):
   - Check task-specific dependencies (e.g., `node_modules`, Xcode project)
   - Auto-repair minor issues (e.g., `npm install`)
4. **Reconcile** (if `reconcile` or `repair`):
   - List local briefs: `ls .mino/briefs/issue-*.md`
   - List source tasks: `gh issue list --state all`
   - Detect gaps: missing briefs, orphan briefs, stale metadata
   - Refresh stale brief metadata without overwriting human content
5. **Report** — print a concise health report:

   ```
   [checkup] Health Report
   ───────────────────────
   Skills wired:      ✅ 4/4
   Briefs directory:  ✅
   gh auth:           ✅

   Briefs: 7 active
   ─ issue-8  ✅ synced
   ─ issue-9  ⚠️  stale (updated 3 days ago)
   ─ issue-10 ❌ orphan (source issue closed)

   Auto-fixed: 2
   Blocked:    0
   ```

## Constraints

- Do NOT create or close work items.
- Do NOT overwrite meaningful human-authored content automatically.
- Do NOT bypass pre-flight failures — if the environment is broken, block execution.

## References

- [references/brief-contract.md](references/brief-contract.md)
- [references/workflow-state-contract.md](references/workflow-state-contract.md)
