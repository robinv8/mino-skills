---
title: 快速上手
---

# 快速上手

## 1. 编写需求文档

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

## 2. 录入 (`task`)

```
/task feature.md
```

`task` 读取文档、分类、提取 DAG、计算带版本的 task graph，并在创建任何 issues 或 briefs 前**请求你的审批**。生成的 `.mino/briefs/` 文件是本地工作流缓存，不应被提交。

## 3. 执行 (`run`)

```
/run issue-8
```

`run` 从 DAG 中选取下一个符合条件的 task，解析 canonical `Task Key`，递增 attempt counter，执行修改，然后移交验证。

## 4. 验证 (`verify`)

由 `run` 自动触发，或直接调用：

```
/verify issue-8
```

运行 build、tests、linters。结果：

- **pass** → 推进到 `checkup`
- **retryable** → 将 `Failure Context` 反馈给 `run`（最多 3 次 retry）
- **terminal** → 阻断 task
- **manual acceptance** → 停止等待人工审阅，然后继续执行 `/mino-checkup accept issue-8`

## 5. Reconcile (`checkup`)

```
/checkup reconcile
/checkup accept issue-8
/checkup aggregate issue-1
```

`checkup` 处理 pre-flight 检查、brief reconciliation、记录 manual acceptance、聚合 composite parents，并在 task 到达 `done` 前打印集中式的 `Pending Acceptance` 列表。

## Loop Mode (v0.6.0+)

Loop Mode 是 `/mino-task` 的默认行为。审批后，orchestrator 自动驱动 `run` → `verify` → `checkup` 直到触发 halt 条件：

- `approval-required`
- `pending_acceptance`
- `fail_terminal`
- `blocked`
- `reapproval_required`
- `loop_budget_exhausted`

恢复被 halt 的 loop：

```
/mino-task resume <loop_id>
```

Stepwise opt-out：直接调用 `/mino-run`、`/mino-verify`、`/mino-checkup`。
