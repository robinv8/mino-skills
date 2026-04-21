---
name: task
description: |
  Convert a local Markdown requirement document into an execution DAG.
  Read a spec or bug note, extract tasks with dependencies, classify them,
  and prepare GitHub issues + local briefs. Always asks for approval
  before creating anything. Use when starting new work from a Markdown doc.
---

# Task Intake Engine

Read a user-specified local Markdown requirement document, extract a Directed Acyclic Graph (DAG) of work items, classify each, and prepare them for execution — but only after explicit user approval.

This skill is the protocol's entry point. Its outputs (issue body, local brief, structured event) are consumed by `run`, `verify`, and `checkup`. Field names, section names, and event schemas are **fixed by template files** in `templates/` so downstream skills can rely on them deterministically.

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

5. **Post the `task_published` event** by adding an issue comment rendered from `templates/event-task-published.yml.tmpl`. The event YAML schema is fixed by `../references/workflow-state-contract.md`; do not invent fields.

6. **Handle event publish failure** — if step 5 fails (network error, gh CLI failure, permissions):
   - Do NOT roll back the created issue or brief; both are already valid artifacts.
   - Mark this task in the publish report as `event_publish_failed`.
   - Continue publishing remaining tasks; one failure does not abort the batch.
   - In the final report, instruct the user to manually re-post via `gh issue comment {N} --body-file <rendered-event-file>`, or to run `/checkup reconcile` once available.

### Step 6: Report

Summarize the publish results in a single table-shaped block:

```
| Task Key | Issue | Status   | Depends On |
| ...      | #N    | created  | ...        |
| ...      | #M    | refreshed| ...        |
| ...      | #K    | event_publish_failed | ... |
```

Conclude with the next-step hint:

> "Run `/run issue-{N}` to start the first ready task: `{task-key}`."

Pick `{N}` as the lowest-numbered issue whose `Workflow Entry State: ready_to_start` and whose `Depends On` is `none`.

## Templates

All artifact shapes are defined by template files; this skill MUST NOT generate freehand variations.

- `templates/brief.md.tmpl` — local brief format (16 sections, matches `../references/brief-contract.md`)
- `templates/issue-body.md.tmpl` — GitHub issue body format
- `templates/event-task-published.yml.tmpl` — `task_published` event payload (matches the schema in `../references/workflow-state-contract.md`)

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

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/brief-contract.md](../references/brief-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
