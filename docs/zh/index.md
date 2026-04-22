---
layout: home

hero:
  name: Mino Skills
  text: 从需求文档到代码合并
  tagline: 四个 skill。一套协议。不挑 agent。在终端里把一份 Markdown 一路推到已验证的代码。
  actions:
    - theme: brand
      text: 开始使用
      link: /zh/guide/installation
    - theme: alt
      text: 查看 GitHub
      link: https://github.com/robinv8/mino-skills

features:
  - icon: 🪨
    title: task → run → verify → checkup
    details: 四个 prompt 组成完整流水线。每个负责 Iron Tree Protocol 的一个阶段；合起来把需求一路推到 done。
  - icon: 🔌
    title: 不挑 Agent
    details: 兼容 Claude Code、Cursor、Copilot CLI、Goose、Gemini CLI 及任何 Agent Skills 兼容的 host。任务进行中切换 agent 也不丢状态。
  - icon: ♾️
    title: Loop Mode
    details: 一次审批，orchestrator 自动驱动 run/verify/checkup 跑完整个 DAG，直到 done 或清晰地停在边界上。
  - icon: 📜
    title: 本地优先的事件日志
    details: 每次状态转换都是 .mino/events/ 下不可变的 YAML 事件。GitHub 评论默认静默，只在真正需要人介入时才发声。
---

## 30 秒体验

```bash
$ /mino-task specs/login.md
Resolved 3 task(s) into Loop 2026-04-22-2300-a3f9c2.
  1. #142  Add login form         add-login-form  [published]
  2. #143  Wire JWT issuer        wire-jwt-issuer [published]
  3. #144  Persist refresh token  persist-refresh [published]
Approve and start Loop? (yes / edit / cancel)
> yes

Loop 2026-04-22-2300-a3f9c2 started; driving 3 task(s).
→ /mino-run    add-login-form     ✓ committed (a4f1c2d)
→ /mino-verify add-login-form     ✓ pass
→ /mino-run    wire-jwt-issuer    ✓ committed (b8e2d31)
→ /mino-verify wire-jwt-issuer    ✓ pass
→ /mino-run    persist-refresh    ✓ committed (c1d8f49)
→ /mino-verify persist-refresh    ✓ pass

Loop 2026-04-22-2300-a3f9c2 completed: 3 task(s) done in 6 transition(s).
```

没有 GUI。没有运行时。没有后台守护进程。只是 agent 本来就会读的 prompt。

## 四个技能

| 技能 | 职责 |
|---|---|
| [**mino-task**](/zh/skills/task) | 读取 spec 或 adopt 已有 issue；产出 task DAG；编排 Loop Mode |
| [**mino-run**](/zh/skills/run) | 实现一个已批准的 task；提交改动 |
| [**mino-verify**](/zh/skills/verify) | 构建、测试、lint；产出 pass / retryable / terminal 并附上下文 |
| [**mino-checkup**](/zh/skills/checkup) | reconcile briefs、人工验收、composite 父任务聚合 |

## 为什么叫 "Iron Tree"

> **Iron**（铁）—— 协议提供的铁律：不可篡改的 event log、确定性状态机、幂等 publish、天然 audit trail。**Tree**（树）—— 每个工作流落地的数据结构：由 composite parent 与子 task 通过 `depends_on` 连成的 DAG。

完整的状态机、halt 条件与契约细节见 [协议规范](/zh/reference/iron-tree-protocol)。
