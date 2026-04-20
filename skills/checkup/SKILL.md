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
   - `.mino/` and `.mino/briefs/` exist
   - Source adapter (GitHub CLI `gh`) is authenticated
   - Claude Code is available
3. **Skill ecosystem scan** (all modes):
   - Discover all installed skills across scopes:
     - Project: `.claude/skills/`, `.agents/skills/`
     - Global: `~/.claude/skills/`, `~/.agents/skills/`
   - Verify **core skills** (required for Iron Tree): `task`, `run`, `verify`, `checkup`
   - Check **complementary skills** by role (quality boosters, optional):
     | Role | Skill examples | Why it helps |
     |------|---------------|--------------|
     | Design validation | `think` | Pressure-test architecture before building |
     | Code review | `check` | Review diff, flag hard stops before merge |
     | Debugging | `hunt` | Systematic root-cause analysis |
     | Planning | `writing-plans` | Break specs into tracer-bullet plans |
     | Test-driven | `test-driven-development` | Write tests before implementation |
     | Sub-agent | `subagent-driven-development` | Parallel execution of independent tasks |
   - Report: which roles are covered, which are gaps
4. **Pre-flight** (if requested with task):
   - Check task-specific dependencies (e.g., `node_modules`, Xcode project)
   - Auto-repair minor issues (e.g., `npm install`)
5. **Reconcile** (if `reconcile` or `repair`):
   - List local briefs: `ls .mino/briefs/issue-*.md`
   - List source tasks: `gh issue list --state all`
   - Detect gaps: missing briefs, orphan briefs, stale metadata
   - Refresh stale brief metadata without overwriting human content
6. **Report** — print a concise health report:

   ```
   [checkup] Health Report
   ───────────────────────
   Iron Tree core:    ✅ 4/4 (all project-level)
     task, run, verify, checkup

   Complementary:     2/6 roles covered
     ✅ think       (global)    → design validation
     ❌ check                  → recommend: npx skills add tw93/Waza -a claude-code
     ❌ hunt                   → recommend: npx skills add tw93/Waza -a claude-code
     ✅ writing-plans (project) → planning
     ❌ test-driven-development → recommend: npx skills add <source> -a claude-code
     ❌ subagent-driven-development

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

- [../references/brief-contract.md](../references/brief-contract.md)
- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
