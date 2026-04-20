# Mino Skills

[Agent Skills](https://agentskills.io) compatible skill pack for task-driven development.

Turn a Markdown spec into executed, verified code — regardless of which AI agent you use.

## Skills

| Skill | Purpose |
|-------|---------|
| **task** | Read a Markdown doc, extract a task DAG, ask for approval, create issues + briefs |
| **run** | Execute an approved DAG serially, self-correct from prior failures |
| **verify** | Build, test, lint. Pass/fail with actionable context |
| **checkup** | Environment check, brief reconciliation, health report |

## Install

### Into your project

```bash
cd your-project
git clone https://github.com/your-org/mino-skills.git .agents/skills/mino
```

Any [Agent Skills](https://agentskills.io)-compatible agent will auto-discover them:
- **Claude Code**: `/task feature.md`, `/run issue-8`
- **Cursor**: Mention `@task` in chat
- **Copilot**: Agent picks skills automatically
- **Goose, Gemini CLI, OpenCode**, etc.

### Direct use (any agent)

```bash
cat task/SKILL.md   # Copy into any AI chat
```

## Usage

Write a requirement doc:

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

Then:
```
/task feature.md
```

`task` presents a DAG draft. Approve it. Then:
```
/run issue-8
```

`run` executes, `verify` validates, `checkup` reconciles.

## References

- [references/iron-tree-protocol.md](references/iron-tree-protocol.md) — the execution loop
- [references/workflow-state-contract.md](references/workflow-state-contract.md) — stage vocabulary
- [references/brief-contract.md](references/brief-contract.md) — brief format

## Requirements

- Agent Skills compatible agent (Claude Code, Cursor, Copilot, Goose, etc.)
- `gh` CLI for GitHub issue creation
- `.mino/briefs/` directory (created automatically on first use)

## License

MIT
