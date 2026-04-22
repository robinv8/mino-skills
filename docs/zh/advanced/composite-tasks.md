---
title: Composite Tasks
---

# Composite Tasks

在 Iron Tree Protocol 中，task 分为三种类型：

| 分类 | 行为 |
|---|---|
| `atomic` | 可执行的叶子 task。有 target files 和 acceptance criteria。 |
| `composite` | 父 task，分解为子 tasks。不可直接执行。 |
| `container` | 类似于 composite；作为分组父节点。不可直接执行。 |

## DAG 结构

composite parent 及其 children 通过 `depends_on` 连成 DAG：

```
┌─────────────────┐
│   composite-1   │
│  (container)    │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐ ┌───────┐
│ task-A│ │ task-B│
│atomic │ │atomic │
└───┬───┘ └───┬───┘
    │         │
    └────┬────┘
         ▼
    ┌─────────┐
    │  task-C │
    │ atomic  │
    └─────────┘
```

## 执行规则

- `run` 跳过 `container` 和 `composite` tasks；它们分解而非执行。
- 子 tasks 必须全部达到 `done`，parent 才能被 aggregate。
- `checkup aggregate <issue>` 在所有 required children 为 `done` 后 finalize composite parent。

## Aggregation

`aggregate` 执行以下步骤：

1. 确认 `Classification` 为 `composite` 或 `container`
2. 从 `Work Breakdown` 和 `Dependencies` 解析 required child task keys
3. 确认每个 child 的 `Pass/Fail Outcome: pass` 且 `Current Stage: done`
4. 用 aggregate evidence list 替换 `Verification Summary`
5. 设置 `Completion Basis: aggregated`，`Code Publication State: not_applicable`
6. Emit `checkup_aggregate_recorded` event（silent）
7. 进入 **Finalize**

## 接管防护

接管现有 issues 时，composite issues（≥ 3 个开放 checkbox）会被 **拒绝** 并打上 `iron-tree:needs-breakdown` 标签。必须先拆分为子 issues，再逐个接管。
