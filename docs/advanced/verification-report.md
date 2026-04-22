---
title: Verification Report Artifact
---

# Verification Report Artifact

Since v0.6.1, `mino-verify` authors a human-readable evidence report at `.mino/reports/issue-{N}/report.md` and optionally promotes it to the project's docs tree.

## Report Content

The report captures:

- **Environment**: language toolchain, framework versions, key dependencies
- **Commands tested**: the resolved verification command list and their outputs
- **Acceptance criteria walk**: which criteria were covered by automation, which require manual acceptance
- **Configuration recipes**: steps necessary to make verification pass in this environment
- **Recurring gotchas**: patterns worth documenting for future maintainers

## Optional Note

Since v0.6.3, you can append a free-form note to `/mino-verify`:

```
/mino-verify #1842 在小红书场景下需要 ...
```

This note is appended verbatim under `## Manual Verifier Note` in the report. The agent synthesises from it during reply dispatch; the verbatim copy is the audit source of truth.

## Promotion

Controlled by `.mino/config.yml`:

```yaml
report:
  promotion: auto   # auto | always | never
  docs_path: docs/integrations/
```

| Setting | Behavior |
|---|---|
| `never` | Skip promotion |
| `always` | Always promote |
| `auto` | Apply the heuristic from protocol § Verification Report ("Promotion Heuristic"). When in doubt, do NOT promote. |

On promote, the report is copied to `docs/integrations/{slug}.md` as a **separate commit** pushed alongside `verify_passed`. This keeps the promotion commit distinct from the run commit for clean history.

## Event Fields

`verify_*` events gain optional fields (backward compatible):

- `report_path: .mino/reports/issue-{N}/report.md`
- `promoted_doc: docs/integrations/{slug}.md` (when promotion occurs)

`mino-checkup finalize` surfaces the promoted doc link in its close-out comment when present.

## When Reports Are Skipped

Reports are **not** authored for:

- `verify_failed_retryable` — context lives in brief `Failure Context`
- `verify_publication_failed` — context lives in brief `Failure Context`

For `verify_passed`, `verify_failed_terminal`, and `verify_pending_acceptance`, a report is written when there is substantive evidence to record. Trivial one-line test passes may set `report_path = null`.
