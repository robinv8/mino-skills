---
title: 接管既有 Issue
---

# 接管既有 Issue

如果你的仓库在本协议之前已存在，并且已经有开放的 issue，你可以在不重写历史的情况下将其标准化。

## 用法

```bash
/task adopt issue-12
```

这会生成与原生 `/mino-task PRD.md` 流程相同的 brief、events 和 label 集合。

## 接管规则

| 条件 | 结果 |
|---|---|
| Issue 为 OPEN | 继续 |
| Issue 为 CLOSED | 拒绝，返回错误信息 |
| Issue 有 ≥ 3 个开放 checkbox | Composite — 拒绝，打上 `iron-tree:needs-breakdown` 标签 |

## Label 集合

| Label | 含义 |
|---|---|
| `iron-tree:adopted` | 已纳入 Iron Tree 工作流（永久） |
| `iron-tree:needs-breakdown` | 复合 issue — 拆分后再接管 |
| `stage:task` | 等待审批 |
| `stage:run` | 已审批，准备或正在执行 run |
| `stage:verify` | run 已提交，等待 verify |
| `stage:done` | verify 已通过 |

Pull Request 不会被接管；它们像往常一样继续合并到 issue 上。

## 重新接管

如果已接管的 issue 在 GitHub 上被编辑了标题或 body，重新接管会归档之前的链：

- `.mino/archive/issue-{N}-rev-{OLD}/brief.md` —— 原始 brief
- `.mino/archive/issue-{N}-rev-{OLD}/events/` —— 原始 events
- 新的 brief 带有不同的 `Spec Revision`
- `task_re_adopted` event 包含 `previous_revision` 和 `archive_path`

## Brief 质量（v1.11+）

adopt 生成的 brief 是 **结构化的**，不是原文照搬。agent 提取：

- 可测试的 acceptance criteria
- 推导出的 verification steps
- 推断的目标文件

这与 `/mino-task PRD.md` 生成的 brief 在字段填充模式上保持一致。GitHub 上的 issue body **永远不会被修改**；标准化内容仅保存在 `.mino/briefs/issue-{N}.md` 中。

## 高质量接管

对于带有 reproduction steps 和 stack traces 的 bug issues，agent：

- 设置 `type: bug`
- 生成 ≥ 3 个可测试 checklist items
- 从 stack traces 列出目标文件（排除 JDK 内部类）
- 纳入细化 scope 的 maintainer 评论
- 忽略未获认可的噪音评论

对于模糊的功能请求，agent 会 halt 并在 `Open Questions / Warnings` 列出具体 gaps，要求人工解决后再继续。
