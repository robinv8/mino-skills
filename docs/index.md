---
layout: home

hero:
  name: Mino Skills
  text: From Markdown to Merged
  tagline: Four skills. One protocol. Drive a Markdown spec to verified code — in your terminal.
  image:
    src: /logo.png
    alt: Mino Skills
  actions:
    - theme: brand
      text: Get Started
      link: /guide/installation
    - theme: alt
      text: View on GitHub
      link: https://github.com/robinv8/mino-skills

features:
  - icon: 🪨
    title: task → run → verify → checkup
    details: Four prompts compose the entire pipeline. Each one owns a stage of the Iron Tree Protocol; together they take a spec from intake to done.
  - icon: 🔌
    title: Agent-agnostic
    details: Runs on Claude Code, Cursor, Copilot CLI, Goose, Gemini CLI, and any Agent Skills compatible host. Switch agents mid-task without losing state.
  - icon: ♾️
    title: Loop Mode
    details: Approve once. The orchestrator drives run/verify/checkup autonomously across an entire DAG until done — or halts cleanly on a defined boundary.
  - icon: 📜
    title: Local-first event log
    details: Every transition is an immutable YAML event in .mino/events/. GitHub comments stay silent unless a human truly needs to be interrupted.
---

## Try it in 30 seconds

```bash
$ /mino-task specs/login.md
Resolved 3 task(s) into Loop 2026-04-22-2300-a3f9c2.
  1. #142  Add login form         add-login-form  [published]
  2. #143  Wire JWT issuer        wire-jwt-issuer [published]
  3. #144  Persist refresh token  persist-refresh [published]
Approve and start Loop? (yes / edit / cancel)
> yes

Loop 2026-04-22-2300-a3f9c2 started; driving 3 task(s).
→ /mino-run    add-login-form     ✓ committed (a4f1c2d)
→ /mino-verify add-login-form     ✓ pass
→ /mino-run    wire-jwt-issuer    ✓ committed (b8e2d31)
→ /mino-verify wire-jwt-issuer    ✓ pass
→ /mino-run    persist-refresh    ✓ committed (c1d8f49)
→ /mino-verify persist-refresh    ✓ pass

Loop 2026-04-22-2300-a3f9c2 completed: 3 task(s) done in 6 transition(s).
```

No GUI. No runtime. No background daemon. Just prompts your agent already knows how to read.

## The Four Skills

| Skill | What it does |
|---|---|
| [**mino-task**](/skills/task) | Read a spec or adopt existing issues; produce a task DAG; orchestrate Loop Mode |
| [**mino-run**](/skills/run) | Implement one approved task; commit the change |
| [**mino-verify**](/skills/verify) | Build, test, lint; report pass / retryable / terminal with context |
| [**mino-checkup**](/skills/checkup) | Reconcile briefs, accept manual verifications, aggregate composite parents |

## Why "Iron Tree"

> **Iron** — iron-clad guarantees: an immutable event log, deterministic state machine, idempotent publish, audit trail by construction. **Tree** — every workflow is a DAG of composite parents and child tasks linked by `depends_on`.

Read the [full protocol specification](/reference/iron-tree-protocol) for the state machine, halt conditions, and contract details.
