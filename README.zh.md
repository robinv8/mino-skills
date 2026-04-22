[English](README.md) | **简体中文**

# Mino Skills

[![Version](https://img.shields.io/badge/release-v0.6.1-brightgreen)](https://github.com/robinv8/mino-skills/releases/tag/v0.6.1)
[![Protocol](https://img.shields.io/badge/Iron%20Tree%20Protocol-v1.13-blue)](skills/references/iron-tree-protocol.md)
[![Validated](https://img.shields.io/badge/E2E-28%2F28-brightgreen)](reports/phase2-regression-report.md)
[![Agent Skills](https://img.shields.io/badge/Agent%20Skills-Compatible-blue)](https://agentskills.io)

[Agent Skills](https://agentskills.io) 兼容的技能包，用于任务驱动的开发。

将一份 Markdown spec 转化为已执行、已验证的代码 —— 无论使用哪款 AI agent。

## 这是什么？

一套包含四个工程技能的实现，遵循 **Iron Tree Protocol**：一个规范化的工作流，将 Markdown 需求文档一路推进到执行、验证、按需验收、复合聚合与 reconcile。

```
Markdown spec → /mino-task → DAG approval → /mino-run → /mino-verify → /mino-checkup → done
```

没有 GUI。没有运行时。没有 deposition events。只有 agent 遵循的 prompts —— 所有产物（issue body、brief section、YAML event）均从 `skills/<skill>/templates/` 下的固定模板渲染，因此不同 agent 能产出完全一致的输出。

## 已验证场景

`v0.1.0` 在 `https://github.com/robinv8/mino-skills` 上通过了以下端到端覆盖：

| Phase | 范围 | 结果 |
|-------|-------|--------|
| **Phase 1 — Happy path** | spec → DAG → run → verify → checkup → done (TC-1.1 / 1.2 / 1.2b) | 14 / 14 ✅ |
| **Phase 2 — 真实不完美** | retry、dirty tree、publication failure、manual acceptance、composite aggregate、brief rebuild、sequence-gap reconcile (TC-2.1 ~ 6.2) | 14 / 14 ✅ |
| **Phase 3 — 协议决策** | external close、parallel run、mid-verify code drift (TC-7.1 / 7.2 / 7.3) | 已决策并写回协议 |

完整回归证据：[reports/phase2-regression-report.zh.md](reports/phase2-regression-report.zh.md)。

## 技能

| 技能 | 用途 |
|-------|---------|
| **task** | 读取 Markdown 文档，提取 task DAG，请求审批，创建 issues + 本地 briefs |
| **run** | 串行执行已批准的 DAG，根据 prior verification failures 自我修正 |
| **verify** | Build、test、lint。Pass/fail 并附带可操作的上下文 |
| **checkup** | 环境检查、brief reconciliation、manual acceptance、composite aggregation |

## 目录结构

```
mino-skills/
├── skills/
│   ├── task/
│   │   ├── SKILL.md                     # Markdown → DAG → issues + briefs
│   │   └── templates/                   # brief / issue body / task_published event
│   ├── run/
│   │   ├── SKILL.md                     # 串行执行 + commit + run.lock
│   │   └── templates/                   # 3 个 events + execution-summary section + lock
│   ├── verify/
│   │   ├── SKILL.md                     # Build/test/lint 与 Verify Anchor SHA
│   │   └── templates/                   # 5 个 outcome events + 3 个 brief sections
│   ├── checkup/
│   │   ├── SKILL.md                     # 7 种模式，包括 pre-flight 与 finalize
│   │   └── templates/                   # 7 个 events + 4 个 brief sections
│   └── references/
│       ├── iron-tree-protocol.md        # 执行循环规范 (v1.8)
│       ├── workflow-state-contract.md   # Stage 词汇表 + event 白名单
│       └── brief-contract.md            # Brief 格式（17 个 sections）
├── reports/
│   └── phase2-regression-report.zh.md   # E2E 验证证据
├── README.zh.md
└── LICENSE
```

## 安装（Plugin Marketplace，推荐）

支持的 host：Copilot CLI、Claude Code、Cursor。

```bash
/plugin marketplace add robinv8/mino-skills
/plugin install mino@mino-skills
/plugin update mino@mino-skills    # 后续升级 —— 终于丝滑了
```

## 安装（备选 —— vercel-labs CLI）

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

### 手动安装

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

### 直接使用（任意 agent）

无需工具 —— 直接复制 prompt：

```bash
cat skills/mino-task/SKILL.md
# 粘贴到 ChatGPT、Claude、Cursor 或任意 AI 对话中
```

### 接管既有 Issue

如果你的仓库在本协议之前已存在，并且已经有开放的 issue，可以通过以下命令将其标准化：

```bash
/task adopt issue-12
```

这会生成与原生 task 相同的 brief、events 和 label 集合。复合 issue（包含 3 个及以上开放 checkbox）会被拒绝，并打上 `iron-tree:needs-breakdown` 标签，以便你先拆分它们。已关闭的 issue 也会被拒绝。

工作流位置通过 GitHub label 镜像展示：

| Label | 含义 |
|---|---|
| `iron-tree:adopted` | 已纳入 Iron Tree 工作流（永久） |
| `iron-tree:needs-breakdown` | 复合 issue — 拆分后再接管 |
| `stage:task` | 等待审批 |
| `stage:run` | 已审批，准备或正在执行 run |
| `stage:verify` | run 已提交，等待 verify |
| `stage:done` | verify 已通过 |

Pull Request 不会被接管；它们像往常一样继续合并到 issue 上。

**Brief 质量（v1.11+）**：adopt 生成的 brief 是结构化的，不是原文照搬。agent 从 issue body 和符合条件的评论中提取可测试的验收标准、推导出的验证步骤以及推断的目标文件 —— 与 `/mino-task PRD.md` 生成的 brief 在字段填充模式上保持一致。GitHub 上的 issue body 永远不会被修改；标准化内容仅保存在 `.mino/briefs/issue-{N}.md` 中。

### Comment 策略（v1.10）

Iron Tree Protocol v1.10 将本地 `.mino/events/issue-N/*.yml` 视为唯一真相源。GitHub issue comment 是通知渠道，不是事件日志：

- 常规成功过渡（adopt、run、verify pass）**静默** —— 不发 comment。
- 需要人工介入的 halt / 失败仍会立即发 comment，让你看到信号。
- 完成时，`/mino-checkup done` 发**一条简短完成通知**（标题 + Completion Basis + Code Ref + Code Publication State），不再内联 event log。本地 `.mino/events/issue-{N}/` 目录是唯一权威记录；如需持久化请自行备份。

v1.10 之前的问题（per-event comment）继续可被 `/mino-checkup reconcile` 作为 fallback 来源读取。v1.10–v1.11 完成的 issue 上仍残留 legacy 的 inline-YAML 完成 comment；本地日志丢失时 `/mino-checkup reconcile` 会回退到该格式解析。v1.12+ 的完成 comment 不再含 YAML。

## 更新

`mino-skills` 遵循标准 Agent Skills 约定：`main` 始终是已发布分支，`skills` CLI 按需拉取最新内容。

```bash
# 在你的项目里 —— 把项目内所有已安装 skill 升级到 main 最新
npx skills update -y

# 查看当前已安装的 skill
npx skills list
```

> 注意：CLI 用 **skill 名字**（`task` / `run` / `verify` / `checkup`）追踪，而不是源仓库。所以 `npx skills update robinv8/mino-skills` 会报 "no installed skills found"，使用裸的 `npx skills update -y` 即可。

更新只会覆盖 `.claude/skills/<name>/` 或 `.agents/skills/<name>/` 内的文件。你的本地工作流数据 —— `.mino/briefs/`、`.mino/run.lock`、GitHub issue 和 event comment —— 完全不会被动到。

协议小版本升级保证向后兼容：升级时正在运行的状态机不需要迁移。大版本升级（如 `v2.0`）如有破坏性变更，会在 README 顶部写明迁移步骤。

> v0.6.1 — **Verification Report Artifact。**`mino-verify` 现在在 `.mino/reports/issue-{N}/report.md` 生成可读验证报告，并可选择将其提升为 `docs/integrations/<slug>.md` 的独立提交。`verify_*` 事件新增可选的 `report_path` 和 `promoted_doc` 字段。协议保持在 v1.13。
> v0.6.0 — **Loop Mode 现已成为 /mino-task 的默认行为。** Approve 后，orchestrator 自动驱动 run/verify/checkup，直到触发 halt 条件（approval-required, pending_acceptance, fail_terminal, blocked, reapproval_required, loop_budget_exhausted）。新增 /mino-task resume <loop_id> 子命令用于显式 halt 处置。新增 .mino/loops/ 目录存放 Loop entity + repo 级 lease。Stepwise opt-out：直接调用 /mino-run、/mino-verify、/mino-checkup。**BREAKING** 行为变更。
> v0.5.2 — GitHub comment 精简为纯人类通知（不再内联 YAML / `Local events:` 路径）；commit 改用 `[run] #{N}` 以便 GitHub 在 issue 时间线自动关联。Protocol 升至 v1.12（策略变更，schema 不变）。
> v0.5.1 — slash command 重命名为 `/mino-task`、`/mino-run`、`/mino-verify`、`/mino-checkup`，避免 palette 撞名。
> v0.5.0 — 支持 Plugin marketplace（`/plugin install mino@mino-skills`）。

## 使用

### 1. 编写需求文档

```bash
cat > feature.md << 'EOF'
# 添加深色模式

## Acceptance Criteria
- [ ] 设置中有切换开关
- [ ] 跨启动持久化
- [ ] 默认尊重系统偏好

## Target Files
- SettingsView.swift
- Theme.swift
EOF
```

### 2. 录入 (`task`)

```
/task feature.md
```

`task` 读取文档、分类、提取 DAG、计算带版本的 task graph，并在创建任何 issues 或 briefs 前**请求你的审批**。生成的 `.mino/briefs/` 文件是本地工作流缓存，不应被提交。

### 3. 执行 (`run`)

```
/run issue-8
```

`run` 从 DAG 中选取下一个符合条件的 task，解析 canonical `Task Key`，递增 attempt counter，执行修改，然后移交验证。

### 4. 验证 (`verify`)

由 `run` 自动触发，或直接调用：

```
/verify issue-8
```

运行 build、tests、linters。结果：
- ✅ **pass** → 推进到 `checkup`
- ❌ **retryable** → 将 `Failure Context` 反馈给 `run`（最多 3 次 retry）
- 🚫 **terminal** → 阻断 task
- ⏸️ **manual acceptance** → 停止等待人工审阅，然后继续执行 `/mino-checkup accept issue-8`

### 5. Reconcile (`checkup`)

```
/checkup reconcile
/checkup accept issue-8
/checkup aggregate issue-1
```

`checkup` 处理 pre-flight 检查、brief reconciliation、记录 manual acceptance、聚合 composite parents，并在 task 到达 `done` 前打印集中式的 `Pending Acceptance` 列表。

## Iron Tree Protocol

> **名字由来** — `Iron`（铁）指协议提供的铁律保证：不可篡改的 event log、确定性状态机、幂等 publish、天然 audit trail；`Tree`（树）指每个工作流落地的数据结构 —— 由 composite parent 与子 task 通过 `depends_on` 连成的 DAG。**与"铁树开花"无关**；协议靠工程不靠奇迹。

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

- **Self-correction**：`verify` 失败将 `Failure Context` 反馈给 `run` 以尝试不同方案
- **Serial execution**：DAG nodes 一次执行一个（v1），遵循 `depends_on`。由 `.mino/run.lock` 强制执行（V3 将重新审视并行 run）
- **Approval gates**：人类必须在任何执行开始前审批 DAG
- **Manual acceptance**：如果自动化无法证明正确性，`verify` 停在 `pending_acceptance`，`checkup accept` 记录人工决策
- **Shared visibility**：详细的 manual checklists 保留在本地 briefs 中，而 issue labels/comments 让协作者可见 acceptance 状态
- **Revision-aware approval**：已发布的工作仅在 `Spec Revision` 与 `Approved Revision` 匹配时才可执行
- **Canonical identity**：`Task Key` 是协议身份标识；`issue-8` 是发布后的用户可见 locator
- **Verify Anchor SHA**：每次 verify 结果都绑定到启动时的 `HEAD` SHA，不受 mid-flight code drift 影响
- **External-event awareness**：如果 issue 在工作流外被关闭，`checkup reconcile` 会在 `External Event` section 中记录，而非静默同步到 `done`
- **Template-driven artifacts**：所有 events、briefs 和 issue bodies 均从 `templates/*.tmpl` 文件渲染 —— agent 无关且 diff 稳定

## 参考

- [skills/references/iron-tree-protocol.md](skills/references/iron-tree-protocol.md) — 执行循环
- [skills/references/workflow-state-contract.md](skills/references/workflow-state-contract.md) — stage 词汇表
- [skills/references/brief-contract.md](skills/references/brief-contract.md) — brief 格式

## 要求

- Agent Skills 兼容的 agent（Claude Code、Cursor、Copilot、Goose、Gemini CLI 等）
- `gh` CLI，用于创建 GitHub issues
- `.mino/briefs/` 目录（首次使用时自动创建，仅本地使用，不提交）

## 相关项目

- [Agent Skills 规范](https://agentskills.io/specification)
- [skills CLI](https://github.com/vercel-labs/skills) — 安装技能到 45+ 款 AI 工具
- [Mino](https://github.com/robinv8/Mino) — 本技能集从中提取的 macOS GUI 应用

## 卸载

### 使用 `skills` CLI

```bash
# 从当前项目移除
npx skills remove robinv8/mino-skills

# 全局移除（从所有项目）
npx skills remove robinv8/mino-skills -g

# 从特定 agent 全局移除
npx skills remove robinv8/mino-skills -a claude-code -g
```

### 手动卸载

```bash
# 项目级
rm -rf .claude/skills/mino
rm -rf .agents/skills/mino

# 或全局
rm -rf ~/.claude/skills/mino
rm -rf ~/.agents/skills/mino
```

## 贡献

欢迎提交技能改进。贡献流程：

1. **Fork** 本仓库
2. **修改** 相关 `SKILL.md` 或添加新技能目录
3. **提交 PR**，描述变更动机和影响范围

有疑问？开 issue 或通过 GitHub Discussions 联系。

## License

MIT
