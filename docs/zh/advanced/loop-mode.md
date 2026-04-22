---
title: Loop Mode
---

# Loop Mode

自 v0.6.0 起，Loop Mode 是 `/mino-task` 的默认执行模式。你批准 task set 后，orchestrator 自动驱动 `run` → `verify` → `checkup finalize`，无需进一步人工调用，直到 halt 条件触发。

## 入口

`/mino-task` 接受自然语言参数并解析为冻结的 task set：

| 模式 | 动作 |
|---|---|
| `PRD.md`（文件路径） | 原生 PRD 流程 → publish → Loop，goal_kind: `set_done` |
| `#123` | 接管单个 issue → Loop，goal_kind: `task_done` |
| `#45 #47` | 接管多个 issues → Loop，goal_kind: `set_done` |
| `前 N 条 issue` / `first N` / `top N` | 标准查询 → Loop，goal_kind: `set_done` |
| `all open` / `所有 OPEN` | 所有已接管的开放 issues → Loop，goal_kind: `set_done` |
| `resume <loop_id>` | 恢复被 halt 的 Loop |

## 解析计划与审批

进入 Loop Mode 前，`/mino-task` 打印 **Resolved Plan** 并要求显式 `yes`：

```
You are authorizing Loop Mode to autonomously execute the following plan.

Loop ID:        {loop_id}
Goal:           {task_done | set_done}
Intent:         {verbatim user input}
Resolved query: {one-line summary or "n/a (file path)"}
Tasks ({len}):  budget = {budget_max_transitions} transitions
  1. #<N>  <title>  <task_key>
  2. ...
Excluded ({len}, see notes):
  - #<N>  <reason: composite / closed / not adopted / etc>

Halts on: approval-required, pending_acceptance, fail_terminal, blocked,
          reapproval_required, loop_budget_exhausted.
Stepwise opt-out: invoke /mino-run, /mino-verify, /mino-checkup directly.

Approve and start Loop? (yes / edit / cancel)
```

`yes` 是协议 § Invariants 要求的 **显式 Loop Mode  opt-in**。

## Halt 条件

Loop 在以下第一个条件触发时停止（按协议顺序评估）：

1. `approval-required` — 需要 DAG 审批
2. `pending_acceptance` — verify 无法自动证明正确性
3. `fail_terminal` — retry 预算耗尽
4. `blocked` — pre-flight 或外部事件阻断进度
5. `reapproval_required` — spec revision 与已批准版本不一致
6. `loop_budget_exhausted` — 达到最大 transitions 数（安全护栏）
7. `protocol_gap` — 不可恢复的状态不一致

Halt 会停止 **整个** Loop。Loop Mode 不会自动跳过有问题的 task。跳过是人工行为，通过 `/mino-task resume <loop_id> skip <task_key>` 执行。

## Resume 模式

```
/mino-task resume <loop_id> [continue | skip <task_key> | cancel]
```

| 子命令 | 效果 |
|---|---|
| `continue` | 重新获取 lease，emit `loop_resumed`，重新进入 Driver Iteration |
| `skip <task_key>` | 标记指定 task 为 cancelled；级联取消所有依赖它的 in-loop 任务 |
| `cancel` | 设置 status 为 `cancelled`，释放 lease，退出 |

## Loop Entity

每个 Loop 在 `.mino/loops/{loop_id}.yml` 写入权威实体：

- `goal`、`frozen_task_set`、`budget`、`status`、`halt_reason`
- `transitions` 数组：`{iso, task_key, skill, outcome}`

repo 级 lease `.mino/loops/active.lock` 防止并发 Loop。stale lease（PID 消失或心跳 > 6h）在接管时自动检测并清理。

## Stepwise Opt-Out

直接调用 `/mino-run`、`/mino-verify`、`/mino-checkup` 仍然完全按原有方式工作。这些技能通过 `.mino/loops/active.lock` 检测 orchestrator 模式，仅在 orchestrator 持有 lease 时切换为 silent return。
