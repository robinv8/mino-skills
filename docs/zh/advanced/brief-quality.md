---
title: Brief 质量 (v1.11+)
---

# Brief 质量 (v1.11+)

自 protocol v1.11 起，`/mino-task adopt issue-N` 生成的 briefs 遵循与原生 `/mino-task PRD.md` briefs 相同的结构化字段填充模式。

## 提取行为

agent 解析 issue body 和符合条件的 comments 以提取：

| 字段 | 来源 |
|---|---|
| `type` | Issue labels（`bug`、`enhancement` 等） |
| `acceptance_criteria_checklist` | Reproduction steps、expected behavior、body assertions |
| `verification_steps` | Reproduction steps + expected/actual after fix |
| `target_files_list` | Body、stack traces、comments 中提到的文件路径 |
| `Open Questions / Warnings` | 需要人工解决的细节 gaps |

## 质量规则

- **仅可测试的 criteria**：每个 checklist item 必须是可验证的陈述（`throws IllegalArgumentException`），不能是感受性转述（`should work better`）
- **无占位文本**：`acceptance_criteria_checklist`、`verification_steps`、`target_files_list` 必须包含真实内容或显式的 `_(insufficient detail)_` 标记
- **尊重 scope**：纳入 maintainer 细化 scope 的评论；忽略未获认可的噪音评论
- **Stack trace 卫生**：JDK 内部 frames（`sun.reflect.*`）排除在 `target_files_list` 之外

## 处理模糊 Issue

对于模糊的功能请求（例如 `It would be great if the app supported dark mode`）：

1. `acceptance_criteria_checklist` 仅包含：`- [ ] _(insufficient detail — see Open Questions)_`
2. `Open Questions / Warnings` 列出 ≥ 3 个具体 gaps
3. 技能 **halt** 并重新提示审批，引用这些问题
4. 用户必须先解决 gaps，adoption 才能继续

## 不可变性保证

adopt 过程 **永远不会** 调用 `gh issue edit {N} --body`。GitHub 上的 issue body 保持不变。所有标准化内容仅保存在 `.mino/briefs/issue-{N}.md` 中。
