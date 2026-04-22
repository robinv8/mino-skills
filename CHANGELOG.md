# Changelog

## v0.6.2 — Slim-Comment Cleanup + Commit Auto-Link

**Bug fixes (v0.5.2 debt paid)**

- `mino-verify` audible comments no longer dump rendered event yaml. The
  per-outcome operative instructions in SKILL.md (Steps 6.A–6.E) had
  contradicted the L102 constraint since v0.5.2; v0.6.0 and v0.6.1
  inherited the bug. Now every audible comment renders from a dedicated
  `comment-verify-*.md.tmpl` (markdown only, no yaml).
- `mino-checkup` `checkup_done` comment no longer inlines every local
  event yml in sequence order. The L196 constraint was a leftover from
  v0.5.0 aggregate behavior that v0.5.2 forgot to delete.
- `Commit:` auto-link is now wired into all audible comments when a sha
  is bound (the second half of v0.5.2's promise).

**New invariant in protocol** (v1.13 additive, no header bump): § Slim
Comment Invariant explicitly enumerates what audible GitHub comments may
and may not contain. Skills enforce it via dedicated `comment-*.md.tmpl`
files. The local `.mino/events/issue-{N}/` tree is reaffirmed as the sole
authoritative structured log.

**No retroactive cleanup.** Existing GitHub issue comments authored under
v0.5.2 / v0.6.0 / v0.6.1 are left untouched.

**No breaking changes.** Local event yml schemas unchanged. Brief schemas
unchanged. No new event types.

## v0.6.1 — Verification Report Artifact

**Highlights**

- New `.mino/reports/issue-{N}/report.md` artifact authored during `verify`,
  capturing human-readable evidence: environment, steps tested, findings,
  configuration recipes.
- Optional **promotion** of the report into the project's docs tree
  (`docs/integrations/<slug>.md` by default), as a separate commit pushed
  alongside `verify_passed`. Controlled by `.mino/config.yml > report.promotion`
  (`auto` | `never` | `always`).
- `verify_*` events gain optional `report_path` and `promoted_doc` fields
  (backward compatible — absence tolerated).
- `mino-checkup finalize` close-out comment surfaces the promoted doc link
  when present.
- Protocol header stays at **v1.13** (additive change).

**No breaking changes.** Existing `.mino/events/...` files without the new
fields continue to validate.

**Multi-agent git hygiene**: promotion is always a separate commit, never
an amend. Push remains forward-only (no `--force`).

## v0.6.0 (2026-04-22)

**Loop Mode by default.** Protocol bumped to **v1.13**. **BREAKING** behavior change.

### Highlights

- `/mino-task` is now the workflow's autonomous orchestrator. After approval, it drives `run` → `verify` → `checkup finalize` for every in-scope task without further human invocation, until one of the 7 protocol halt conditions fires.
- New natural-language entry: `/mino-task PRD.md`, `/mino-task #123`, `/mino-task 修一下 #45 和 #47`, `/mino-task 前十条 issue`, `/mino-task 把所有 OPEN 的都跑完` all parse into a frozen task set + Loop authorization prompt.
- New `/mino-task resume <loop_id>` sub-command for explicit halt resolution: `continue` / `skip <task_key>` (with dependency cascade) / `cancel`.
- New `.mino/loops/{loop_id}.yml` Loop Entity holds authoritative goal, frozen task set, budget, status, halt reason. Loop-level events live at `.mino/loops/{loop_id}/events/`.
- New repo-level lease `.mino/loops/active.lock` prevents concurrent Loops and stepwise interference. Stale leases are auto-detected (PID + 6h heartbeat) and cleaned on takeover.
- New event types: `loop_started`, `loop_halted`, `loop_resumed`, `loop_completed`, `loop_cancelled`. Schema: `loop:` block (vs `iron_tree:` for issue events).
- `mino-checkup`'s `finalize` sub-mode (already implemented in source) is now formally part of the Decision Function step 4 path. Comment-template inconsistency from v0.5.2 fixed in passing.

### Halt semantics (read this if upgrading)

Halts stop the **entire** Loop. Loop Mode never auto-skips an offending task. Skipping is a human act via `/mino-task resume <loop_id> skip <task_key>`. Skipped tasks recursively cancel their in-loop dependents.

### Stepwise opt-out (no breakage)

`/mino-run #N`, `/mino-verify #N`, `/mino-checkup ...` continue to work exactly as before when invoked directly. They detect orchestrator mode by the presence of `.mino/loops/active.lock` and switch to silent return only when an orchestrator holds the lease — direct invocation without an active Loop is unchanged.

### Compatibility

- Issue event schema (`iron_tree:`): unchanged.
- Brief schema: `Halt Reason` field already existed (v1.10); no new brief fields. (The earlier draft proposal to add `Loop Goal` to briefs was rejected during design — Loop state lives in `.mino/loops/`.)
- Commit message format: unchanged from v0.5.2 (`[run] #N: ...`).
- Slash command names: unchanged from v0.5.1.
- Existing Loops do not exist (this is the first Loop release), so no migration needed.

### Documentation

- `skills/references/iron-tree-protocol.md` v1.13: new § Loop Entity, new § Halt Semantics block, updated § Invariants. Removed the obsolete "finalize not yet implemented" caveat.
- `skills/references/workflow-state-contract.md`: registered 5 loop_* event types with schemas. Clarified `Halt Reason` is a brief-side mirror; Loop Entity is authoritative.
- `skills/mino-task/SKILL.md`: new § Intent Resolution, § Loop Driver, § Resume Mode at top. Existing Adopt Mode + native PRD flow kept as callable subroutines.

## v0.5.2 (2026-04-22)

GitHub-output policy change. **No schema changes.** Protocol bumped to **v1.12**.

- **Comment hygiene** — all audible GitHub issue comments are now pure human-readable notifications (heading + `Reason:` + `Action:`). Removed the `Local events: .mino/events/issue-{N}/` pointer line and the inline rendered `iron_tree:` YAML block from every audible comment template.
- **No more consolidated done comment** — `checkup_done` no longer posts the multi-block "consolidated summary" that inlined every event YAML in `sequence` order. The done comment is now a four-line notice (heading + Completion Basis + Code Ref + Code Publication State). Recovery from a lost local log via GitHub is no longer supported; the local `.mino/events/issue-{N}/*.yml` is the sole authoritative record. (Reconcile keeps a legacy fallback for issues completed under v1.10–v1.11.)
- **Commit messages link the issue** — commit message format changed from `[run] issue-{N}: …` to `[run] #{N}: …`. GitHub now auto-creates a "mentioned this issue in commit X" event on the issue timeline. **No `Closes`/`Fixes`/`Resolves` keyword is used** — `mino-checkup` retains exclusive ownership of the `done` transition.
- Updated `skills/mino-checkup/templates/comment-checkup-summary.md.tmpl`, four audible-comment specs in `skills/mino-verify/SKILL.md`, and the reconcile audible specs in `skills/mino-checkup/SKILL.md`.
- Updated `README.md`, `README.zh.md`, `skills/references/iron-tree-protocol.md`, `skills/references/workflow-state-contract.md` to describe the new policy and the v1.12 protocol boundary.
- No event log schema, brief schema, or `iron_tree:` field changes. Existing v0.5.1 briefs and event logs remain valid.

## v0.5.1 (2026-04-22)

- **Renamed all 4 skills with `mino-` prefix** to prevent slash-command collisions in shared host palettes:
  - `task` → `mino-task` (slash command: `/mino-task`)
  - `run` → `mino-run` (slash command: `/mino-run`)
  - `verify` → `mino-verify` (slash command: `/mino-verify`)
  - `checkup` → `mino-checkup` (slash command: `/mino-checkup`)
- Renamed `skills/{task,run,verify,checkup}/` directories accordingly and updated SKILL.md `name:` frontmatter.
- Updated all documentation (README, TEST_PLAN, protocol references, SKILL.md cross-refs) to use the new slash command names.
- **Protocol field names unchanged**: commit prefixes (`[task]`, `[run]`, `[verify]`, `[checkup]`), event types (`task_published`, `run_*`, `verify_*`, `checkup_*`), and stage labels (`stage:task`, `stage:run`, `stage:verify`, `stage:done`) keep their bare names — they identify protocol roles, not user-facing commands.
- Existing v0.5.0 users: run `claude plugin update mino@mino-skills` (or `copilot plugin update mino`) and use the new `/mino-*` commands.

## v0.5.0 (2026-04-22)

- Plugin marketplace support for Copilot CLI, Claude Code, and Cursor.
- Added `plugin.json`, `.claude-plugin/plugin.json`, `.cursor-plugin/plugin.json`, and `.claude-plugin/marketplace.json`.
- Retained backward compatibility with vercel-labs `npx skills add` installation path.
- Restructured README install section: Plugin Marketplace commands listed first, vercel-labs CLI as alternative fallback.
