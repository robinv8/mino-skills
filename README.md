# Mino Skills

[![Agent Skills](https://img.shields.io/badge/Agent%20Skills-Compatible-blue)](https://agentskills.io)

[Agent Skills](https://agentskills.io) compatible skill pack for task-driven development.

Turn a Markdown spec into executed, verified code — regardless of which AI agent you use.

## What is this?

A set of four engineering skills that implement the **Iron Tree Protocol**: an opinionated workflow for taking a Markdown requirement document all the way through execution, verification, acceptance when needed, composite aggregation, and reconciliation.

```
Markdown spec → /task → DAG approval → /run → /verify → /checkup → done
```

No GUI. No runtime. No deposition events. Just prompts that agents follow.

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
│   ├── task/SKILL.md                    # Markdown → DAG → issues + briefs
│   ├── run/SKILL.md                     # Serial execution with self-correction
│   ├── verify/SKILL.md                  # Build/test/lint validation
│   ├── checkup/SKILL.md                 # Health check + reconciliation
│   └── references/
│       ├── iron-tree-protocol.md        # Execution loop specification
│       ├── workflow-state-contract.md   # Stage vocabulary
│       └── brief-contract.md            # Brief format
├── README.md
└── LICENSE
```

## Install

### Using `skills` CLI (recommended)

The [`skills`](https://github.com/vercel-labs/skills) CLI installs skills to 45+ AI tools with one command.

在项目目录下运行：

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

`checkup` handles pre-flight checks, brief reconciliation, recording manual acceptance, and aggregating composite parents before a task can reach `done`.

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
- **Serial execution**: DAG nodes run one at a time (v1), respecting `depends_on`
- **Approval gates**: Human must approve the DAG before any execution begins
- **Manual acceptance**: if automation cannot prove correctness, `verify` stops at `pending_acceptance` and `checkup accept` records the human decision
- **Revision-aware approval**: published work stays executable only while `Spec Revision` matches `Approved Revision`
- **Canonical identity**: `Task Key` is the protocol identity; `issue-8` is a user-facing locator after publish

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

## License

MIT
