---
title: Verification Report Artifact
---

# Verification Report Artifact

自 v0.6.1 起，`mino-verify` 在 `.mino/reports/issue-{N}/report.md` 生成人类可读的验证证据报告，并可选择将其提升到项目的文档树中。

## 报告内容

报告捕获：

- **环境**：语言工具链、框架版本、关键依赖
- **已测试的命令**：解析出的验证命令列表及其输出
- **Acceptance criteria walk**：哪些 criteria 被自动化覆盖，哪些需要 manual acceptance
- **配置 recipes**：在此环境中使验证通过所需的步骤
- **Recurring gotchas**：值得为 future maintainers 记录的模式

## 可选 Note

自 v0.6.3 起，可以在 `/mino-verify` 后附加自由文本 note：

```
/mino-verify #1842 在小红书场景下需要 ...
```

此 note 被逐字追加到报告的 `## Manual Verifier Note` 下。agent 在 reply dispatch 时从中综合；逐字副本是审计来源。

## 提升

由 `.mino/config.yml` 控制：

```yaml
report:
  promotion: auto   # auto | always | never
  docs_path: docs/integrations/
```

| 设置 | 行为 |
|---|---|
| `never` | 跳过提升 |
| `always` | 总是提升 |
| `auto` | 应用协议 § Verification Report 中的启发式规则（"Promotion Heuristic"）。不确定时，不提升。 |

提升时，报告被复制到 `docs/integrations/{slug}.md` 作为与 `verify_passed` 一起推送的**独立 commit**。这让提升 commit 与 run commit 保持分离，历史更清晰。

## 事件字段

`verify_*` 事件新增可选字段（向后兼容）：

- `report_path: .mino/reports/issue-{N}/report.md`
- `promoted_doc: docs/integrations/{slug}.md`（当提升发生时）

`mino-checkup finalize` 在存在 promoted doc 时会在 close-out comment 中展示其链接。

## 何时跳过报告

以下情况**不**生成报告：

- `verify_failed_retryable` —— context 已保存在 brief 的 `Failure Context` 中
- `verify_publication_failed` —— context 已保存在 brief 的 `Failure Context` 中

对于 `verify_passed`、`verify_failed_terminal` 和 `verify_pending_acceptance`，当有实质性证据需要记录时写入报告。 trivial 的单行测试通过可设置 `report_path = null`。
