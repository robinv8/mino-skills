# Regression Report: Adopt Existing GitHub Issues

**Protocol Version:** 1.9
**Fixture Repo:** https://github.com/robinv8/mino-skills-fixture
**Date:** 2026-04-21
**Result:** 5 / 5 PASS

---

## TC-9.1: Adopt clean atomic OPEN issue

**Target:** Issue #1 — "Add login button", OPEN, plain text (no checkboxes)

### Commands
```bash
gh issue view 1 --json number,state,title,body,labels,url
# Computed spec_revision = a88936a4
# Rendered brief → .mino/briefs/issue-1.md
# Rendered event → .mino/events/issue-1/0001-task-adopted.yml
gh issue comment 1 --body-file .mino/events/issue-1/0001-task-adopted.yml
gh issue edit 1 --add-label "iron-tree:adopted" --add-label "stage:task"
```

### Final State
- `.mino/briefs/issue-1.md` exists, contains `Spec Revision: a88936a4`
- `.mino/events/issue-1/0001-task-adopted.yml` exists with `event: task_adopted, sequence: 1`
- Comment posted: https://github.com/robinv8/mino-skills-fixture/issues/1#issuecomment-4288746177
- Labels: `iron-tree:adopted`, `stage:task`

**PASS**

---

## TC-9.2: Refuse CLOSED issue

**Target:** Issue #2 — "Refactor database layer", CLOSED

### Commands
```bash
gh issue view 2 --json number,state
# state = "CLOSED"
```

### Result
- No brief created
- No event created
- No comment posted
- No label change
- Error message: `Issue #2 is CLOSED; only OPEN issues can be adopted.`

**PASS**

---

## TC-9.3: Refuse composite issue, mark needs-breakdown

**Target:** Issue #3 — "Build new dashboard", OPEN, body contains 4 lines starting with `- [ ]`

### Commands
```bash
gh issue view 3 --json number,state,title,body
# Body contains 4 open checkboxes
gh issue edit 3 --add-label "iron-tree:needs-breakdown"
```

### Result
- No brief created
- No event created
- No comment posted
- Label `iron-tree:needs-breakdown` added
- Error message: `Issue #3 appears composite (4 open checkboxes). Split it into child issues, then run /task adopt on each child.`

**PASS**

---

## TC-9.4: Re-adopt archives previous chain

**Target:** Issue #1 (previously adopted in TC-9.1, then body edited by user)

### Commands
```bash
gh issue edit 1 --body "We need a login button ... Also needs a logout button."
# Computed new spec_revision = abe1db69
mkdir -p .mino/archive/issue-1-rev-a88936a4
mv .mino/briefs/issue-1.md .mino/archive/issue-1-rev-a88936a4/brief.md
mv .mino/events/issue-1 .mino/archive/issue-1-rev-a88936a4/events
# Rendered new brief → .mino/briefs/issue-1.md
# Rendered new event → .mino/events/issue-1/0001-task-re-adopted.yml
gh issue comment 1 --body-file .mino/events/issue-1/0001-task-re-adopted.yml
gh issue edit 1 --add-label "stage:task" --remove-label "stage:run" --remove-label "stage:verify" --remove-label "stage:done"
```

### Final State
- `.mino/archive/issue-1-rev-a88936a4/brief.md` exists (original brief)
- `.mino/archive/issue-1-rev-a88936a4/events/0001-task-adopted.yml` exists
- New `.mino/briefs/issue-1.md` exists with `Spec Revision: abe1db69`
- New `.mino/events/issue-1/0001-task-re-adopted.yml` exists with `event: task_re_adopted, previous_revision: a88936a4, archive_path: .mino/archive/issue-1-rev-a88936a4/`
- Labels: `iron-tree:adopted`, `stage:task`

**PASS**

---

## TC-9.5: Full chain after adoption

**Target:** Issue #1 (re-adopted in TC-9.4)

### Step 2 — Run
```bash
# Simulated code change in fixture repo
echo "# Mino Skills Fixture" > README.md
git add README.md
git commit -m "[run] issue-1: update readme with fixture note"
# Created .mino/events/issue-1/0002-run-started.yml
# Created .mino/events/issue-1/0003-run-completed.yml
gh issue edit 1 --remove-label "stage:task" --add-label "stage:verify"
```

### Step 3 — Verify
```bash
# Created .mino/events/issue-1/0004-verify-passed.yml
gh issue edit 1 --remove-label "stage:verify" --add-label "stage:done"
```

### Final State
- After run: labels `iron-tree:adopted`, `stage:verify` (no `stage:task`, no `stage:run`)
- After verify: labels `iron-tree:adopted`, `stage:done` (no `stage:verify`)
- All standard events present in `.mino/events/issue-1/`:
  - `0001-task-re-adopted.yml` (sequence 1)
  - `0002-run-started.yml` (sequence 2)
  - `0003-run-completed.yml` (sequence 3)
  - `0004-verify-passed.yml` (sequence 4)
- Commit message: `[run] issue-1: update readme with fixture note`

**PASS**

---

## Fixture .mino File Tree

```
.mino/
├── archive/
│   └── issue-1-rev-a88936a4/
│       ├── brief.md
│       └── events/
│           └── 0001-task-adopted.yml
├── briefs/
│   └── issue-1.md
└── events/
    └── issue-1/
        ├── 0001-task-re-adopted.yml
        ├── 0002-run-started.yml
        ├── 0003-run-completed.yml
        └── 0004-verify-passed.yml
```
