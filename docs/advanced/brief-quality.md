---
title: Brief Quality (v1.11+)
---

# Brief Quality (v1.11+)

Since protocol v1.11, briefs produced by `/mino-task adopt issue-N` follow the same structured field-filling pattern as native `/mino-task PRD.md` briefs.

## Extraction Behavior

The agent parses the issue body and qualifying comments to extract:

| Field | Source |
|---|---|
| `type` | Issue labels (`bug`, `enhancement`, etc.) |
| `acceptance_criteria_checklist` | Reproduction steps, expected behavior, body assertions |
| `verification_steps` | Reproduction steps + expected/actual after fix |
| `target_files_list` | File paths mentioned in body, stack traces, comments |
| `Open Questions / Warnings` | Gaps in detail that need human resolution |

## Quality Rules

- **Testable criteria only**: every checklist item must be a verifiable statement (`throws IllegalArgumentException`), not a feeling paraphrase (`should work better`)
- **No placeholder text**: `acceptance_criteria_checklist`, `verification_steps`, and `target_files_list` must contain real content or an explicit `_(insufficient detail)_` marker
- **Scope-respecting**: maintainer comments that refine scope are incorporated; unendorsed noise comments are ignored
- **Stack trace hygiene**: JDK internal frames (`sun.reflect.*`) are excluded from `target_files_list`

## Handling Vague Issues

For vague feature requests (e.g., `It would be great if the app supported dark mode`):

1. `acceptance_criteria_checklist` contains exactly: `- [ ] _(insufficient detail — see Open Questions)_`
2. `Open Questions / Warnings` lists ≥ 3 specific gaps
3. The skill **halts** and re-prompts for approval, citing the questions
4. The user must resolve the gaps before the adoption proceeds

## Immutability Guarantee

The adopt process **never** calls `gh issue edit {N} --body`. The issue body on GitHub remains unchanged. All standardization lives in `.mino/briefs/issue-{N}.md` only.
