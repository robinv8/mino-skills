---
title: Silent Events (v1.10)
---

# Silent Events (v1.10)

Iron Tree Protocol v1.10 treats local `.mino/events/issue-N/*.yml` as the **single source of truth**. GitHub issue comments are a notification channel, not an event log.

## Policy

| Event Type | GitHub Comment |
|---|---|
| Routine successful transitions (adopt, run, verify pass) | **Silent** — no comment |
| Halts / failures requiring human action | **Audible** — immediate comment |
| Completion (`checkup_done`) | **One short completion notice** |

## Completion Notice Format

The done comment contains:

- Heading
- Completion Basis
- Code Ref
- Code Publication State

No inline event log. The local `.mino/events/issue-{N}/` directory is the sole authoritative record.

## Audible Comments

Audible comments are pure human notifications:

- Short heading line
- `Reason:`
- `Action:`

No `Local events:` pointer. No rendered YAML block. No `.mino/*` paths.

## Fallback Sources

`/checkup reconcile` uses the following fallback chain when the local event log is missing:

1. **Primary**: `.mino/events/issue-{N}/*.yml`
2. **Terminal summary** (pre-v1.12 only): parse the done comment's inline YAML blocks
3. **Legacy per-event comments** (pre-v1.10): parse every YAML block from individual comments

v1.12+ done comments contain no YAML, so terminal summary fallback is unavailable for issues completed under v1.12+.

## Why Silent?

- Reduces GitHub notification noise
- Keeps issue threads readable for humans
- Prevents agent internals from leaking into user-facing comments
- Makes the local event log the durable, authoritative source
