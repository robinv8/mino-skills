---
title: Installation
---

# Installation

## Plugin Marketplace (Recommended)

Supported hosts: Copilot CLI, Claude Code, Cursor.

```bash
/plugin marketplace add robinv8/mino-skills
/plugin install mino@mino-skills
/plugin update mino@mino-skills    # later upgrades
```

## Alternative — vercel-labs CLI

If your agent host doesn't yet support `/plugin marketplace`, use this fallback.

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

## Manual Install

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
| **Claude Code** | `/mino-task feature.md`, `/mino-run issue-8` |
| **Cursor** | Mention `-task` or `-run` in chat |
| **GitHub Copilot** | Agent picks skills automatically based on context |
| **Goose** | Skills loaded automatically from `.agents/skills/` |
| **Gemini CLI** | Loaded from local skills directory |
| **OpenCode** | Auto-discovered from workspace |

## Direct Use (Any Agent)

No tool required — just copy the prompt:

```bash
cat skills/mino-task/SKILL.md
# Paste into ChatGPT, Claude, Cursor, or any AI chat
```

## Requirements

- Agent Skills compatible agent (Claude Code, Cursor, Copilot, Goose, Gemini CLI, etc.)
- `gh` CLI for GitHub issue creation
- `.mino/briefs/` directory (created automatically on first use, local-only and not committed)
