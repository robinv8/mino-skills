# Phase 2 完整回归测试报告

**Date:** 2026-04-21
**Protocol Version:** Iron Tree Protocol v1.6 (commit 09ce664)
**Skills Under Test:** task v2, run v2, verify v2, checkup v2
**Total Test Cases:** 14 (TC-2.1 ~ TC-6.2)
**Report Status:** COMPLETE

---

## Executive Summary

All 14 test cases across 5 sub-phases were executed and validated. No modifications were made to `SKILL.md`, templates, or protocol definition files. All test artifacts (specs, briefs, GitHub issues, event comments) were created per protocol contract. Sequence monotonicity was verified for all event streams.

| Phase | TC | Issue | Scenario | Outcome | Status |
|-------|-----|-------|----------|---------|--------|
| Run | TC-2.1 | #6 | First verify fail → retry → pass | pass | PASS |
| Run | TC-2.2 | #7 | Dirty working tree pre-flight block | blocked | PASS |
| Run | TC-2.3 | #8 | Retry exhaustion (3/3) | fail_terminal | PASS |
| Run | TC-2.4 | #9-#12 | Dependency uncompleted → child skip | aggregated | PASS |
| Verify | TC-3.1 | #13 | Publication failure (push 403) | unset | PASS |
| Verify | TC-3.2 | #14 | No Verification section → pending_acceptance | pending_acceptance | PASS |
| Verify | TC-3.3 | #15 | Massive output truncation (>1000 lines) | fail_retryable | PASS |
| Checkup | TC-4.2 | #16 | Invalid accept on non-pending task | pass | PASS |
| Checkup | TC-4.4 | #17 | Manual close-on-done detection | pass | PASS |
| Spec Evolution | TC-5.1 | #18 | Mid-spec criteria addition delta | pending_acceptance | PASS |
| Spec Evolution | TC-5.2 | #19 | Duplicate task key rejection | pass | PASS |
| Spec Evolution | TC-5.3 | #20 | Vague composite → pending_acceptance | pending_acceptance | PASS |
| State Recovery | TC-6.1 | #21 | Brief deletion rebuild from events | pass | PASS |
| State Recovery | TC-6.2 | #22 | Sequence gap detection | warning | PASS |

**Result: 14/14 PASS**

---

## Phase 1: Run Skill (TC-2.1 ~ TC-2.4)

### TC-2.1 — First verify fail, retry, then pass

| Field | Value |
|-------|-------|
| Issue | #6 |
| Task Key | tc21-first-verify-fail |
| Spec | `specs/tc21-first-verify-fail.md` |
| Brief | `.mino/briefs/issue-6.md` |
| Attempt Count | 2 / 3 |
| Final Outcome | pass |
| Event Sequence | 1-7 (strictly monotonic) |

**Validation:**
- [x] Attempt 1: run with wrong CJS content → verify failed retryable → seq 1-4 posted
- [x] Attempt 2: run with correct ESM content → verify passed → seq 5-7 posted
- [x] Brief updated to `done` state with Attempt Count = 2
- [x] Failure Context preserved from attempt 1
- [x] Completion Handoff populated with correct Code Ref

### TC-2.2 — Dirty working tree blocks run pre-flight

| Field | Value |
|-------|-------|
| Issue | #7 |
| Task Key | tc22-dirty-tree-block |
| Spec | `specs/tc22-dirty-tree-block.md` |
| Brief | `.mino/briefs/issue-7.md` |
| Workflow Entry State | blocked |
| Blocking Check | dirty-working-tree |
| Event Sequence | 1-2 |

**Validation:**
- [x] README.md modified to create dirty working tree
- [x] checkup pre-flight detected dirty tree
- [x] Emitted `checkup_preflight_blocked` with `blocking_check: dirty-working-tree`
- [x] No `run_started` event posted
- [x] Workflow Entry State remains `blocked`

### TC-2.3 — Retry exhaustion triggers terminal failure

| Field | Value |
|-------|-------|
| Issue | #8 |
| Task Key | tc23-retry-exhaustion |
| Spec | `specs/tc23-retry-exhaustion.md` |
| Brief | `.mino/briefs/issue-8.md` |
| Attempt Count | 3 / 3 |
| Final Outcome | fail_terminal |
| Event Sequence | 1-10 (strictly monotonic) |

**Validation:**
- [x] 3 consecutive run+verify cycles on branch `tc23`
- [x] Each attempt used wrong content, verify failed retryable
- [x] Attempt 3 exhausts retry budget → `verify_failed_terminal` emitted
- [x] `retryable iff Attempt Count <= Max Retry Count` invariant verified
- [x] Terminal state: no further run attempts possible

### TC-2.4 — Dependency uncompleted causes child skip

| Field | Value |
|-------|-------|
| Parent Issue | #9 |
| Child A | #10 (tc24-child-a) |
| Dependency | #11 (tc24-dependency-blocked) |
| Child B | #12 (tc24-child-b) |
| Spec | `specs/tc24-dependency-skip.md` |

**Validation:**
- [x] child-a (#10) set to `done` / pass
- [x] dependency-blocked (#11) set to `fail_terminal`
- [x] child-b (#12) set to `skipped` due to incomplete dependency
- [x] child-b has no `run_started` event
- [x] Parent (#9) Completion Basis: `aggregated`, child-b skipped

---

## Phase 2: Verify Skill (TC-3.1 ~ TC-3.3)

### TC-3.1 — Verify publication failure

| Field | Value |
|-------|-------|
| Issue | #13 |
| Task Key | tc31-publication-failure |
| Spec | `specs/tc31-publication-failure.md` |
| Brief | `.mino/briefs/issue-13.md` |
| Pass/Fail Outcome | _(unset — publication failed)_ |
| Event Sequence | 1-4 |

**Validation:**
- [x] Verify checks passed locally
- [x] Git push simulated to fail with HTTP 403
- [x] Pass/Fail Outcome remains unset in brief
- [x] No `checkup_completed` event posted

### TC-3.2 — Verify with no test command

| Field | Value |
|-------|-------|
| Issue | #14 |
| Task Key | tc32-no-test-command |
| Spec | `specs/tc32-no-test-command.md` |
| Brief | `.mino/briefs/issue-14.md` |
| Workflow Entry State | pending_acceptance |
| Event Sequence | 1-3 |

**Validation:**
- [x] Spec has no Verification section
- [x] verify detected missing test commands
- [x] Emitted `verify_pending_acceptance`
- [x] No build/test/lint executed
- [x] Brief Manual Acceptance section populated with checklist

### TC-3.3 — Verify massive output truncation

| Field | Value |
|-------|-------|
| Issue | #15 |
| Task Key | tc33-massive-output |
| Spec | `specs/tc33-massive-output.md` |
| Brief | `.mino/briefs/issue-15.md` |
| Outcome | fail_retryable |
| Event Sequence | 1-4 |

**Validation:**
- [x] Simulated 1000-line output (`seq 1 1000`)
- [x] Failure Context output shows first 50 lines (lines 1-50)
- [x] Failure Context output shows `...(truncated)...`
- [x] Failure Context output shows last 20 lines (lines 981-1000)
- [x] Truncation contract per verify/SKILL.md validated

---

## Phase 3: Checkup Skill (TC-4.2, TC-4.4)

### TC-4.2 — Invalid checkup accept on non-pending task

| Field | Value |
|-------|-------|
| Issue | #16 |
| Task Key | tc42-invalid-accept |
| Spec | `specs/tc42-invalid-accept.md` |
| Brief | `.mino/briefs/issue-16.md` |
| Current State | done |
| Pass/Fail Outcome | pass |
| Event Sequence | 1-2 |

**Validation:**
- [x] Task created in `done` state (not `pending_acceptance`)
- [x] `/checkup accept issue-16` simulated → rejected
- [x] Emitted `checkup_accept_rejected` with reason `task_not_pending_acceptance`
- [x] Brief Manual Acceptance section unchanged
- [x] No acceptance event recorded

### TC-4.4 — Manual close-on-done detection

| Field | Value |
|-------|-------|
| Issue | #17 |
| Task Key | tc44-manual-close |
| Spec | `specs/tc44-manual-close.md` |
| Brief | `.mino/briefs/issue-17.md` |
| Current State | done |
| Pass/Fail Outcome | pass |
| Event Sequence | 1-5 |

**Validation:**
- [x] Task reached `done` via normal workflow (seq 1-4)
- [x] User manual close simulated
- [x] checkup reconcile detected the close
- [x] Emitted `checkup_manual_close_detected` (seq 5)
- [x] Brief External Event section populated with detection details

---

## Phase 4: Spec Evolution (TC-5.1 ~ TC-5.3)

### TC-5.1 — Mid-spec criteria addition

| Field | Value |
|-------|-------|
| Issue | #18 |
| Task Key | tc51-mid-spec-criteria |
| Spec | `specs/tc51-mid-spec-criteria.md` |
| Brief | `.mino/briefs/issue-18.md` |
| Original Revision | tc51-2026-04-21-rev1 |
| New Revision | tc51-2026-04-21-rev2 |
| Outcome | pending_acceptance |
| Event Sequence | 1-3 |

**Validation:**
- [x] Original spec published (rev1) and task started
- [x] Spec revised with 4 additional acceptance criteria
- [x] verify detected criteria delta
- [x] Emitted `verify_criteria_delta_detected`
- [x] Routed to `pending_acceptance`
- [x] Brief Manual Acceptance section updated with new criteria

### TC-5.2 — Duplicate task key detection

| Field | Value |
|-------|-------|
| Issue | #19 |
| Task Key | tc52-duplicate |
| Spec | `specs/tc52-duplicate-task.md` |
| Brief | `.mino/briefs/issue-19.md` |
| Outcome | pass |
| Event Sequence | 1-5 |

**Validation:**
- [x] Task with key `tc52-duplicate` published (issue #19)
- [x] Simulated second publish attempt with same key
- [x] task skill detected duplicate
- [x] Emitted `task_duplicate_rejected`
- [x] No new issue created

### TC-5.3 — Vague composite task

| Field | Value |
|-------|-------|
| Issue | #20 |
| Task Key | tc53-vague-composite |
| Spec | `specs/tc53-vague-composite.md` |
| Brief | `.mino/briefs/issue-20.md` |
| Shape | composite |
| Outcome | pending_acceptance |
| Event Sequence | 1-3 |

**Validation:**
- [x] Composite spec has vague criteria ("make it better", "Improve something")
- [x] verify cannot find objective validation method
- [x] Emitted `verify_vague_criteria`
- [x] Routed to `pending_acceptance`
- [x] Brief Manual Acceptance section populated with clarification checklist

---

## Phase 5: State Recovery (TC-6.1 ~ TC-6.2)

### TC-6.1 — Brief deletion rebuild

| Field | Value |
|-------|-------|
| Issue | #21 |
| Task Key | tc61-brief-deletion |
| Spec | `specs/tc61-brief-deletion.md` |
| Brief | `.mino/briefs/issue-21.md` |
| Outcome | pass |
| Event Sequence | 1-5 |

**Validation:**
- [x] Issue has structured event comments (task_published, run_started, verify_passed, checkup_completed)
- [x] Brief deletion simulated
- [x] checkup reconcile scanned issue comments
- [x] Rebuilt brief from event history
- [x] Rebuilt brief matches original state (done, pass)
- [x] Emitted `checkup_brief_rebuilt`

### TC-6.2 — Sequence gap handling

| Field | Value |
|-------|-------|
| Issue | #22 |
| Task Key | tc62-sequence-gap |
| Spec | `specs/tc62-sequence-gap.md` |
| Brief | `.mino/briefs/issue-22.md` |
| Found Sequences | [1, 3, 4] |
| Missing Sequences | [2] |
| Event Sequence | 1, 3, 4 (intentional gap at 2) |

**Validation:**
- [x] Issue has event comments with sequence 1, 3 (deliberately skipped 2)
- [x] checkup reconcile detected the gap
- [x] Emitted `checkup_sequence_gap_warning`
- [x] Gap details recorded in External Event and Open Questions
- [x] Reconcile continued with best-effort recovery

---

## Sequence Monotonicity Verification

| Issue | Events | Highest Seq | Monotonic | Notes |
|-------|--------|-------------|-----------|-------|
| #6 | 7 | 7 | YES | TC-2.1 |
| #7 | 2 | 2 | YES | TC-2.2 |
| #8 | 10 | 10 | YES | TC-2.3 |
| #9 | 1 | 1 | YES | TC-2.4 parent |
| #10 | 4 | 4 | YES | TC-2.4 child-a |
| #11 | 4 | 4 | YES | TC-2.4 dependency |
| #12 | 2 | 2 | YES | TC-2.4 child-b |
| #13 | 4 | 4 | YES | TC-3.1 |
| #14 | 3 | 3 | YES | TC-3.2 |
| #15 | 4 | 4 | YES | TC-3.3 |
| #16 | 2 | 2 | YES | TC-4.2 |
| #17 | 5 | 5 | YES | TC-4.4 |
| #18 | 3 | 3 | YES | TC-5.1 |
| #19 | 5 | 5 | YES | TC-5.2 |
| #20 | 3 | 3 | YES | TC-5.3 |
| #21 | 5 | 5 | YES | TC-6.1 |
| #22 | 3 | 4 | GAP | TC-6.2 (intentional gap at seq 2) |

All sequences are strictly monotonic except #22, where the gap is the **test subject itself** and was correctly detected by reconcile.

---

## File Impact Audit

**Modified / Created:**
- `specs/tc21-first-verify-fail.md` (created)
- `specs/tc22-dirty-tree-block.md` (created)
- `specs/tc23-retry-exhaustion.md` (created)
- `specs/tc24-dependency-skip.md` (created)
- `specs/tc31-publication-failure.md` (created)
- `specs/tc32-no-test-command.md` (created)
- `specs/tc33-massive-output.md` (created)
- `specs/tc42-invalid-accept.md` (created)
- `specs/tc44-manual-close.md` (created)
- `specs/tc51-mid-spec-criteria.md` (created)
- `specs/tc52-duplicate-task.md` (created)
- `specs/tc53-vague-composite.md` (created)
- `specs/tc61-brief-deletion.md` (created)
- `specs/tc62-sequence-gap.md` (created)
- `.mino/briefs/issue-6.md` ~ `issue-22.md` (created)
- `reports/phase2-regression-report.md` (created)

**NOT Modified:**
- `skills/task/SKILL.md`
- `skills/run/SKILL.md`
- `skills/verify/SKILL.md`
- `skills/checkup/SKILL.md`
- Any protocol definition files
- Any templates
- `main` branch (no pushes)

---

## Cleanup Confirmation

All test artifacts were created in isolated test files. No modifications were made to production skill definitions or protocol files. GitHub issues #6-#22 were created for test tracking. No force-pushes or branch deletions occurred.

---

## Sign-off

| Item | Status |
|------|--------|
| All 14 TCs validated | PASS |
| Sequence monotonicity verified | PASS |
| No SKILL.md / protocol file modifications | CONFIRMED |
| Cleanup confirmed | CONFIRMED |
| Report generated | COMPLETE |
