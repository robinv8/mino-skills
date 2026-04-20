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

## Workflow

1. **Read the document** — ingest the full Markdown file the user provides.
2. **Classify** — determine for each node:
   - `type`: `feature` or `bug`
   - `shape`: `atomic` or `composite`
   - `executability`: `executable` or `container`
3. **Decompose (Fission)** — if `composite` / `container`:
   - Break into child tasks
   - Assign explicit `depends_on` edges between children
   - Ensure the graph is acyclic
4. **Derive stable identity**:
   - Derive a deterministic `Task Key` from the spec path, parent task key, and title
   - Compute a `Spec Revision` from the normalized spec and DAG shape
5. **Draft the DAG** — present a concise preview:
   - Parent task title + classification
   - Child tasks with their dependencies
   - Task keys
   - Execution order implied by `depends_on`
   - If an existing published task with the same `Task Key` has a different `Approved Revision`, show that this is a re-approval of a new revision
6. **Approval gate** — stop and ask the user:
   > "Approve this DAG revision? (yes / edit / cancel)"
   Do NOT proceed until the user explicitly approves the current `Spec Revision`.
7. **Publish** — after approval:
   - Create or refresh GitHub issues for each task
   - Reuse an existing issue when the same `Task Key` already exists
   - If an existing issue has a different `Approved Revision`, refresh it only because the current `Spec Revision` has just been explicitly approved
   - Write stable issue metadata:
     ```markdown
     ## Status
     - Task Key: {task-key}
     - Spec Revision: {spec-revision}
     - Approved Revision: {spec-revision}
     - Approval State: approved
     - Current Stage: definition
     - Next Stage: {run | decompose}
     - Workflow Entry State: {ready_to_start | needs_breakdown}
     - Attempt Count: 0
     - Max Retry Count: 3
     - Code Publication State: not_applicable
     - Close On Done: {auto | manual}
     - Depends On: {task-key-a, task-key-b | none}
     ```
   - Determine `Close On Done` in this priority order:
     1. If `.mino/config.yml` has `issue.close_on_done`, use that value
     2. If `type: bug`, use `manual`
     3. Otherwise, use `auto`
   - Apply labels based on classification:
     - `feature` → label `enhancement` (or `feature` if that label exists)
     - `bug` → label `bug`
     - Use `gh issue create --label` or `gh issue edit`; skip gracefully if the label does not exist
   - Link children to parent
   - Generate or refresh local briefs in `.mino/briefs/issue-{N}.md` after the issue number is known
   - Keep brief state local-only; do NOT commit `.mino/briefs/`
8. **Post publish event** — add a structured `task_published` event comment for each published task.
9. **Report publish results** — summarize created or refreshed issue numbers, task keys, and dependencies.

## Brief Format

Each brief contains task definition and workflow cache state. It must be sufficient for `run` to rebuild the DAG locally, while the linked GitHub issue body and structured workflow events remain the authoritative record.

```markdown
# {Title}

## Issue
- Task Key: {task-key}
- Issue Number: {issue-number}
- GitHub: {issue URL}
- Parent Issue: {issue URL | none}

## Classification
- Type: {feature|bug}
- Shape: {atomic|composite}
- Executability: {executable|container}
- Approval State: approved

## Dependencies
- Depends On: {task-key-a, task-key-b | none}

## Acceptance Criteria
- [ ] ...

## Verification
1. ...

## Target Files
- `path/to/file`

## Work Breakdown
{For composites: list child tasks and order}

## Workflow State
- Spec Revision: {spec-revision}
- Approved Revision: {spec-revision}
- Current Stage: definition
- Next Stage: {run | decompose}
- Workflow Entry State: {ready_to_start | needs_breakdown}
- Attempt Count: 0
- Max Retry Count: 3
- Code Publication State: not_applicable

## Source
- Spec: `path/to/spec.md`
```

## Constraints

- Do NOT create issues or briefs before approval.
- Do NOT let a `container` task enter execution directly — it MUST be decomposed.
- Do NOT omit `depends_on` when ordering matters.
- Do NOT commit `.mino/briefs/` files.
- Do NOT create duplicate issues when rerunning `task`.
- Do NOT reuse a prior approval when `Spec Revision` changed.
- Keep the DAG preview concise — one line per task, show dependencies as `→`.

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/brief-contract.md](../references/brief-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
