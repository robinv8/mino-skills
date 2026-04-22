# Iron Tree Protocol

> Version: 1.13
> Purpose: Define the recursive, low-touch execution engine.

## Concept

A human-approved requirement document unfolds into a fully implemented, verified, accepted when necessary, aggregated when composite, and reconciled feature without further human steering unless the workflow is blocked.

The protocol shifts work from ad hoc prompting to a state-machine driven loop with explicit gates, durable metadata, revision-aware approval, and deterministic recovery.

### Why "Iron Tree"

- **Iron** — the protocol's guarantees are iron-clad by construction: an append-only local event log is the single source of truth, every transition is recorded as a templated YAML event, `Task Key` and `Spec Revision` are deterministic, and `verify` results bind to a `Verify Anchor SHA` so mid-flight code drift cannot silently invalidate them.
- **Tree** — every workflow materialises as a DAG: composite parents decompose into children linked by `depends_on`, `checkup aggregate` rolls completion upward from leaves to root, and a single requirement document compiles to a tree of issues.

The name is descriptive of the protocol's mechanics. It is **not** a reference to the Chinese idiom *铁树开花* ("iron-tree blooming") and carries no implication of rarity or miracle — the workflow is engineered to be repeatable.

## Core Loop

### Phase 1: Document Intake (`task`)

1. Read the full Markdown document
2. Extract a DAG draft with explicit `depends_on`
3. Classify each node
4. Compute a deterministic `Task Key` per node
5. Compute a `Spec Revision` for the proposed DAG
6. Wait for human approval before publishing

### Phase 2: Publish (`task`)

1. Create or refresh source tasks idempotently using stable `Task Key`s
2. If the approved revision changed, require re-approval before refresh
3. Write initial issue metadata and local briefs
4. Initialize each task at `Current Stage: definition`
5. Record `Approved Revision = Spec Revision`

### Phase 3: Readiness Gate (`checkup pre-flight`)

1. Validate environment readiness before execution
2. Check task-specific prerequisites
3. If the runway is broken, mark the task `blocked`
4. Never transition a task to `done` during pre-flight

## Execution Lock

The protocol guarantees serial execution of `run` at the repository level. Only one `run` may be active at any time.

- Before entering Phase 4, `run` checks for `.mino/run.lock`
- If the lock exists, `run` refuses execution and reports the active task key and start time
- On successful entry, `run` writes the lock with its task key and ISO timestamp
- On normal completion or unrecoverable failure, `run` removes the lock
- A stale lock (older than a configurable threshold, default 2 hours) may be overridden with explicit confirmation

V1 executes DAG nodes serially even when they have no mutual dependencies. Parallel execution is deferred to v3 after conflict detection and file-lock advisory semantics are fully specified.

### Phase 4: Execution (`run`)

1. Check `.mino/run.lock`; if present, refuse execution
2. Acquire lock with task key and ISO timestamp
3. Rebuild the DAG from brief metadata, using issue metadata as fallback
4. Select the next atomic task whose `depends_on` are `done`
5. Assert an advisory lock on target files
6. Increment `Attempt Count`
7. Modify the codebase
8. Commit changes locally so `verify` has a stable SHA to anchor against
9. Handoff to `verify`

If the commit step fails (e.g., a pre-commit hook rejects, identity not configured), `run` does not consume retry budget: it persists the error in `Failure Context`, emits `run_commit_failed`, decrements `Attempt Count` back to the pre-run value, and halts. This mirrors `verify_publication_failed` for symmetry between the two skills' publication boundaries.

### Phase 5: Validation (`verify`)

1. Record `Verify Anchor SHA` = current HEAD commit
2. Execute explicit repo-native checks, preferring `.mino/config.yml` overrides
3. Compare observed results to acceptance criteria
4. On recoverable failure, write `Failure Context` and hand control back to `run`
5. On unrecoverable failure, mark the task `blocked`
6. If automation cannot complete verification, stop at `pending_acceptance`
7. On success, publish code first if needed, then transition to `checkup`
8. If publication fails after checks pass, stay in `verify`, preserve the local code state, record publication failure context, and retry publication instead of falsely recording success; this does not consume retry budget

## Verify Anchor

`verify` evaluates the codebase at a specific commit, not the working tree. This guarantees deterministic verification even if the user edits files during a long-running check.

- `run` must commit its changes before handoff to `verify`
- `verify` records `Verify Anchor SHA` = `HEAD` at the moment it starts
- The verification result is bound to that SHA; subsequent uncommitted changes do not invalidate the result
- If `verify` fails with `fail_retryable`, the next `run` starts from the same SHA or a newer committed state

This field is required in Loop Mode so the Decision Function can distinguish between a verify-pass at a stale SHA and a verify-pass at the current SHA.

## Verification Report

`verify` produces a human-readable artifact at
`.mino/reports/issue-{N}/report.md` whenever it has substantive evidence to
record (passed checks with a meaningful run log, terminal failure analysis,
or pending-acceptance with a real checklist). Pure retryable failures and
publication failures do not author a report — their context lives in the
brief's `Failure Context` section.

The report is rendered from `mino-verify/templates/report.md.tmpl`. Sections:
title, verify anchor SHA, outcome, environment (versions table), steps tested,
findings, configuration recipe, promotion decision.

### Promotion to Project Docs

When the report contains generally-applicable integration knowledge — config
matrix, version compatibility, env setup, recurring gotchas, framework
integration walkthrough — it MAY be promoted into the project's docs tree as
a separate commit, published in the same `git push` as `verify_passed`.

- **Default target**: `docs/integrations/{kebab-slug-of-issue-title}.md`
- **Override**: `.mino/config.yml > report.docs_path`
- **Mode**: `.mino/config.yml > report.promotion` ∈ `auto` (default) | `never` | `always`
- **When promoting and the file already exists**: append a section
  `## Update {YYYY-MM-DD} (verify_anchor: {sha7})`. Never overwrite.
- **Commit**: a separate commit titled
  `docs(issue-N): {short title} integration notes` (with the standard
  Co-authored-by trailer). Never amend the `run` commit. The subsequent
  `git push` (already part of 6.A) publishes both commits.

### Promotion Heuristic (auto mode)

Promote when ALL of:
- The findings would be useful to a future user of the project, not just
  reviewers of this specific issue.
- The configuration / version / setup advice is permanent (or stable until
  a major upstream version change).
- The content is substantial (not trivial one-liner).

Do NOT promote when ANY of:
- Bug reproduction logs / debugging traces tightly bound to this fix.
- Internal refactor notes meaningful only to maintainers.
- Speculative or experimental findings not yet validated.

When in doubt: do NOT promote. False positives pollute the repo docs tree;
false negatives only mean the report stays local — recoverable by a future
manual re-promote (out of scope for v0.6.1).

### Optional Event Fields

`verify_passed`, `verify_pending_acceptance`, and `verify_failed_terminal`
events MAY include:

- `report_path: .mino/reports/issue-{N}/report.md` — set whenever a report
  was authored (most cases).
- `promoted_doc: {docs_path}/{slug}.md` — set only when promotion occurred.
  Absent or `null` otherwise.

These fields are optional for backward compatibility with pre-v0.6.1 events.

### Phase 6: Acceptance Or Aggregation (`checkup`)

1. Record human acceptance when the task is in `pending_acceptance`
2. Or aggregate child completion for composite/container parents
3. Bind completion evidence to a published `code_ref` when code changed
4. If code publication fails during acceptance, do not record acceptance; preserve `pending_acceptance`, keep the local code state, and record publication failure context for retry
5. Reconcile the local brief with the authoritative issue record
6. Transition `checkup` → `done` only when completion evidence exists
7. Evaluate the next serially eligible DAG node

## Source Of Truth

- Stable task metadata lives in the GitHub issue body
- **Workflow transitions live in the local event log** (`.mino/events/issue-{N}/*.yml`) — this is the authoritative record
- GitHub issue comments are a notification mirror, not the primary source
- Local briefs are a cache for scheduling, inspection, and recovery
- `checkup reconcile` repairs local drift by replaying valid events from the local log for the active approved revision

## External Events

The source system (GitHub) is authoritative for issue existence and visibility, but the local event log is authoritative for workflow state and completion. When an external event contradicts workflow state, the protocol must not silently overwrite workflow state.

### Issue Closed Externally

When `checkup reconcile` discovers that an issue is closed on GitHub but has no `checkup_done` or equivalent terminal event in its workflow history:

1. Do **not** sync the brief to `done`
2. Record `External Event: issue_closed` on the brief
3. Set `Workflow Entry State: blocked`
4. Post a `checkup_reconcile_external_close_detected` event to the issue
5. Require human confirmation before the task can advance

This preserves the invariant that `done` requires recorded completion evidence, not just a closed issue tracker entry.

## Event Recording & Comment Policy

Each workflow event has two persistence targets:

1. **Local event log** — `.mino/events/issue-{N}/{sequence:04d}-{event-kebab}.yml` is the **authoritative source of truth**. Every skill that mutates workflow state MUST write this file before taking any further action. If the local write fails, the event did not happen.

2. **GitHub issue comment** — a mirror / notification channel, governed by the event's category:

   - **silent** events: do not post any comment. The local yml is the complete record.
   - **audible** events: post an independent comment with human narrative only (heading + Reason + Action). The yml block is written to the local event file and not echoed to GitHub.
   - **terminal** events (`checkup_done`): post a short completion notice (heading + Completion Basis / Code Ref / Code Publication State). The notice is a notification only; recovery from a lost local log is not supported via GitHub under v1.12.

`workflow-state-contract.md` § Event Categories lists exact assignments.

Comment posting is always best-effort and non-fatal. On failure, the skill logs `comment_post_failed: <reason>` in its own report and continues. The local yml remains correct.

`checkup reconcile` reads **local events first**, falls back to the terminal summary comment for `done` issues (legacy fallback — issues completed under protocol v1.10 or v1.11 had inline YAML in the done comment; v1.12+ done comments contain no YAML and are not parseable as a recovery source. Try this fallback only when the comment matches the pre-v1.12 signature), and only falls back to per-event comments for pre-v1.10 (legacy v1.9 or earlier) issues.

Under protocol v1.12, the only recovery source is the local event log at `.mino/events/issue-{N}/`. Operators must back up `.mino/` themselves (sync, commit to a private branch, etc.). Future protocol versions may reintroduce a separate cloud durability layer; the GitHub issue stream is intentionally not used for that purpose.

## Multi-Agent Git Hygiene

The protocol assumes the remote (origin/main and any shared branches) is **append-only from each agent's perspective**. Multiple agents may run sequentially or concurrently against the same repository, and any one of them rewriting shared history silently destroys another agent's work.

Hard rules for every skill that touches git:

1. **Never `git push --force` or `--force-with-lease`** against a branch that has been pushed by anyone else, including a previous agent run.
2. **Never `git reset --hard` to a commit older than the current `origin/<branch>` tip** when intending to push afterwards. Use `git revert` to undo published commits instead.
3. **Never rebase or amend a commit that already exists on the remote.** Local-only commits may be rebased; once pushed, they are immutable.
4. **Never delete a remote branch** unless the user explicitly asks.
5. If reconciliation requires rewriting history (e.g., secret leak), stop and ask the user — this is outside the agent's authority.

Rationale: agent collaboration without these rules degenerates into a last-writer-wins race where revert commits, `.gitignore` rules, and protocol fixes vanish without trace. The cost of `git revert` (one extra commit) is always lower than the cost of recovering lost work from a forced push.

## Adopting Existing Issues

The protocol's native entry point is `/mino-task <spec.md>`, which posts `task_published` at sequence 1. Repositories that pre-existed the protocol have issues without this event; `run` / `verify` / `checkup` therefore refuse to operate on them.

`/mino-task adopt issue-N` is the standard on-ramp. It produces the **same shape of artifacts** as native publication so downstream skills cannot tell the difference:

- A local brief at `.mino/briefs/issue-{N}.md` (same template as native)
- An event yml at `.mino/events/issue-{N}/0001-task-adopted.yml` with `event: task_adopted` and `sequence: 1`
- A pair of GitHub labels marking workflow position: `iron-tree:adopted` (permanent) + `stage:task` (mutable)

Adoption events are **silent** in v1.10; no GitHub comment is posted. The local yml is the authoritative record, and `checkup reconcile` reads it directly.

### Eligibility

`/mino-task adopt issue-N` accepts an issue iff:

1. The issue exists and is `OPEN` on the host repository
2. The issue body does not declare more than `COMPOSITE_THRESHOLD = 3` open checkboxes (`- [ ]`); composite issues must be broken into child issues by the human first
3. Any prior `iron-tree:adopted` label triggers **re-adopt** semantics, not refusal

Closed issues are refused with a clear error.

### Spec Identity for Adopted Issues

Adopted issues have no Markdown spec. The protocol synthesizes:

```
spec_path     = github://issue-{N}
spec_revision = sha256( normalize(issue_title) + "\n---\n" + normalize(issue_body) )[:8]
```

`normalize` is the same transformation used by `task` (strip trailing whitespace, CRLF→LF, collapse blank-line runs).

### Re-adoption

If `iron-tree:adopted` already exists on the issue:

1. Move the existing `.mino/briefs/issue-{N}.md` and `.mino/events/issue-{N}/` into `.mino/archive/issue-{N}-rev-{previous_spec_revision}/`
2. Compute fresh `spec_revision` from current title + body
3. Emit `task_re_adopted` (sequence 1 of the new chain) instead of `task_adopted`; the event carries `previous_revision` and `archive_path`
4. Remove any `stage:run` / `stage:verify` / `stage:done` label and apply `stage:task`
5. Approval must be re-requested from the user (same gate as a native re-approval)

### Stage Label Lifecycle

`stage:*` labels are **mutually exclusive** and mirror the active workflow stage:

| Transition | Performed by | Trigger |
|---|---|---|
| (none) → `stage:task` | `task` | `/mino-task adopt` succeeds, or native `task` publishes |
| `stage:task` → `stage:run` | `task` | User approval recorded |
| `stage:run` → `stage:verify` | `run` | Step 7 commit succeeds (or `not_applicable` path) |
| `stage:verify` → `stage:done` | `verify` | `verify_passed` recorded |

Failure paths **do not move the label**: a stuck issue's GitHub label tells the human exactly where it halted. Label sync failures (gh CLI down, missing permission) are non-fatal: the skill records a `stage_label_sync_failed` warning in its report and continues; the local yml remains authoritative.

### PR Out of Scope

Pull Requests are not adopted. PRs continue to merge against an issue; merging does not by itself transition any label.

### Brief Standardization (v1.11+)

The brief produced by adopt MUST be field-equivalent to a native publish brief. The agent treats the issue body (plus comments from the issue author or accounts with OWNER/MEMBER/COLLABORATOR association) as PRD-equivalent input and extracts structured `acceptance_criteria_checklist`, `verification_steps`, and `target_files_list` rather than copying issue body verbatim.

If extraction yields insufficient detail, the brief MUST surface gaps in `Open Questions / Warnings` and halt approval until the user resolves them — sparse briefs are allowed but must be explicitly sparse.

The issue body itself is NEVER edited by adopt. Standardization lives only in `.mino/briefs/issue-{N}.md`.

## DAG Rules

- A task cannot enter `run` until all `depends_on` tasks are `done`
- V1 executes DAG nodes serially, even sibling tasks
- Container tasks never enter execution directly
- Composite parents may complete through aggregation once all required children are `done`

## Required Capabilities

- **Predictive DAG (`task`)**: Extract full tree structure before execution
- **Revision-Aware Approval (`task`)**: Tie approval to a specific `Spec Revision`
- **Idempotent Publish (`task`)**: Refresh existing issues and briefs instead of duplicating them on rerun
- **Pre-flight (`checkup`)**: Ensure the runway is clear before takeoff
- **Serial Scheduling (`run`)**: Respect `depends_on`, order, attempt budget, and advisory file locks
- **Structured Failure (`verify`)**: Feed actionable errors back to `run`
- **Publish-Before-Pass (`verify`)**: Publish accepted code before recording a successful terminal state
- **Manual Acceptance (`verify` + `checkup`)**: Emit explicit human steps, then record acceptance against a published code ref
- **Aggregate Completion (`checkup`)**: Finalize composite parents from child completion evidence
- **Structured State Events (all skills)**: Post machine-readable workflow events so `checkup reconcile` can recover state deterministically

## Execution Modes

The Core Loop above defines **what** transitions are legal. Execution Modes define **who drives** the transitions and **when control returns to a human**.

Two modes are defined. Both operate on the same state machine and produce identical state events. Only the driver differs.

### Stepwise Mode (default)

- A human invokes one skill at a time (`/mino-task`, `/mino-run`, `/mino-verify`, `/mino-checkup`)
- Each skill performs exactly one transition and returns control to the human
- The human decides what to invoke next based on the resulting state
- This is the safe baseline; every skill must remain usable in stepwise mode

### Loop Mode

- The agent invokes skills autonomously in sequence until a `Halt Condition` is reached
- The agent does not require a human invocation between transitions
- The agent must continuously observe authoritative state (issue body and events, repaired by `checkup reconcile` when needed) before each next action; it must not act on stale local cache
- All transitions remain governed by the Core Loop and `workflow-state-contract.md`; Loop Mode adds no new transitions and weakens no gates
- Loop Mode is opt-in per goal; the default remains stepwise

### Goal

Loop Mode requires an explicit, observable goal. Two goal shapes are recognized:

- `task_done`: a specific `Task Key` (and, for composites, all required children) reaches `done`
- `set_done`: every `Task Key` in a named set reaches `done`

A goal is **observable** when its completion can be determined from issue state plus structured workflow events alone. Goals that require subjective judgement are not valid Loop Mode goals.

### Halt Conditions

The agent **must immediately halt** Loop Mode and return control to a human when any of the following are true. Halting is not failure; halting is the protocol explicitly handing the next decision to a human.

Halt conditions are evaluated **in the order listed**. The first matching condition wins; this prevents more general gates from masking more specific outcomes.

| # | Halt Reason | Trigger | Resolution |
|---|---|---|---|
| 1 | `goal_reached` | Goal predicate satisfied | None; report success |
| 2 | `fail_terminal` | Any in-scope task has `Pass/Fail Outcome: fail_terminal` | Human revises spec, code, or tooling |
| 3 | `pending_acceptance` | Any in-scope task has `Workflow Entry State: pending_acceptance` | Human runs `/mino-checkup accept` |
| 4 | `reapproval_required` | Any in-scope task has `Spec Revision != Approved Revision` | Human re-approves via `/mino-task` |
| 5 | `blocked` | Any in-scope task has `Workflow Entry State: blocked` **and** `Pass/Fail Outcome != fail_terminal` | Human investigates and resolves |
| 6 | `protocol_gap` | The agent encounters a state not covered by the Core Loop or contract | Human extends the protocol |
| 7 | `loop_budget_exhausted` | The configured maximum number of consecutive transitions is reached | Human inspects and decides whether to resume |

`fail_terminal` is evaluated before `blocked` because a terminal verification failure always also sets `Workflow Entry State: blocked`; without ordered evaluation, terminal failures would be reported under the more general `blocked` reason and lose diagnostic specificity.

When halting, the agent must:

1. Post a structured event with `event: loop_halted` and a `halt_reason` field matching one of the values above
2. Set `Halt Reason` on the affected task's brief (see `workflow-state-contract.md`)
3. Stop issuing further skill invocations until a human re-engages

### Halt Semantics

`Halt = entire Loop stops.` Loop Mode does **not** automatically skip the offending task and continue with the rest of the in-scope set, even when other tasks are unaffected. The protocol treats the halt conditions as full-loop circuit breakers because (a) the human is the only authority on whether the offending state is recoverable, ignorable, or fatal, and (b) silent per-task skipping would make resume semantics ambiguous.

A human resumes a halted Loop via `/mino-task resume <loop_id>` and chooses explicitly: `continue` (re-evaluate halt; if the underlying state changed, proceed), `skip <task_key>` (mark that task `cancelled` in the loop, recursively cancel its DAG dependents within the loop, then continue with the rest), or `cancel` (the entire Loop transitions to `cancelled`).

Skipping is a human act, never an autonomous one.

### Decision Function

Within a Loop Mode iteration, the agent selects the next action by evaluating in order:

1. If any halt condition is true → halt
2. If a task is in `Current Stage: run` with `Next Stage: verify` → invoke `/mino-verify`
3. If a task is in `Current Stage: verify` with `Pass/Fail Outcome: fail_retryable` and `Attempt Count <= Max Retry Count` → invoke `/mino-run` (which performs its own pre-flight before scheduling)
4. If a task is in `Current Stage: checkup` with `Next Stage: done` and `Completion Basis` set → invoke `/mino-checkup finalize <issue>` to bind completion evidence and transition to `done`
5. If all required children of an in-scope composite parent are `done` and the parent is not yet `done` → invoke `/mino-checkup aggregate` on the parent
6. If a DAG node has all `depends_on` `done` and `Workflow Entry State: ready_to_start` → invoke `/mino-run` on the eligible node with the lowest `sequence` of last activity (deterministic tie-break)
7. Otherwise → halt with `protocol_gap`

Notes:

- `pre-flight` is not a Decision Function step. The contract requires `run` to perform `checkup pre-flight` internally before scheduling work; Loop Mode relies on that internal call rather than scheduling pre-flight separately.
- Step 4 invokes `mino-checkup`'s `finalize` sub-mode, which performs only the finalization sub-step (bind `code_ref`, write the terminal event, transition to `done`) and never performs `accept` or `aggregate` work; those remain explicit, separately invoked modes. As of v1.13 this mode is implemented; see `skills/mino-checkup/SKILL.md` § Finalize.

The decision function is intentionally narrow. New decision branches must be added to the protocol before any skill or agent uses them in Loop Mode.

### Loop Budget

- A Loop Mode session must declare a `Max Consecutive Transitions` upper bound (default `50`)
- Each invoked skill counts as one transition regardless of outcome
- Reaching the budget halts with `loop_budget_exhausted`; this is a safety valve against runaway loops, not a verdict on the work

### Loop Entity

A Loop's authoritative state lives in `.mino/loops/{loop_id}.yml`. Briefs do **not** store the goal or the task set; they only mirror `Halt Reason` for diagnostic convenience.

`loop_id` format: `{ISO date}-{HHMM}-{6-hex random}`, e.g. `2026-04-22-1432-a3f8b2`.

Schema:

```yaml
loop_id: 2026-04-22-1432-a3f8b2
created_at: 2026-04-22T14:32:11Z
goal_kind: task_done | set_done
intent: |
  <verbatim user input that produced this loop, e.g. "前十条 issue">
resolved_query: |
  <the exact gh / file-path resolution applied at approval time, e.g.
   "gh issue list --state open --label iron-tree:adopted --limit 10 --sort created --json number,title">
task_keys:
  - <task_key_1>
  - <task_key_2>
  # ...frozen at approval, never re-queried during the loop
budget_max_transitions: <int>      # default max(50, 10 * len(task_keys))
budget_used: 0
status: running | halted | completed | cancelled
halt_reason: null                  # populated when status=halted; one of the 7 protocol values
halt_at_task_key: null             # which in-scope task triggered the halt
halt_at_iso: null
transitions:                       # append-only audit log
  - {iso, task_key, skill, outcome}
```

`status` transitions:

- `running` <- initial; `loop_started` written.
- `running` -> `halted`: any halt condition fires; `loop_halted` written; `active.lock` released.
- `halted` -> `running`: `/mino-task resume <loop_id> continue` re-acquires lock; `loop_resumed` written.
- `halted` -> `cancelled`: `/mino-task resume <loop_id> cancel`; `loop_cancelled` written; loop is dead.
- `running` -> `completed`: every `task_key` is `done`; `loop_completed` written; `active.lock` released.

Loop-level events live at `.mino/loops/{loop_id}/events/{seq:04d}-{event-kebab}.yml`. They share no sequence space with per-issue events.

### Active Lease

`.mino/loops/active.lock` enforces single-Loop semantics at repo scope.

Schema:

```yaml
loop_id: 2026-04-22-1432-a3f8b2
holder_pid: 12345                  # process that acquired the lease
holder_agent: mino-task            # always "mino-task" -- only the orchestrator holds the lease
acquired_at: 2026-04-22T14:32:11Z
heartbeat_at: 2026-04-22T14:33:00Z # refreshed at every transition
```

Acquisition rules:

1. `/mino-task` (when not in resume mode) refuses to start a new Loop if `active.lock` exists AND the holder process is alive AND `heartbeat_at` is within the last `STALE_THRESHOLD` (default 6 hours). Error: `Loop {loop_id} already active (held by PID {pid}); run /mino-task resume {loop_id} or cancel it before starting a new one.`
2. If the holder process is gone or `heartbeat_at` is older than `STALE_THRESHOLD`, the lock is **stale**. The new orchestrator may take over: write a `loop_halted` event with `halt_reason: stale_takeover` against the previous loop, mark it `halted`, then proceed.
3. Stepwise commands (`/mino-run`, `/mino-verify`, `/mino-checkup`) **never** acquire the lease. They check for it: if present and `holder_agent: mino-task`, they detect orchestrator mode and return silently after their state write. If present but stale, they refuse with `stale loop lease detected; run /mino-task to clean up`.

### Invariants

- Loop Mode never bypasses approval gates. Initial DAG approval still requires a human; only post-approval execution is automated.
- Loop Mode never weakens `pending_acceptance`. Human acceptance always remains a required state transition for the affected task.
- Loop Mode emits the same workflow events as stepwise mode. A reconciliation run cannot tell the two modes apart from event history alone, except for `loop_halted` and `loop_resumed` markers.
- A skill that is not safe in stepwise mode is not safe in Loop Mode. There is no "Loop Mode only" capability.
- A Loop holds the repo-level `.mino/loops/active.lock` lease for its entire `running` duration. The lease is released on halt, completion, or cancellation. Stepwise skill invocations check the lease and adapt their hand-off behavior (silent return when the orchestrator holds the lease) but never acquire it themselves.
- The Loop Entity (`.mino/loops/{loop_id}.yml`) is the sole source of truth for goal, task set, budget, and status. Briefs may mirror `Halt Reason` for the most recent halt of a task, but mirror status only -- the Loop Entity is authoritative.
- `/mino-task` approval is the per-goal Loop-Mode opt-in required by these invariants. The approval prompt MUST explicitly name the Loop Mode authorization and list the frozen task set; an implicit "yes" to a DAG preview without that text does not constitute valid Loop authorization.

### Slim Comment Invariant (since v0.6.2; clarifying v0.5.2 intent)

Audible GitHub comments MUST contain ONLY:

- A short human-readable heading
- `Reason:` (one line)
- `Action:` (one line, may be `—`)
- Optional `Commit: <github URL>` when a sha is bound
- Optional `Docs: <github URL>` when a doc was promoted
- Optional `Report: <github URL>` when a report was promoted

Audible GitHub comments MUST NOT contain:

- Rendered event yaml of any kind (no `iron_tree:` blocks, no fences)
- Any `.mino/*` filesystem path (those are local-only artifacts)
- Inline event-log dumps (no `--- sequence N · ... ---` dividers)
- Pointers like "Local events: `.mino/events/...`"

The local `.mino/events/issue-{N}/` tree is the authoritative structured
log. GitHub comments are the human notification channel.

Skills enforce this by rendering audible comments exclusively from
`comment-{event-kebab}.md.tmpl` files (no yaml fences inside). The
existing `event-{event-kebab}.yml.tmpl` files are reserved for writing
to `.mino/events/`.
