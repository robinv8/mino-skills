---
name: mino-task
description: |
  Convert a local Markdown requirement document into an execution DAG,
  or act as the Loop orchestrator that drives an approved task set through
  run, verify, and checkup autonomously until a halt condition fires.
  Read a spec or bug note, extract tasks with dependencies, classify them,
  and prepare GitHub issues + local briefs. Always asks for approval
  before creating anything or entering Loop Mode. Use when starting new work
  from a Markdown doc, adopting existing issues, or resuming a halted Loop.
---

# Task Intake Engine

Read a user-specified local Markdown requirement document, extract a Directed Acyclic Graph (DAG) of work items, classify each, and prepare them for execution — but only after explicit user approval.

This skill is the protocol's entry point and, as of v0.6.0, its **Loop orchestrator**. In addition to creating issues and briefs from a Markdown spec or adopting existing issues, `/mino-task` can drive the entire workflow autonomously after explicit Loop Mode approval. Its outputs (issue body, local brief, structured event, loop entity) are consumed by `run`, `verify`, and `checkup`. Field names, section names, and event schemas are **fixed by template files** in `templates/` so downstream skills can rely on them deterministically.

## Intent Resolution

`/mino-task` accepts a free-form text argument. Parse it into one of the following intents. **Resolve once, freeze the result, then proceed**; never re-query mid-Loop.

| Pattern | Intent | Action |
|---|---|---|
| Empty argument | `none` | Print usage and halt. Do not enter Loop Mode. |
| Argument matches `resume <loop_id>[ continue\|skip <task_key>\|cancel]` | `resume` | Branch to § Resume Mode. |
| Argument matches `adopt issue-<N>` (legacy syntax) | `adopt_single` | Equivalent to `#<N>`; branch to § Adopt Mode then enter Loop with `goal_kind: task_done`. |
| Argument is an existing file path ending in `.md` | `prd` | Branch to existing native PRD flow (Steps 1–6). After publish, enter Loop with `goal_kind: set_done` over the published `task_keys`. |
| Argument contains exactly one `#<N>` token | `single_issue` | Branch to § Adopt Mode for that issue, then Loop with `task_done`. |
| Argument contains multiple `#<N>` tokens | `multi_issue` | Adopt each in given order, then Loop with `set_done`. |
| Argument matches a "前 N 条" / "first N" / "top N" pattern (any language) | `top_n` | Run the canonical query (see below). Loop with `set_done`. |
| Argument matches an "all open / 所有 OPEN / 全部" pattern | `all_open` | Run the canonical query without limit. Loop with `set_done`. |
| Anything else | `ambiguous` | Refuse with `Cannot infer intent from "{input}". Provide a path, one or more #<issue>, "前 N 条 issue", "all open", or use /mino-task adopt issue-<N>.` Do not guess. |

**Canonical query** for `top_n` and `all_open`:

```
gh issue list --state open --label iron-tree:adopted --json number,title,body,labels --sort created --order asc [--limit <N>]
```

(adopted-only; sorted by issue number ascending = oldest first)

For `top_n`, parse the integer N from the input. For `all_open`, omit `--limit`.

### Resolved Plan & Approval

After parsing, freeze the resolved task set. Print the **Resolved Plan** in this exact form:

```
You are authorizing Loop Mode to autonomously execute the following plan.

Loop ID:        {loop_id}
Goal:           {task_done | set_done}
Intent:         {verbatim user input}
Resolved query: {one-line summary or "n/a (file path)"}
Tasks ({len}):  budget = {budget_max_transitions} transitions
  1. #<N>  <title>  <task_key>
  2. ...
Excluded ({len}, see notes):
  - #<N>  <reason: composite / closed / not adopted / etc>

Halts on: approval-required, pending_acceptance, fail_terminal, blocked, reapproval_required, loop_budget_exhausted.
Stepwise opt-out: invoke /mino-run, /mino-verify, /mino-checkup directly on a single issue instead.

Approve and start Loop? (yes / edit / cancel)
```

`yes` is the **explicit Loop Mode opt-in** required by protocol § Invariants. Do NOT proceed without it. `edit` returns control to the user to refine input. `cancel` exits without acquiring the lease.

## Loop Driver

After approval and only after approval:

1. **Acquire active lease** —
   - Read `.mino/loops/active.lock` if present. If holder PID is alive AND `heartbeat_at` is within the last 6 hours, refuse with the error in protocol § Loop Entity. Do NOT take over.
   - If the lock is stale (PID gone or heartbeat older than 6 hours): write a `loop_halted` event with `halt_reason: stale_takeover` against the previous Loop, set its status to `halted`, then continue.
   - Render `templates/active-lock.yml.tmpl` with current PID and ISO timestamp; write to `.mino/loops/active.lock` atomically (write to a temp file, then rename).
2. **Write Loop Entity** — generate `loop_id` (`{ISO-date}-{HHMM}-{6-hex-random}`); render `templates/loop.yml.tmpl`; write to `.mino/loops/{loop_id}.yml`.
3. **Emit `loop_started`** — render `templates/event-loop-started.yml.tmpl`; write to `.mino/loops/{loop_id}/events/0001-loop-started.yml`.
4. **Print** `Loop {loop_id} started; driving {len(task_keys)} task(s).`

### Driver Iteration

Repeat:

1. **Heartbeat** — rewrite `active.lock` with current `heartbeat_at`.
2. **Re-read authoritative state** — for each in-scope `task_key`: read its brief, scan its `.mino/events/issue-{N}/*.yml`, optionally `gh issue view` if external close suspected. The Loop Entity, briefs, and events together are the authoritative state — local cache is never trusted standalone.
3. **Evaluate halt conditions** in protocol order (1–7). On any halt:
   - Update Loop Entity: `status: halted`, `halt_reason`, `halt_at_task_key`, `halt_at_iso`, append to `transitions`.
   - Render `templates/event-loop-halted.yml.tmpl`; write next sequence event under `.mino/loops/{loop_id}/events/`.
   - Mirror `Halt Reason` onto the affected task's brief (single line, surgical edit; the brief field already exists per contract).
   - Release the lease (`rm .mino/loops/active.lock`).
   - Print: `Loop {loop_id} halted: {halt_reason} on {task_key}. Resume with: /mino-task resume {loop_id}`.
   - Exit.
4. **Evaluate Decision Function** (protocol L305, steps 2–6 — step 1 is the halt check above, step 7 is `protocol_gap` which ALSO halts).
5. **Invoke the chosen skill in-process**: `/mino-run <issue>`, `/mino-verify <issue>`, or `/mino-checkup finalize <issue>`. The skill is loaded as a sub-task within the same agent session; you call it the same way a stepwise human would, then wait for it to return.
6. **Increment** `budget_used`. If `budget_used >= budget_max_transitions`, halt with `loop_budget_exhausted`.
7. **Append transition** to the Loop Entity `transitions:` array: `{iso, task_key, skill, outcome}`.
8. **Re-iterate.**

When all `task_keys` are `done`:
- Update Loop Entity: `status: completed`, `completed_at`.
- Render `templates/event-loop-completed.yml.tmpl`; write next sequence event.
- Release the lease.
- Print: `Loop {loop_id} completed: {N} task(s) done in {transitions_used} transition(s).`

## Resume Mode

`/mino-task resume <loop_id> [continue | skip <task_key> | cancel]`

1. Read `.mino/loops/{loop_id}.yml`. If absent, halt with `Loop {loop_id} not found.`. If `status != halted`, halt with `Loop {loop_id} status is {status}; only halted loops can be resumed.`
2. If no sub-action given: print the halt context (halt_reason, halt_at_task_key, remaining task_keys, dependency graph showing which would be cancelled if skip is chosen) and prompt for one of `continue / skip <task_key> / cancel`.
3. Branch:
   - **`continue`** — Re-acquire the lease (same rules as fresh start). Set status `running`. Render `templates/event-loop-resumed.yml.tmpl`; write next sequence event under `.mino/loops/{loop_id}/events/`. Re-enter Driver Iteration.
   - **`skip <task_key>`** — Mark the named task as cancelled within the Loop. **Cascade**: any in-scope task that lists `<task_key>` (transitively) in its `depends_on` is also cancelled. Render `templates/event-loop-cancelled.yml.tmpl` for each cancelled task with `cancellation_scope: task`. Then re-acquire lease, status `running`, emit `loop_resumed`, re-enter Driver Iteration. The cancelled tasks are excluded from future halt-condition evaluation and from the goal-reached predicate.
   - **`cancel`** — Set status `cancelled`. Render `templates/event-loop-cancelled.yml.tmpl` with `cancellation_scope: loop`. Do NOT re-acquire lease. Print: `Loop {loop_id} cancelled.`

Skipping is the **only** way a task is excluded from a Loop. The driver itself never skips.

## Adopt Mode

When the user invokes `/mino-task adopt issue-N` (instead of `/mino-task <path-to-spec.md>`), branch into Adopt Mode. This mode produces artifacts shape-compatible with native publication, then rejoins the standard flow at the approval gate.

### Adopt-Step 1: Pre-flight

1. Verify `gh auth status` succeeds; on failure halt with: `gh CLI not authenticated; run \`gh auth login\` then retry.`
2. Ensure the six standard labels exist on the host repo. For each label in the table below, run `gh label list --search "<name>" --json name -q '.[].name'`; if absent, run `gh label create "<name>" --color "<color>" --description "<desc>" --force` (idempotent).

   | Name | Color | Description |
   |---|---|---|
   | `iron-tree:adopted` | `0E8A16` | Issue is under Iron Tree workflow |
   | `iron-tree:needs-breakdown` | `D93F0B` | Composite issue — must be split into child issues before adoption |
   | `stage:task` | `FBCA04` | In task stage (awaiting approval) |
   | `stage:run` | `1D76DB` | Approved, awaiting or executing run |
   | `stage:verify` | `5319E7` | Run committed, awaiting or executing verify |
   | `stage:done` | `0E8A16` | Verify passed |

### Adopt-Step 2: Fetch & validate

```
gh issue view {N} --json number,state,title,body,labels,url
```

Refuse with explicit error message if any of these hold:

- `state != "OPEN"` → `Issue #{N} is {state}; only OPEN issues can be adopted.`
- body contains `≥ COMPOSITE_THRESHOLD (=3)` lines matching regex `^\s*-\s*\[\s\]` → composite. Run `gh issue edit {N} --add-label "iron-tree:needs-breakdown"` and halt with: `Issue #{N} appears composite ({k} open checkboxes). Split it into child issues, then run \`/mino-task adopt\` on each child.`

### Adopt-Step 3: Detect re-adopt

If the fetched labels include `iron-tree:adopted`:

1. Read existing `.mino/briefs/issue-{N}.md`, extract the `Approved Revision: <hex>` value → `previous_revision`
2. `archive_path = .mino/archive/issue-{N}-rev-{previous_revision}/`
3. `mkdir -p {archive_path}` then `mv .mino/briefs/issue-{N}.md {archive_path}/brief.md` and `mv .mino/events/issue-{N} {archive_path}/events`
4. Mark `mode = re_adopted`

Otherwise `mode = adopted`.

### Adopt-Step 4: Compute identity

```
task_key      = slugify(issue_title)            # same slugify rules as Step 3 of native flow
spec_path     = "github://issue-{N}"
spec_revision = sha256( normalize(issue_title) + "\n---\n" + normalize(issue_body) )[:8]
adopted_at    = ISO 8601 timestamp, e.g. 2026-04-21T10:54:00Z
```

### Adopt-Step 5: Approval Gate

Show the user:

```
Adopting issue #{N}: {title}
  task_key:      {task_key}
  spec_revision: {spec_revision}
  mode:          {adopted | re_adopted}
  {if re_adopted}: archived previous chain to {archive_path}

Approve adoption? (yes / cancel)
```

Halt until explicit `yes`. `cancel` rolls back archival (re-`mv` files back) when `mode = re_adopted`.

On approval (`yes`), before exiting Adopt-Step 5, run `gh issue edit {N} --remove-label "stage:task" --add-label "stage:run"`. Failures are warnings, not errors. (This mirrors native Step 5's approval-time label flip.)

### Adopt-Step 6: Standardize & render brief

Treat the issue as a PRD-equivalent input. Reuse the same extraction reasoning the native `/mino-task <spec>` flow applies to a PRD `.md` file — the resulting brief MUST be indistinguishable in field-filling pattern from a native publish.

**Inputs to consider** (in priority order):
1. Issue body (primary PRD text)
2. Issue comments authored by the issue author or by accounts with `OWNER`, `MEMBER`, or `COLLABORATOR` author_association — these often carry clarifications, scope tightening, or accepted solutions. Ignore comments from `NONE`/`CONTRIBUTOR` unless they are explicitly endorsed (👍 reaction or in-thread acknowledgement) by an OWNER/MEMBER/COLLABORATOR.
3. Existing labels on the issue (e.g. `bug`, `enhancement`, `area:*`) — hint at type and scope.

**Field derivation** (substitute into `templates/brief.md.tmpl`):

- `title` ← issue title
- `task_key`, `issue_number`, `github_url`, `spec_revision` — as before
- `parent_issue_url_or_none` ← `none`
- `type` ← infer: `bug` if labels match `/^(bug|defect|fix|regression)$/i` OR if body describes a defect (reproduction steps + expected vs actual); else `feature`
- `shape` ← `atomic` (composite was refused in Step 2; if you discover during extraction that the issue actually contains multiple unrelated work items, halt and instruct the user to either split the issue manually or invoke `/mino-task adopt issue-N --force-atomic` to merge them)
- `executability` ← `executable`
- `depends_on_task_keys_or_none` ← `none` (cross-issue dependency discovery is out of scope for adopt)
- `acceptance_criteria_checklist` ← **structured extraction**, not verbatim. Produce a markdown checklist (`- [ ] ...` lines) of testable outcomes derived from the issue body and qualifying comments. Each item MUST be a verifiable statement (e.g. `- [ ] Calling foo() with null returns NullPointerException with message "x"`), not a paraphrase of feelings (e.g. `- [ ] Fix the bug`). If the issue is too vague to yield ≥1 testable item, write a single line `- [ ] _(insufficient detail — see Open Questions)_` and populate `Open Questions / Warnings` with the specific gaps.
- `verification_steps` ← derived from the issue, not a placeholder. Examples by type:
  - **bug** → list the reproduction steps from the issue, then `- Expected: <expected>` `- Actual after fix: <expected>` lines
  - **feature** → list the user-visible behaviors that must hold after implementation
  - If the issue provides none and none can be inferred, write `_(verification will route to pending_acceptance — manual user sign-off required)_` AND set the brief's `Manual Acceptance` section header note to make this expectation explicit.
- `target_files_list` ← best-effort inference. Sources:
  - Filenames or paths mentioned in the issue body or comments (e.g. `src/foo.ts:123`)
  - Symbol/function names mentioned, resolved against the repo via grep
  - Stack traces in the issue (extract file paths)
  - If nothing can be inferred, write `_(unknown at adoption — run will populate from grep/codebase exploration)_` (this is the only field where a placeholder remains acceptable, because run can legitimately discover targets at execution time)
- `work_breakdown_or_not_applicable` ← `not_applicable`
- `next_stage` ← `run`
- `workflow_entry_state` ← `ready_to_start`
- `Open Questions / Warnings` section ← if extraction surfaced ambiguities, list them as `- Q: ...` lines so the user can edit the brief or add an issue comment before approving (Adopt-Step 5 already gated this; if questions exist, halt and re-prompt for approval citing the questions).

Write to `.mino/briefs/issue-{N}.md`.

**Quality bar**: a brief produced by this step must, when read in isolation, be sufficient for `run` to attempt implementation without re-reading the issue. If you cannot reach that bar from the available inputs, the brief is allowed to be sparse — but the sparsity MUST be explicit (`Open Questions / Warnings` populated), not hidden behind placeholder text.

### Adopt-Step 7: Render & record event (silent)

If `mode = adopted`, render `templates/event-task-adopted.yml.tmpl`. If `mode = re_adopted`, render `templates/event-task-re-adopted.yml.tmpl`. In either case substitute:

- `task_key`, `issue_number`, `spec_revision`, `issue_url`, `original_title`, `adopted_at_iso`
- (re_adopted only) `previous_revision`, `archive_path`

Write the rendered file to `.mino/events/issue-{N}/0001-task-{adopted|re-adopted}.yml`.

`task_adopted` and `task_re_adopted` are **silent** events: do NOT post a GitHub comment. If the local write fails, halt with the filesystem error; the adoption did not take effect.

### Adopt-Step 8: Apply labels

```
gh issue edit {N} --add-label "iron-tree:adopted" --add-label "stage:task"
# if mode = re_adopted, also remove any leftover stage labels:
gh issue edit {N} --remove-label "stage:run" --remove-label "stage:verify" --remove-label "stage:done"
```

Label-edit failures are warnings, not errors — log `stage_label_sync_failed: <reason>` in the report and continue. Local yml remains authoritative.

### Adopt-Step 9: Report & next-step hint

```
Adopted #{N} as {task_key} (revision {spec_revision}, mode {adopted|re_adopted}).
Run `/mino-run issue-{N}` to start.
```

After Adopt Mode finishes, control returns to the user. Adopt Mode does **not** fall through to native Step 1–6.

## Workflow

### Step 1: Intake

- Read the full Markdown file the user provides.
- Preserve the source path; it is required for `Spec Revision` and the `Source` section of every brief.

### Step 2: Decompose & Classify

For each work item discovered in the document, determine:

- `type`: `feature` or `bug`
- `shape`: `atomic` or `composite`
- `executability`: `executable` or `container`

If `composite` or `container`:

- Break into child tasks
- Assign explicit `depends_on` edges between children
- Ensure the graph is acyclic
- A `container` task MUST be decomposed; it cannot be executed directly

### Step 3: Compute Identity

For every node, compute two deterministic values:

**Task Key** — stable logical identity, never changes across reruns.

```
task_key = slugify( parent_task_key + "/" + title )
```

Where `slugify`:
- Lowercase ASCII
- Spaces and underscores → `-`
- Strip non-`[a-z0-9-]` characters
- Collapse repeated `-`
- Trim leading/trailing `-`
- Truncate to 64 characters

For top-level tasks, omit the `parent_task_key + "/"` prefix.

**Spec Revision** — fingerprint of the approved DAG snapshot.

```
spec_revision = sha256( normalize(spec_markdown) + "\n---\n" + canonical_json(dag_nodes) )[:8]
```

Where:
- `normalize(spec_markdown)`: strip trailing whitespace per line, convert CRLF to LF, collapse runs of blank lines to a single blank line
- `canonical_json(dag_nodes)`: sorted keys, no extra whitespace; each node serialized as `{"task_key":..., "title":..., "type":..., "shape":..., "executability":..., "depends_on":[sorted task keys]}`
- `[:8]`: first 8 hex characters of the digest

The same spec + same DAG MUST produce the same `Spec Revision` regardless of which agent computes it.

### Step 4: Approval Gate

Show a concise DAG preview to the user. One line per task:

```
{Task Key} [{type}/{shape}] {Title}{ → depends_on: a, b}
```

For each previously-published task whose `Task Key` exists with a different `Approved Revision`, mark the line as `(re-approval needed)`.

End with the prompt:

> "Approve this DAG revision `{spec_revision}`? (yes / edit / cancel)"

Do NOT proceed to Step 5 until the user explicitly approves the current `Spec Revision`. An unrelated previous approval does not carry forward.

### Step 5: Publish

After approval, for each task in dependency order:

1. **Resolve issue identity**
   - Look up an existing issue by `Task Key` (search the repo's issues for one whose body contains `Task Key: {task_key}`)
   - If found: this is a refresh
   - If not found: this is a create

2. **Render issue body** from `templates/issue-body.md.tmpl`
   - `next_stage`: `decompose` if shape is composite, else `run`
   - `workflow_entry_state`: `needs_breakdown` if container/composite, else `ready_to_start`
   - `close_on_done`: resolve in this priority order:
     1. `.mino/config.yml` → `issue.close_on_done` if present
     2. `manual` if `type: bug`
     3. `auto` otherwise
   - `depends_on_task_keys_or_none`: comma-separated list of dependency `Task Key`s, or `none`
   - `depends_on_github_links_block`: one `- Depends on #N` line per resolved dependency issue, or `_(no dependencies)_`. Resolve each dependency's issue number from the publish set so GitHub auto-links the issues; if a dependency is being created in the same publish pass, render its line **after** the dependency's own create completes.

3. **Create or update the GitHub issue**
   - Use `gh issue create` for new issues, `gh issue edit` for existing ones
   - Apply the label resolved from `type`: `feature` → `enhancement` (or `feature` if that label exists), `bug` → `bug`. Skip the label flag gracefully if `gh label list` does not contain the chosen label.

4. **Render local brief** from `templates/brief.md.tmpl`
   - Write to `.mino/briefs/issue-{N}.md` after the issue number is known
   - Brief state is local cache only — do NOT stage or commit `.mino/briefs/`

5. **Record the `task_published` event locally** — render `templates/event-task-published.yml.tmpl`, write to `.mino/events/issue-{N}/0001-task-published.yml`. This event is **silent**: do NOT post a GitHub comment.

6. **Handle local event write failure** — if step 5 fails (filesystem error, permission):
   - Do NOT roll back the created issue or brief.
   - Mark this task in the publish report as `local_event_write_failed` and print the exact filesystem error.
   - The user must resolve the write failure and manually re-invoke `/mino-task <spec>` (idempotent — task_key resolves to the existing issue, only the event file is re-created).

7. **Sync stage label** — `gh issue edit {N} --add-label "stage:task"` (idempotent).

   For native publishes the issue starts at `stage:task`; transition to `stage:run` happens when the user approves and `/mino-run` is invoked. (Approval-time label flip: when `task` records the user's `yes`, before exiting Step 5, run `gh issue edit {N} --remove-label "stage:task" --add-label "stage:run"` for each newly-approved task. Failures here are warnings, not errors. Label sync is not an event and does not write a local yml.)

### Step 6: Report

Summarize the publish results in a single table-shaped block:

```
| Task Key | Issue | Status   | Depends On |
| ...      | #N    | created  | ...        |
| ...      | #M    | refreshed| ...        |
| ...      | #K    | event_publish_failed | ... |
```

Conclude with the next-step hint:

> "Run `/mino-run issue-{N}` to start the first ready task: `{task-key}`."

Pick `{N}` as the lowest-numbered issue whose `Workflow Entry State: ready_to_start` and whose `Depends On` is `none`.

## Templates

All artifact shapes are defined by template files; this skill MUST NOT generate freehand variations.

- `templates/brief.md.tmpl` — local brief format (16 sections, matches `https://github.com/robinv8/mino-skills/blob/main/skills/references/brief-contract.md`)
- `templates/issue-body.md.tmpl` — GitHub issue body format
- `templates/event-task-published.yml.tmpl` — `task_published` event payload (matches the schema in `https://github.com/robinv8/mino-skills/blob/main/skills/references/workflow-state-contract.md`)

Variable syntax is `{{ variable_name }}`. Replace literally; do not introduce conditional logic in templates. Variants (composite vs atomic, has-deps vs no-deps) are handled by passing different placeholder values, not by branching the template.

## Constraints

- Do NOT create issues or briefs before approval.
- Do NOT let a `container` task enter execution directly — it MUST be decomposed.
- Do NOT omit `depends_on` when ordering matters.
- Do NOT commit `.mino/briefs/` files.
- Do NOT create duplicate issues when rerunning `task`; resolve by `Task Key`.
- Do NOT reuse a prior approval when `Spec Revision` changed.
- Do NOT invent fields in the YAML event; the schema is fixed by `workflow-state-contract.md`.
- Do NOT roll back on event-publish failure; surface it in the report instead.
- Keep the DAG preview concise — one line per task, show dependencies as `→`.
- Do NOT run Adopt Mode against a CLOSED issue or a composite issue (≥3 `- [ ]` checkboxes).
- Do NOT delete the archived directory created during re-adopt; it is the audit trail.
- Do NOT treat label sync failures as fatal; the local yml is authoritative.
- Do NOT skip the `gh auth status` precheck — the user's project may not have gh logged in.
- Do NOT post a GitHub comment for `task_published`, `task_adopted`, or `task_re_adopted` — all three are silent events in v1.10.
- Do treat local event file write failure as fatal for the task; the event did not happen unless the local file exists.

## References

- [references/workflow-state-contract.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/workflow-state-contract.md)
- [references/brief-contract.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/brief-contract.md)
- [references/iron-tree-protocol.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/iron-tree-protocol.md)
