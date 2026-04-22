---
layout: home

hero:
  name: Mino Skills
  text: 任务驱动开发
  tagline: 将一份 Markdown spec 转化为已执行、已验证的代码 —— 无论使用哪款 AI agent。
  image:
    src: /logo.png
    alt: Mino Skills
  actions:
    - theme: brand
      text: 开始使用
      link: /zh/guide/installation
    - theme: alt
      text: 查看 GitHub
      link: https://github.com/robinv8/mino-skills

features:
  - title: 四项技能，一个协议
    details: task → run → verify → checkup。Iron Tree Protocol v1.13 管理每一次状态转换。
  - title: 不挑 Agent
    details: 兼容 Claude Code、Cursor、Copilot、Goose、Gemini CLI 及任何 Agent Skills 兼容的 host。
  - title: 模板驱动的产物
    details: 所有 events、briefs 和 issue bodies 均从固定模板渲染 —— 不同 agent 输出完全一致。
  - title: Loop Mode (v0.6.0)
    details: 只需审批一次，orchestrator 自动驱动 run/verify/checkup 直到完成或中断。
  - title: 已验证 E2E
    details: 28/28 回归测试通过，覆盖 happy path、不完美现实和协议边界场景。
  - title: 默认静默
    details: GitHub 评论仅用于打断。本地 .mino/events/ 是唯一真相源。
---

## 这是什么？

一套包含四个工程技能的实现，遵循 **Iron Tree Protocol**：一个规范化的工作流，将 Markdown 需求文档一路推进到执行、验证、按需验收、复合聚合与 reconciliation。

```
Markdown spec → /mino-task → DAG approval → /mino-run → /mino-verify → /mino-checkup → done
```

没有 GUI。没有运行时。没有 deposition events。只有 agent 遵循的 prompts。

## Iron Tree Protocol

> **名字由来** — `Iron`（铁）指协议提供的铁律保证：不可篡改的 event log、确定性状态机、幂等 publish、天然 audit trail；`Tree`（树）指每个工作流落地的数据结构 —— 由 composite parent 与子 task 通过 `depends_on` 连成的 DAG。

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

## 技能

| 技能 | 用途 |
|-------|---------|
| **task** | 读取 Markdown 文档，提取 task DAG，请求审批，创建 issues + 本地 briefs |
| **run** | 串行执行已批准的 DAG，根据 prior verification failures 自我修正 |
| **verify** | Build、test、lint。Pass/fail 并附带可操作的上下文 |
| **checkup** | 环境检查、brief reconciliation、manual acceptance、composite aggregation |

## 快速链接

- [安装](/zh/guide/installation) — 几分钟内上手
- [快速上手](/zh/guide/quickstart) — 写一份 spec 并驱动它到 done
- [Iron Tree Protocol](/zh/reference/iron-tree-protocol) — 完整执行循环规范
- [更新日志](/zh/migration/changelog) — 发布历史和迁移说明
