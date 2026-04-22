---
title: Manual Acceptance
---

# Manual Acceptance

并非每个 task 都能完全自动化。当 `verify` 无法仅通过 build/test/lint 证明正确性时，它会将 verdict 路由到 `pending_acceptance` 并停止等待人工审阅。

## 触发条件

`verify` 在以下情况进入 `pending_acceptance`：

- brief 没有带显式命令的 `Verification` section，且自动检测未找到任何工具
- 存在没有自动化覆盖的 acceptance criteria（例如"UI 感觉流畅"）
- 操作者明确要求人工签字

## 流程

```
verify ──▶ pending_acceptance ──▶ /checkup accept issue-N ──▶ finalize ──▶ done
```

## 记录 Acceptance

```
/checkup accept issue-8
```

`accept` 执行以下步骤：

1. 确认 `Workflow Entry State: pending_acceptance`
2. **先发布任何仅本地的代码** —— stage、commit、push（如有未提交的更改）
3. 渲染 `Acceptance Summary`，包含 reviewer 名字、时间戳、code ref 和 notes
4. 更新 brief metadata：
   - `Current Stage: checkup`
   - `Next Stage: done`
   - `Pass/Fail Outcome: pass`
   - `Completion Basis: accepted`
5. Emit `checkup_accept_recorded` event（silent）
6. 移除 `pending-acceptance` label
7. 进入 **Finalize**

## 可选 Note

自 v0.6.3 起，可以在 issue ref 后附加自由文本：

```
/mino-checkup accept #1842 已在 staging 用 1000 并发用户验证
```

此 note 被逐字追加到 brief 的 `Manual Acceptance` section 下 `### Accept Note` 子节，并被 agent 用作 reply dispatch 的主要决策输入。

## Pending Acceptance 列表

`/checkup reconcile` 会输出 `Pending Acceptance` 子节，列出所有仍在等待人工审阅的 tasks 及下一步所需的人工操作。这让你对所有被阻塞的工作有一个集中视图。
