---
title: Quick Start
---

# Quick Start

## 1. Write a Requirement Doc

```bash
cat > feature.md << 'EOF'
# Add dark mode

## Acceptance Criteria
- [ ] Toggle in settings
- [ ] Persists across launches
- [ ] Respects system preference by default

## Target Files
- SettingsView.swift
- Theme.swift
EOF
```

## 2. Intake (`task`)

```
/task feature.md
```

`task` reads the doc, classifies it, extracts a DAG, computes a revisioned task graph, and **asks for your approval** before creating any issues or briefs. The generated `.mino/briefs/` files are local workflow cache and should not be committed.

## 3. Execute (`run`)

```
/run issue-8
```

`run` picks the next eligible task from the DAG, resolves the canonical `Task Key`, increments the attempt counter, makes changes, and hands off to verification.

## 4. Verify (`verify`)

Triggered automatically by `run`, or call directly:

```
/verify issue-8
```

Runs build, tests, linters. Results:

- **pass** → advances to `checkup`
- **retryable** → feeds `Failure Context` back to `run` (max 3 retries)
- **terminal** → blocks the task
- **manual acceptance** → stops for human review, then continue with `/mino-checkup accept issue-8`

## 5. Reconcile (`checkup`)

```
/checkup reconcile
/checkup accept issue-8
/checkup aggregate issue-1
```

`checkup` handles pre-flight checks, brief reconciliation, recording manual acceptance, aggregating composite parents, and printing a centralized `Pending Acceptance` list before a task can reach `done`.

## Loop Mode (v0.6.0+)

Loop Mode is the default for `/mino-task`. After approval, the orchestrator drives `run` → `verify` → `checkup` automatically for every in-scope task until a halt condition fires:

- `approval-required`
- `pending_acceptance`
- `fail_terminal`
- `blocked`
- `reapproval_required`
- `loop_budget_exhausted`

Resume a halted loop:

```
/mino-task resume <loop_id>
```

Stepwise opt-out: invoke `/mino-run`, `/mino-verify`, `/mino-checkup` directly.
