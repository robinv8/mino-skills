---
title: Silent Events (v1.10)
---

# Silent Events (v1.10)

Iron Tree Protocol v1.10 将本地 `.mino/events/issue-N/*.yml` 视为**唯一真相源**。GitHub issue comment 是通知渠道，不是事件日志。

## 策略

| 事件类型 | GitHub Comment |
|---|---|
| 常规成功过渡（adopt、run、verify pass） | **静默** —— 不发 comment |
| 需要人工介入的 halt / 失败 | **可听** —— 立即发 comment |
| 完成（`checkup_done`） | **一条简短完成通知** |

## 完成通知格式

done comment 包含：

- 标题
- Completion Basis
- Code Ref
- Code Publication State

不再内联 event log。本地 `.mino/events/issue-{N}/` 目录是唯一权威记录。

## 可听评论

可听评论是纯人类通知：

- 简短标题行
- `Reason:`
- `Action:`

无 `Local events:` 指针。无渲染的 YAML block。无 `.mino/*` 路径。

## Fallback 来源

当本地 event log 丢失时，`/checkup reconcile` 使用以下 fallback 链：

1. **Primary**：`.mino/events/issue-{N}/*.yml`
2. **Terminal summary**（仅 pre-v1.12）：解析 done comment 的内联 YAML blocks
3. **Legacy per-event comments**（pre-v1.10）：从单个 comments 解析每个 YAML block

v1.12+ 的 done comments 不再含 YAML，因此 v1.12+ 完成的 issues 无法使用 terminal summary fallback。

## 为什么静默？

- 减少 GitHub 通知噪音
- 保持 issue threads 对人类可读
- 防止 agent 内部信息泄漏到用户可见的 comments 中
- 让本地 event log 成为持久、权威的来源
