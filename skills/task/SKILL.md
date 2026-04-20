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
2. **Classify** — determine for the top-level item:
   - `type`: `feature` or `bug`
   - `shape`: `atomic` (single deliverable) or `composite` (needs decomposition)
   - `executability`: `executable` (can be coded directly) or `container` (needs child tasks)
3. **Decompose (Fission)** — if `composite` / `container`:
   - Break into child tasks
   - Assign explicit `depends_on` edges between children
   - Ensure the graph is acyclic
4. **Draft the DAG** — present a clear preview:
   - Parent task title + classification
   - Child tasks with their dependencies
   - Execution order implied by `depends_on`
5. **Approval gate** — stop and ask the user:
   > "Approve this DAG? (yes / edit / cancel)"
   Do NOT proceed until the user explicitly approves.
6. **Publish** — after approval:
   - Create GitHub issues for each task (parent + children)
   - Link children to parent
   - Generate local briefs in `.mino/briefs/issue-{N}.md`

## Brief Format

Each brief must contain:

```markdown
# {Title}

## Classification
- Type: {feature|bug}
- Shape: {atomic|composite}
- Executability: {executable|container}

## Acceptance Criteria
- [ ] ...

## Verification Steps
1. ...

## Target Files
- `path/to/file`

## Work Breakdown
{For composites: list child tasks and order}

## Workflow State
- Current Stage: intake
- Next Stage: {run | decompose}
```

## Constraints

- Do NOT create issues or briefs before approval.
- Do NOT let a `container` task enter execution directly — it MUST be decomposed.
- Do NOT omit `depends_on` when ordering matters.
- Keep the DAG preview concise — one line per task, show dependencies as `→`.

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/brief-contract.md](../references/brief-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
