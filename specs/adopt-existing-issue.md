# Spec: Adopt Existing GitHub Issues — Regression

Validates `/task adopt issue-N` against Iron Tree Protocol v1.9 § Adopting Existing Issues.

## TC-9.1: Adopt clean atomic OPEN issue

Given:
- A repo with `iron-tree:adopted` label NOT present on issue #100
- Issue #100 is OPEN, title "Add login button", body is plain text (no checkboxes)

When user runs `/task adopt issue-100` and approves.

Then:
- `.mino/briefs/issue-100.md` exists, contains `Spec Revision: <8 hex>`
- `.mino/events/issue-100/0001-task-adopted.yml` exists with `event: task_adopted, sequence: 1`
- A comment was posted on issue #100 containing the same yaml block
- Issue #100 has labels `iron-tree:adopted` AND `stage:task`
- Report ends with `Run \`/run issue-100\` to start.`

## TC-9.2: Refuse CLOSED issue

Given:
- Issue #101 is CLOSED.

When user runs `/task adopt issue-101`.

Then:
- No brief, no event, no comment, no label change
- Skill exits with error message `Issue #101 is CLOSED; only OPEN issues can be adopted.`

## TC-9.3: Refuse composite issue, mark needs-breakdown

Given:
- Issue #102 is OPEN, body contains 3+ lines starting with `- [ ]`.

When user runs `/task adopt issue-102`.

Then:
- No brief, no event, no comment
- Issue #102 has label `iron-tree:needs-breakdown` added
- Skill exits with error message starting `Issue #102 appears composite`

## TC-9.4: Re-adopt archives previous chain

Given:
- Issue #100 was adopted (TC-9.1 ran), brief exists at `.mino/briefs/issue-100.md` with `Approved Revision: AAAA1111`
- User edits issue #100 title or body on GitHub
- Issue #100 still has `iron-tree:adopted` label

When user runs `/task adopt issue-100` and approves.

Then:
- `.mino/archive/issue-100-rev-AAAA1111/brief.md` exists (the original brief)
- `.mino/archive/issue-100-rev-AAAA1111/events/0001-task-adopted.yml` exists
- New `.mino/briefs/issue-100.md` exists with a different `Spec Revision`
- New `.mino/events/issue-100/0001-task-re-adopted.yml` exists with `event: task_re_adopted, previous_revision: AAAA1111, archive_path: .mino/archive/issue-100-rev-AAAA1111/`
- Issue #100 still has `iron-tree:adopted` label
- Issue #100 has `stage:task` (any leftover `stage:run|verify|done` removed)

## TC-9.5: Full chain after adoption

Given:
- TC-9.1 has completed; issue #100 is adopted with `stage:task`.

When the operator runs:
1. Approve (covered by TC-9.1)
2. `/run issue-100` (run completes, commit lands)
3. `/verify issue-100` (verify_passed)

Then:
- After step 2: issue #100 has `stage:verify`, no `stage:task` or `stage:run`
- After step 3: issue #100 has `stage:done`, no `stage:verify`
- All standard events present in `.mino/events/issue-100/`: `task_adopted`, `run_started`, `run_completed`, `verify_passed`
- The commit message matches `[run] issue-100: <summary>` (existing run convention, no change)
