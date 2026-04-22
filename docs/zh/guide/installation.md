---
title: 安装
---

# 安装

## Plugin Marketplace（推荐）

支持的 host：Copilot CLI、Claude Code、Cursor。

```bash
/plugin marketplace add robinv8/mino-skills
/plugin install mino@mino-skills
/plugin update mino@mino-skills    # 后续升级
```

## 备选 —— vercel-labs CLI

如果你的 agent host 尚不支持 `/plugin marketplace`，使用此备选方案。

[`skills`](https://github.com/vercel-labs/skills) CLI 通过一条命令将技能安装到 45+ 款 AI 工具中。

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

**其他工具**

```bash
# 列出所有支持的 agents
npx skills add --help

# 同时安装到多个 agents
npx skills add robinv8/mino-skills -a claude-code -a codex -y
```

**选项**

| Flag | 含义 |
|------|---------|
| `-g, --global` | 安装到用户目录（在所有项目中可用）。省略则为项目级安装 |
| `-a, --agent` | 指定目标 agent(s)。省略则自动检测 |
| `-y, --yes` | 跳过确认提示 |
| `--copy` | 复制文件而非创建符号链接 |

## 手动安装

```bash
cd your-project

# Claude Code
mkdir -p .claude/skills
git clone https://github.com/robinv8/mino-skills.git .claude/skills/mino

# Cursor / Codex / OpenCode（共享 `.agents/skills/` 路径）
mkdir -p .agents/skills
git clone https://github.com/robinv8/mino-skills.git .agents/skills/mino
```

任何 [Agent Skills](https://agentskills.io) 兼容的 agent 都会自动发现它们：

| 工具 | 使用方法 |
|------|-----------|
| **Claude Code** | `/mino-task feature.md`, `/mino-run issue-8` |
| **Cursor** | 在对话中提及 `-task` 或 `-run` |
| **GitHub Copilot** | Agent 根据上下文自动选择技能 |
| **Goose** | 自动从 `.agents/skills/` 加载技能 |
| **Gemini CLI** | 从本地 skills 目录加载 |
| **OpenCode** | 从工作区自动发现 |

## 直接使用（任意 agent）

无需工具 —— 直接复制 prompt：

```bash
cat skills/mino-task/SKILL.md
# 粘贴到 ChatGPT、Claude、Cursor 或任意 AI 对话中
```

## 要求

- Agent Skills 兼容的 agent（Claude Code、Cursor、Copilot、Goose、Gemini CLI 等）
- `gh` CLI，用于创建 GitHub issues
- `.mino/briefs/` 目录（首次使用时自动创建，仅本地使用，不提交）
