---
layout: home

hero:
  name: Mino Skills
  text: Task-Driven Development
  tagline: Turn a Markdown spec into executed, verified code — regardless of which AI agent you use.
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
  - title: Four Skills, One Protocol
    details: task → run → verify → checkup. Iron Tree Protocol v1.13 governs every transition.
  - title: Agent Agnostic
    details: Works with Claude Code, Cursor, Copilot, Goose, Gemini CLI, and any Agent Skills-compatible host.
  - title: Template-Driven Artifacts
    details: All events, briefs, and issue bodies render from fixed templates — byte-identical across agents.
  - title: Loop Mode (v0.6.0)
    details: Approve once, let the orchestrator drive run/verify/checkup autonomously until done or halted.
  - title: Verified E2E
    details: 28/28 regression tests pass across happy path, imperfect reality, and protocol edge cases.
  - title: Silent by Default
    details: GitHub comments are interrupt-only. Local `.mino/events/` is the single source of truth.
---

## What is this?

A set of four engineering skills that implement the **Iron Tree Protocol**: an opinionated workflow for taking a Markdown requirement document all the way through execution, verification, acceptance when needed, composite aggregation, and reconciliation.

```
Markdown spec → /mino-task → DAG approval → /mino-run → /mino-verify → /mino-checkup → done
```

No GUI. No runtime. No deposition events. Just prompts that agents follow.

## The Iron Tree Protocol

> **Etymology** — `Iron` for the iron-clad guarantees the protocol enforces (immutable event log, deterministic state machine, idempotent publish, audit-trail by construction); `Tree` for the data structure every workflow takes — a DAG of composite parents and child tasks linked by `depends_on`.

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   task      │────▶│    run      │────▶│   verify    │
│  define     │     │  execute    │     │   validate  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                          ┌────────────────────┼────────────────────┐
                          │                    │                    │
                          ▼                    ▼                    ▼
                    ┌──────────┐       ┌──────────┐       ┌──────────┐
                    │   pass   │       │ retryable│       │ terminal │
                    │  checkup │       │   run    │       │  blocked │
                    └──────────┘       └──────────┘       └──────────┘
                          │
                          ▼
                    ┌──────────┐
                    │   done   │
                    └──────────┘
```

## Skills

| Skill | Purpose |
|-------|---------|
| **task** | Read a Markdown doc, extract a task DAG, ask for approval, create issues + local briefs |
| **run** | Execute an approved DAG serially, self-correct from prior verification failures |
| **verify** | Build, test, lint. Pass/fail with actionable context |
| **checkup** | Environment check, brief reconciliation, manual acceptance, composite aggregation |

## Quick Links

- [Installation](/guide/installation) — get up and running in minutes
- [Quick Start](/guide/quickstart) — write a spec and drive it to done
- [Iron Tree Protocol](/reference/iron-tree-protocol) — the full execution loop specification
- [Changelog](/migration/changelog) — release history and migration notes
