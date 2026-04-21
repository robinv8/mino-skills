# Spec: Silent Events & Consolidated Summary — Regression

Validates protocol v1.10 event category policy across task / run / verify / checkup.

## TC-10.1: Happy path produces exactly one GitHub comment

Given:
- Clean atomic OPEN issue #200 on a test repo
- `iron-tree:adopted` label absent

When operator runs:
1. `/task adopt issue-200` → approve
2. `/run issue-200` (code change succeeds, commit lands)
3. `/verify issue-200` (all checks pass)
4. `/checkup accept issue-200` and `/checkup done issue-200` (or unified `/checkup issue-200` if supported)

Then:
- Issue #200 has exactly **1** comment whose body starts with `🏁 Issue 200 done`
- That comment contains 4 inlined yml blocks in order: `task_adopted`, `run_started`+`run_completed` (or just `run_completed` if run_started merged), `verify_passed`, `checkup_done` (and any accept/aggregate if emitted)
- `.mino/events/issue-200/` contains the same number of files as the inlined blocks
- Issue labels: `iron-tree:adopted`, `stage:done`

## TC-10.2: Verify failure produces immediate audible comment

Given:
- Issue #201 adopted, approved, run committed
- Verify check fails (retryable)

When operator runs `/verify issue-201`.

Then:
- Issue #201 has exactly **1** comment so far, body starts with `❌` or similar failure marker, contains yml block `event: verify_failed_retryable`, and contains the literal line `Local events: \`.mino/events/issue-201/\``
- `.mino/events/issue-201/*.yml` contains: `task_adopted`, `run_started`, `run_completed`, `verify_failed_retryable`
- Issue label `stage:verify` (not moved to `stage:done`)

## TC-10.3: Reconcile from local events (primary source)

Given:
- Issue #200 completed (TC-10.1 state).
- User manually runs `/checkup reconcile issue-200`.

Then:
- Reconcile reads `.mino/events/issue-200/` as primary source
- No new comment posted on issue #200 (reconcile only emits on drift)
- Report states "local events: 4 file(s), primary source used"

## TC-10.4: Reconcile fallback from terminal summary

Given:
- Issue #200 completed.
- User deletes `.mino/events/issue-200/` and `.mino/briefs/issue-200.md` (simulating local loss).

When operator runs `/checkup reconcile issue-200`.

Then:
- Reconcile finds no primary events
- Falls back to terminal summary comment, parses 4 inlined yml blocks
- Rebuilds `.mino/events/issue-200/*.yml` (4 files) and `.mino/briefs/issue-200.md`
- Report states "local events: 0 file(s), terminal summary used, rebuilt 4 event(s)"

## TC-10.5: Reconcile fallback for legacy v1.9 issue

Given:
- Issue #300 completed under v1.9 with 4 separate per-event comments (not a terminal summary).
- Local events absent.

When operator runs `/checkup reconcile issue-300` under v1.10 skills.

Then:
- Primary source empty
- Terminal summary signature not matched
- Falls back to legacy per-event parsing; rebuilds local events from comments
- Report states "local events: 0 file(s), legacy per-event comments used, rebuilt 4 event(s)"

## TC-10.6: Halt comment carries local-events pointer

Given:
- Issue #202 in run stage, pre-commit hook configured to reject.

When operator runs `/run issue-202`.

Then:
- Issue #202 gets 1 new comment
- Comment body contains the literal line `Local events: \`.mino/events/issue-202/\`` directly above the yml block
- Comment yml block has `event: run_commit_failed`
