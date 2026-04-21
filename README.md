**English** | [简体中文](README.zh.md)

# Mino Skills

[![Version](https://img.shields.io/badge/release-v0.3.0-brightgreen)](https://github.com/robinv8/mino-skills/releases/tag/v0.3.0)
[![Protocol](https://img.shields.io/badge/Iron%20Tree%20Protocol-v1.10-blue)](skills/references/iron-tree-protocol.md)
[![Validated](https://img.shields.io/badge/E2E-28%2F28-brightgreen)](reports/phase2-regression-report.md)
[![Agent Skills](https://img.shields.io/badge/Agent%20Skills-Compatible-blue)](https://agentskills.io)

[Agent Skills](https://agentskills.io) compatible skill pack for task-driven development.

Turn a Markdown spec into executed, verified code — regardless of which AI agent you use.

## What is this?

A set of four engineering skills that implement the **Iron Tree Protocol**: an opinionated workflow for taking a Markdown requirement document all the way through execution, verification, acceptance when needed, composite aggregation, and reconciliation.

```
Markdown spec → /task → DAG approval → /run → /verify → /checkup → done
```

No GUI. No runtime. No deposition events. Just prompts that agents follow — every artifact (issue body, brief section, YAML event) is rendered from a fixed template under `skills/<skill>/templates/`, so different agents produce byte-identical output.

## Validated scenarios

`v0.1.0` ships with the following end-to-end coverage on `https://github.com/robinv8/mino-skills`:

| Phase | Scope | Result |
|-------|-------|--------|
| **Phase 1 — Happy path** | spec → DAG → run → verify → checkup → done (TC-1.1 / 1.2 / 1.2b) | 14 / 14 ✅ |
| **Phase 2 — Imperfect reality** | retry, dirty tree, publication failure, manual acceptance, composite aggregate, brief rebuild, sequence-gap reconcile (TC-2.1 ~ 6.2) | 14 / 14 ✅ |
| **Phase 3 — Protocol decisions** | external close, parallel run, mid-verify code drift (TC-7.1 / 7.2 / 7.3) | resolved & written back to protocol |

Full regression evidence: [reports/phase2-regression-report.md](reports/phase2-regression-report.md).

## Skills

| Skill | Purpose |
|-------|---------|
| **task** | Read a Markdown doc, extract a task DAG, ask for approval, create issues + local briefs |
| **run** | Execute an approved DAG serially, self-correct from prior verification failures |
| **verify** | Build, test, lint. Pass/fail with actionable context |
| **checkup** | Environment check, brief reconciliation, manual acceptance, composite aggregation |

## Structure

```
mino-skills/
├── skills/
│   ├── task/
│   │   ├── SKILL.md                     # Markdown → DAG → issues + briefs
│   │   └── templates/                   # brief / issue body / task_published event
│   ├── run/
│   │   ├── SKILL.md                     # Serial execution + commit + run.lock
│   │   └── templates/                   # 3 events + execution-summary section + lock
│   ├── verify/
│   │   ├── SKILL.md                     # Build/test/lint with Verify Anchor SHA
│   │   └── templates/                   # 5 outcome events + 3 brief sections
│   ├── checkup/
│   │   ├── SKILL.md                     # 7 modes incl. pre-flight & finalize
│   │   └── templates/                   # 7 events + 4 brief sections
│   └── references/
│       ├── iron-tree-protocol.md        # Execution loop specification (v1.8)
│       ├── workflow-state-contract.md   # Stage vocabulary + event whitelist
│       └── brief-contract.md            # Brief format (17 sections)
├── reports/
│   └── phase2-regression-report.md      # E2E validation evidence
├── README.md
└── LICENSE
```

## Install

### Using `skills` CLI (recommended)

The [`skills`](https://github.com/vercel-labs/skills) CLI installs skills to 45+ AI tools with one command.

Run inside your project directory:

**Claude Code**

```bash
npx skills add robinv8/mino-skills -a claude-code -y
```

**Codex**

```bash
npx skills add robinv8/mino-skills -a codex -y
```

**Cursor**

```bash
npx skills add robinv8/mino-skills -a cursor -y
```

**Other tools**

```bash
# List all supported agents
npx skills add --help

# Install to multiple agents at once
npx skills add robinv8/mino-skills -a claude-code -a codex -y
```

**Options**

| Flag | Meaning |
|------|---------|
| `-g, --global` | Install to user directory (available in all projects). Omit for project-level install |
| `-a, --agent` | Target agent(s). Omit to auto-detect |
| `-y, --yes` | Skip confirmation prompts |
| `--copy` | Copy files instead of symlinks |

### Manual install

```bash
cd your-project

# Claude Code
mkdir -p .claude/skills
git clone https://github.com/robinv8/mino-skills.git .claude/skills/mino

# Cursor / Codex / OpenCode (shared `.agents/skills/` path)
mkdir -p .agents/skills
git clone https://github.com/robinv8/mino-skills.git .agents/skills/mino
```

Any [Agent Skills](https://agentskills.io)-compatible agent will auto-discover them:

| Tool | How to use |
|------|-----------|
| **Claude Code** | `/task feature.md`, `/run issue-8` |
| **Cursor** | Mention `@task` or `@run` in chat |
| **GitHub Copilot** | Agent picks skills automatically based on context |
| **Goose** | Skills loaded automatically from `.agents/skills/` |
| **Gemini CLI** | Loaded from local skills directory |
| **OpenCode** | Auto-discovered from workspace |

### Direct use (any agent)

No tool required — just copy the prompt:

```bash
cat skills/task/SKILL.md
# Paste into ChatGPT, Claude, Cursor, or any AI chat
```

### Adopt Existing Issues

If your repo predates this protocol and already has open issues, standardize them:

```bash
/task adopt issue-12
```

This produces the same brief, events, and label set as a native task. Composite issues (3+ open checkboxes) are refused with the `iron-tree:needs-breakdown` label so you can split them first. Closed issues are refused.

Workflow position is mirrored on GitHub via labels:

| Label | Meaning |
|---|---|
| `iron-tree:adopted` | Under Iron Tree workflow (permanent) |
| `iron-tree:needs-breakdown` | Composite — split before adopting |
| `stage:task` | Awaiting approval |
| `stage:run` | Approved, ready or executing run |
| `stage:verify` | Run committed, awaiting verify |
| `stage:done` | Verify passed |

Pull Requests are not adopted; they continue to merge against issues as usual.

### Comment Policy (v1.10)

Iron Tree Protocol v1.10 treats local `.mino/events/issue-N/*.yml` as the single source of truth. GitHub issue comments are a notification channel, not an event log:

- Routine successful transitions (adopt, run, verify pass) are **silent** — no comment.
- Halts / failures that require human action still post an immediate comment so you see the signal.
- On completion, `/checkup done` posts **one consolidated summary comment** inlining every event yml, so the workflow can be replayed from GitHub alone if the local log is lost.

Pre-v1.10 issues (per-event comments) continue to be readable by `/checkup reconcile` as a fallback source.

## Update

`mino-skills` follows the standard Agent Skills convention: `main` is always the released branch, and the `skills` CLI pulls the latest content on demand.

```bash
# In your project — refresh ALL installed skills to the latest main
npx skills update -y

# Inspect what is currently installed
npx skills list
```

> Heads-up: the CLI tracks each skill by its **skill name** (`task`, `run`, `verify`, `checkup`), not by its source repo. Running `npx skills update robinv8/mino-skills` therefore reports "no installed skills found" — use the bare `npx skills update -y` form instead.

The update overwrites only the files inside `.claude/skills/<name>/` or `.agents/skills/<name>/`. Your local workflow data — `.mino/briefs/`, `.mino/run.lock`, GitHub issues, and event comments — is untouched.

Protocol upgrades are designed to be backward compatible at minor versions: state machines that are mid-flight when you update keep working without migration. Major-version bumps (e.g., `v2.0`) will document any required action at the top of this README.

## Usage

### 1. Write a requirement doc

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

### 2. Intake (`task`)

```
/task feature.md
```

`task` reads the doc, classifies it, extracts a DAG, computes a revisioned task graph, and **asks for your approval** before creating any issues or briefs. The generated `.mino/briefs/` files are local workflow cache and should not be committed.

### 3. Execute (`run`)

```
/run issue-8
```

`run` picks the next eligible task from the DAG, resolves the canonical `Task Key`, increments the attempt counter, makes changes, and hands off to verification.

### 4. Verify (`verify`)

Triggered automatically by `run`, or call directly:

```
/verify issue-8
```

Runs build, tests, linters. Results:
- ✅ **pass** → advances to `checkup`
- ❌ **retryable** → feeds `Failure Context` back to `run` (max 3 retries)
- 🚫 **terminal** → blocks the task
- ⏸️ **manual acceptance** → stops for human review, then continue with `/checkup accept issue-8`

### 5. Reconcile (`checkup`)

```
/checkup reconcile
/checkup accept issue-8
/checkup aggregate issue-1
```

`checkup` handles pre-flight checks, brief reconciliation, recording manual acceptance, aggregating composite parents, and printing a centralized `Pending Acceptance` list before a task can reach `done`.

## The Iron Tree Protocol

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   task      │────▶│    run      │────▶│   verify    │
│  define     │     │  execute    │     │   validate  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                          ┌────────────────────┼────────────────────┐
                          │                    │                    │
                          ▼                    ▼                    ▼
                    ┌──────────┐       ┌──────────┐       ┌──────────┐
                    │   pass   │       │ retryable│       │ terminal │
                    │  checkup │       │   run    │       │  blocked │
                    └──────────┘       └──────────┘       └──────────┘
                          │
                          ▼
                    ┌──────────┐
                    │   done   │
                    └──────────┘
```

- **Self-correction**: `verify` failures feed `Failure Context` back to `run` for a different approach
- **Serial execution**: DAG nodes run one at a time (v1), respecting `depends_on`. Enforced by `.mino/run.lock` (file lock, V3 will revisit parallel runs).
- **Approval gates**: Human must approve the DAG before any execution begins
- **Manual acceptance**: if automation cannot prove correctness, `verify` stops at `pending_acceptance` and `checkup accept` records the human decision
- **Shared visibility**: detailed manual checklists stay in local briefs, while issue labels/comments make acceptance status visible to collaborators
- **Revision-aware approval**: published work stays executable only while `Spec Revision` matches `Approved Revision`
- **Canonical identity**: `Task Key` is the protocol identity; `issue-8` is a user-facing locator after publish
- **Verify Anchor SHA**: every verify result binds to the `HEAD` SHA at start, immune to mid-flight code drift
- **External-event awareness**: if an issue is closed outside the workflow, `checkup reconcile` records an `External Event` section instead of silently syncing to `done`
- **Template-driven artifacts**: all events, briefs, and issue bodies are rendered from `templates/*.tmpl` files — agent-agnostic and diff-stable

## References

- [skills/references/iron-tree-protocol.md](skills/references/iron-tree-protocol.md) — the execution loop
- [skills/references/workflow-state-contract.md](skills/references/workflow-state-contract.md) — stage vocabulary
- [skills/references/brief-contract.md](skills/references/brief-contract.md) — brief format

## Requirements

- Agent Skills compatible agent (Claude Code, Cursor, Copilot, Goose, Gemini CLI, etc.)
- `gh` CLI for GitHub issue creation
- `.mino/briefs/` directory (created automatically on first use, local-only and not committed)

## Related

- [Agent Skills specification](https://agentskills.io/specification)
- [skills CLI](https://github.com/vercel-labs/skills) — install skills to 45+ AI tools
- [Mino](https://github.com/robinv8/Mino) — the macOS GUI app this skill set was extracted from

## Uninstall

### Using `skills` CLI

```bash
# Remove from current project
npx skills remove robinv8/mino-skills

# Remove globally (from all projects)
npx skills remove robinv8/mino-skills -g

# Remove from a specific agent globally
npx skills remove robinv8/mino-skills -a claude-code -g
```

### Manual uninstall

```bash
# Project-level
rm -rf .claude/skills/mino
rm -rf .agents/skills/mino

# Or globally
rm -rf ~/.claude/skills/mino
rm -rf ~/.agents/skills/mino
```

## Contributing

Contributions welcome. Workflow:

1. **Fork** this repository
2. **Edit** a `SKILL.md` or add a new skill
3. **Open a PR** describing the motivation and impact

For questions, contact hello@robinren.me.

## License

MIT
