**English** | [简体中文](TEST_PLAN.zh.md)

# Iron Tree Protocol — End-to-End Validation Plan

> Using a **business-mature and technically typical** real-world scenario — "login/register + user management" — to end-to-end validate whether the Mino skill pack can take a Markdown spec all the way to working, log-in-able, user-manageable product code.

---

## Validation Goals

| | |
|---|---|
| **Primary** | Prove the Mino Protocol can deliver a **truly usable complete feature** from a spec |
| **Secondary** | Validate that the protocol holds up under the unavoidable imperfections of real-world development |
| **Side output** | Capture currently undefined protocol boundaries as v3 design input |

---

## Validation Philosophy

- **Happy-path success is the baseline** — if it doesn't work, the protocol design has failed and no boundary capability can be discussed.
- **Boundary scenarios are "real-world imperfections", not deliberate sabotage** — each boundary TC maps to a situation that naturally occurs in real development, not a trap set for the protocol.
- **Failed boundary scenarios do not block primary validation** — they are inputs for v3 protocol patches, not release blockers.

---

## Test Vehicle

| Item | Description |
|---|---|
| **Host repo** | `~/Projects/todo-list` (standalone project, not Mino itself, to avoid meta-tasks) |
| **Baseline tag** | `v0-mino-baseline` (reset to this each round for reproducibility) |
| **Loaded spec** | `specs/auth-system.md` (register, login, password recovery, user list, edit profile, delete user) |
| **Tech stack** | TypeScript / Node / pnpm (framework determined by todo-list state) |
| **Expected artifacts** | Parent issue × 1, child issues × 5, briefs × 6, actual runnable PRs/commits |

---

## Test Structure & Priority

| Phase | Role | Must pass? | Failure meaning |
|---|---|---|---|
| **Phase 1 — End-to-end mainline** | Full happy path | ✅ Must all pass | Protocol design failure |
| **Phase 2 — Natural imperfection** | Common real-world situations | ⚠️ Most should pass | Skill implementation needs patching |
| **Phase 3 — Protocol TBD** | Boundaries not yet defined by current protocol | ❌ Failure = v3 input | Protocol spec needs new clauses |

Each test record format:
```
**Scenario** — when does this happen in real development?
**Operation** — how to trigger it
**Expected** — how the protocol says it should respond
**Actual** — fill after running
**Root cause** — fill on failure
**Fix** — change skill / change protocol / change spec / won't fix
```

---

## Phase 1 — End-to-End Mainline (must pass)

### TC-1.1 From spec to the first log-in-able user

**Scenario**: A user has a complete auth-system spec and wants to go from opening the terminal to manually registering, logging in, and seeing themselves in the user list — **without skipping any protocol stage**.

**Operation**:
1. `cd ~/Projects/todo-list && git reset --hard v0-mino-baseline`
2. Write `specs/auth-system.md`
3. `/task specs/auth-system.md`
4. Approve the generated DAG
5. Run `/run issue-N` in sequence; each auto → `/verify` → `/checkup`
6. After all sub-tasks done, `/checkup aggregate issue-1`

**Expected artifacts (protocol level)**:
- [ ] Create 6 issues (1 parent + 5 children)
- [ ] Each child brief goes through `run(attempt=1) → verify → checkup → done`
- [ ] Each child `Completion Basis: verified`
- [ ] Parent `Completion Basis: aggregated`
- [ ] All issues closed (`Close On Done: auto`)
- [ ] Commit history traceable to each Task Key

**Expected artifacts (product level — this is real acceptance)**:
- [x] `pnpm install && pnpm build` succeeds
- [x] `pnpm test` all pass
- [x] `pnpm dev` starts
- [x] **Can register a real new user**
- [x] **Can log in / log out with that user**
- [x] **Can see that user on the user list page (task list)**
- [ ] **Can edit that user's profile and persist changes** (not in spec)
- [ ] **Can delete that user as admin** (not in spec)

**Actual**:
- Created 6 issues (#20 parent + #21~25 children) ✅
- All 5 child tasks completed, briefs went run → verify → checkup → done ✅
- Each child Completion Basis: verified ✅
- Parent Completion Basis: aggregated ✅
- Commit history traceable to each Task Key (3cae470 auth-core, b75a9cf auth-pages, b97aef9 auth-routing, f81b908 auth-isolation, 1102c89 auth-tests) ✅
- `pnpm build` succeeded ✅
- `pnpm test` 14 tests all passed ✅
- Can register, log in, log out, add/switch/delete tasks ✅

**Root cause**: None
**Fix**: None

---

### TC-1.2 Protocol artifact structural correctness

**Scenario**: After TC-1.1, check whether all protocol artifact fields are present and contracts honored — this lays the foundation for downstream capabilities like `reconcile` / `aggregate`.

**Expected**:
- [ ] Each brief contains all fields required by contract (see `references/brief-contract.md`)
- [ ] Each issue body contains `Task Key / Spec Revision / Approved Revision / Current Stage / Workflow Entry State / Completion Basis / Code Publication State`
- [ ] Structured event sequences in issue comments are continuous (1, 2, 3, ...)
- [ ] `Spec Revision == Approved Revision` throughout
- [ ] Parent-child issues correctly linked via `depends_on`

**Actual**:
- Core workflow fields present: Task Key, Issue Number, Spec Revision, Approved Revision, Current Stage, Next Stage, Workflow Entry State, Attempt Count, Max Retry Count, Code Publication State, Pass/Fail Outcome, Completion Basis, Code Ref, Dependencies, Work Breakdown, Source ✅
- `Spec Revision == Approved Revision` throughout ✅
- Brief section structure does not fully match `brief-contract.md`: uses `## Status` flat list instead of segmented `## Issue` + `## Classification` + `## Workflow State` ⚠️
- `Approval State` field missing ❌
- `Executability` not explicitly labeled (implicitly inferred from Type) ⚠️
- Issue comments contain **no structured YAML events** (`iron_tree: ...` block) because execution was manual with no skill auto-posting ❌
- Parent-child relation linked via brief Work Breakdown only, but GitHub issue body does not set `depends_on` ❌

**Root cause**:
1. Manual execution of `/task` → `/run` → `/verify` → `/checkup`, no skill automation generating standard-format briefs and structured issue comments
2. `brief-contract.md` defined section structure incompatible with manually created `## Status` list
3. GitHub API not used to set `depends_on` / parent-child links

**Fix**:
- This is a skill implementation gap, not a protocol design error. TC-1.2 expectations should be re-validated after skill implementation is complete.
- Current manually-created brief structure is sufficient for downstream `aggregate` and human review, but not machine-strict.
- **v3 input**: `task` skill must strictly follow `brief-contract.md` section structure when generating briefs; `run`/`verify`/`checkup` skills must post structured events to issue comments.

---

### TC-1.2b v2 Task Skill Regression Validation (vs. #21)

**Scenario**: Re-execute task-publish → run → verify → checkup using the templates/-driven v2 task skill flow, and compare against #21 (auth-core, v1 manual flow) to verify three gaps are closed.

**Vehicle**: `specs/task-search.md` → issues #26~28

**Three gap validations**:

| Gap | #21 (v1 manual) | #27 (v2 template-driven) | Verdict |
|------|---------------|-------------------|------|
| **1. Brief section structure** | Uses `## Status` flat list, missing `Approval State`, `Executability`, etc. ❌ | Uses `templates/brief.md.tmpl`, contains all 16 required sections (Issue, Classification, Dependencies, Acceptance Criteria, Verification, Target Files, Work Breakdown, Workflow State, Manual Acceptance, Failure Context, Completion Handoff, Execution Summary, Verification Summary, Pass/Fail Outcome, Open Questions / Warnings, Source) ✅ | **CLOSED** |
| **2. Structured YAML events** | No `iron_tree:` block in issue comments ❌ | Issue #27 contains 4 continuous sequence events: seq1 `task_published`, seq2 `run_started`, seq3 `verify_passed`, seq4 `checkup_done`, covering full workflow lifecycle ✅ | **CLOSED** |
| **3. GitHub depends_on** | Parent-child relation only described in brief Work Breakdown, no dependency declaration in GitHub issue body ❌ | #28 issue body contains `Depends on #27`, GitHub natively recognizes and establishes cross-issue link ✅ | **CLOSED** |

**Additional validations (v1.4 new resolutions)**:

| Resolution | Validation method | Result |
|------|----------|------|
| Verify Anchor SHA | `verify_passed` event contains `verify_anchor_sha: 39bea0d` | ✅ Recorded |
| Execution Lock | This experiment was single-task serial, lock contention not triggered; `.mino/run.lock` mechanism to be validated after skill implementation | ⚠️ TBD |
| External Events | No external close scenario in this experiment, to be validated after reconcile skill implementation | ⚠️ TBD |

**Root cause**: None
**Fix**: v2 task skill templates and event mechanism cover all three gaps; gaps closed.

---

## Phase 2 — Natural Imperfection (common in real development)

> Each TC is labeled with **🌡️ Real-world frequency** — how often does this happen in real development.

### Run Phase

#### TC-2.1 First implementation fails verify
🌡️ **Real-world frequency: high (30%+)** — agent-written code often fails lint or test on first try.

**Scenario**: Agent implements `auth-core` but misses an import; verify stage compilation fails.
**Operation**: In the spec require "passwords must be hashed with bcrypt" but do not specify bcrypt version, making it easy for the agent to hit ESM/CJS compatibility issues.
**Expected**:
- [ ] verify outputs `fail_retryable`
- [ ] `Attempt Count` not incremented by verify (remains 1)
- [ ] `Failure Context` contains first 50 lines of compilation error
- [ ] Auto returns to run; attempt 2 receives Failure Context and corrects

**Actual / Root cause / Fix**:

#### TC-2.2 Uncommitted local changes in repo
🌡️ **Real-world frequency: medium** — forgetting to stash when switching issues is a classic mistake.

> ⚠️ **Protocol dependency**: Current `skills/run/SKILL.md` **does not define** dirty working tree pre-flight detection. This TC is **expected to fail** until that skill behavior is added; the failure itself is the conclusion. Fix direction is to add pre-flight rules to the run skill first, then re-run this TC.

**Scenario**: Before `/run issue-core`, README.md has uncommitted edits in the editor.
**Expected (target state)**: run pre-flight detects dirty working tree, either requires stash or auto-stashes + restores after completion, **must not let unrelated changes leak into the commit**.
**Expected (current state, confirm before running)**: Protocol undefined → run executes directly → unrelated changes bundled into commit. This result should **trigger patching of the run skill**, not directly mark TC fail.

**Actual / Root cause / Fix**:

#### TC-2.3 Retry budget exhausted
🌡️ **Real-world frequency: low, but severe when it happens** — agent genuinely doesn't know a domain (e.g. some niche ORM).

**Scenario**: Spec requires a framework feature the agent is unfamiliar with; 4 consecutive runs fail to fix it.
**Expected**: After 4th verify failure, mark `fail_terminal`, issue enters `blocked`, no more auto retry, requires human intervention.
**Key invariant**: `retryable iff Attempt Count <= Max Retry Count`

**Actual / Root cause / Fix**:

#### TC-2.4 Skip dependency and execute downstream directly
🌡️ **Real-world frequency: medium** — user forgets the order and directly `/run` a downstream issue.

**Scenario**: `auth-pages` depends on `auth-core`, but user runs `/run issue-pages` first.
**Expected**: run detects dependency not done, rejects execution, prompts which to run first.

**Actual / Root cause / Fix**:

---

### Verify Phase

#### TC-3.1 Tests pass but publication fails
🌡️ **Real-world frequency: medium** — VPN hiccup, token expired, remote URL changed.

**Scenario**: All tests pass, but `git push` fails due to network.
**Expected**:
- [ ] `Current Stage` stays `verify`, does not enter checkup
- [ ] `Code Publication State` stays `local_only`
- [ ] `Pass/Fail Outcome` not set
- [ ] `Attempt Count` unchanged
- [ ] Re-run `/verify issue-N` to retry publication, do not re-run tests

**Actual / Root cause / Fix**:

#### TC-3.2 Project has no test commands at all
🌡️ **Real-world frequency: high** — early projects, prototype code, documentation-only changes.

**Scenario**: todo-list temporarily removes the test script from `package.json`.
**Expected**: verify must not default to pass; must enter `pending_acceptance` and require human sign-off.

**Actual / Root cause / Fix**:

#### TC-3.3 Test error output explosion
🌡️ **Real-world frequency: medium** — snapshot diffs, large e2e, loop errors.

**Scenario**: Tests produce 5000 lines of error output.
**Expected**: `Failure Context` truncated to first 50 lines + `...(truncated)...` + last 20 lines, preserving critical information.

**Actual / Root cause / Fix**:

---

### Checkup Phase

#### TC-4.1 Publish fails during Accept
🌡️ **Real-world frequency: low**

**Scenario**: Task enters `pending_acceptance`, user `/checkup accept` but network drops.
**Expected**: Do not record acceptance, all states unchanged, emit `checkup_accept_publication_failed` event, retry next time.

**Actual / Root cause / Fix**:

#### TC-4.2 Mistakenly accept a task that should not be accepted
🌡️ **Real-world frequency: medium** — hand slipped.

**Scenario**: Execute `/checkup accept` on a task with `Current Stage: run`.
**Expected**: Reject execution, prompt that this task is not in `pending_acceptance`.

**Actual / Root cause / Fix**:

#### TC-4.3 Aggregate parent before all children are done
🌡️ **Real-world frequency: medium** — eager to close.

**Scenario**: Parent `auth-system` still has 1 child task blocked; user runs `/checkup aggregate`.
**Expected**: List incomplete child tasks, reject aggregation.

**Actual / Root cause / Fix**:

#### TC-4.4 `Close On Done: manual` behavior
🌡️ **Real-world frequency: low** — but common for bug-type tasks.

**Scenario**: Set `issue.close_on_done: manual` in config.
**Expected**: Issue stays open after task done, emit close-reminder comment.

**Actual / Root cause / Fix**:

---

### Spec Evolution

#### TC-5.1 Add an acceptance criterion mid-flight
🌡️ **Real-world frequency: high** — discovered during review / code review.

**Scenario**: DAG already approved, 2 child tasks ran, then spec adds "password must contain numbers".
**Expected**:
- [ ] Re-`/task` detects `Spec Revision` change
- [ ] Existing issue body shows `Spec Revision ≠ Approved Revision`
- [ ] Mark `reapproval_required`
- [ ] Do not auto-execute any run, wait for user re-approval

**Actual / Root cause / Fix**:

#### TC-5.2 Run the same spec twice
🌡️ **Real-world frequency: high** — idempotency verification, misoperation.

**Scenario**: `/task specs/auth-system.md` executed twice.
**Expected**: Second run detects existing open issue with same Task Key, skips creation, brief not overwritten.

**Actual / Root cause / Fix**:

#### TC-5.3 Vague composite task
🌡️ **Real-world frequency: high** — spec was vague from the start.

**Scenario**: Spec only says "add real-time collaboration", no acceptance criteria, no target files.
**Expected**: `task` explicitly marks `needs_breakdown` and **refuses to generate DAG**, requesting the user to provide more information.
**Anti-expected (must avoid)**: Rush to generate a seemingly reasonable but actually guessed DAG based on weak information — this violates the protocol's "no rushing" principle.

> Note: Automatic decomposition to sub-task granularity is too demanding for current agent capabilities and is not an acceptance criterion for this TC. Being able to recognize "insufficient information" is already a manifestation of protocol value.

**Actual / Root cause / Fix**:

---

### State Recovery

#### TC-6.1 All briefs deleted, rebuild from scratch
🌡️ **Real-world frequency: low, but severe when it happens** — machine swap, cache clear, accidental deletion.

**Scenario**: After partial workflow completion (`auth-core` done, `auth-pages` running), `rm -rf .mino/briefs/*`, then `/checkup reconcile`.
**Expected**: Replay structured events from issue comments by sequence, rebuild all briefs, state fully recovered (including Attempt Count, both Revision fields).

**Actual / Root cause / Fix**:

#### TC-6.2 Event sequence has gaps
🌡️ **Real-world frequency: low** — user manually deleted a comment.

**Scenario**: In sequence 1, 2, 3, delete 2, then reconcile.
**Expected**: Rebuild based on highest valid sequence, do not fail due to missing sequence, but should warn.

**Actual / Root cause / Fix**:

---

## Phase 3 — Protocol TBD (failure = v3 design input)

> The TCs in this section **are not testing whether skills are implemented correctly**, but **whether the protocol spec itself has been thought through**.
>
> **Process contract**:
> 1. Before running a TC, make a protocol decision on the topic (fill ✅ resolution in this section)
> 2. Run the TC, validate the decision can be implemented
> 3. **Resolution must be written back to `skills/references/iron-tree-protocol.md`** (or corresponding contract file), not allowed to stay only in this file — TEST_PLAN is not the protocol spec, just execution scaffolding

### TC-7.1 Issue closed externally ⭐ High priority
**Why important**: Directly impacts the **state consistency model** — the protocol must answer "who is the source of truth".
**Undefined**: User manually closes an open issue on GitHub; how should reconcile handle it?
**Candidate solutions**:
- A. Sync close brief, mark done
- B. Mark inconsistency, require human confirmation (**preference**: preserve human trust in GitHub, brief not overwritten by external actions)
- C. Re-open the issue
**Resolution**: ✅ B. When reconcile detects an issue was externally closed but no corresponding workflow `done` event exists:
- Do not auto-sync brief to `done` (preserve human trust in GitHub)
- Record `issue_closed` in brief's `External Event` field, mark `blocked` in `Workflow Entry State`
- Post `checkup_reconcile_external_close_detected` event in issue comment
- Require human confirmation before proceeding

→ Written to `iron-tree-protocol.md` § External Events

### TC-7.2 Two runs executed in parallel
**Why important**: Impacts concurrency model, but v1 can temporarily bypass with "explicitly prohibited".
**Undefined**: Two terminals simultaneously `/run issue-A` and `/run issue-B`; how to handle?
**Candidate solutions**:
- A. v1 explicitly prohibits (**preference**: use `.mino/run.lock` file lock, simple and reversible)
- B. Allow dependency-free tasks to run in parallel, with conflict detection
- C. Fully allow, user takes responsibility
**Resolution**: ✅ A. v1 explicitly prohibits parallel `run`.
- `run` checks `.mino/run.lock` before starting; if present, reject execution and show currently running task key
- Lock file contains task key and ISO timestamp
- `run` normally completes (including verify phase) or exits abnormally, then `run` skill cleans up the lock file
- v3 will re-evaluate parallel execution of dependency-free tasks (requires solving file conflicts and state races)

→ Written to `iron-tree-protocol.md` § Execution Lock

### TC-7.3 Code modified during verify ⭐ High priority
**Why important**: Directly impacts the **state consistency model** — the protocol must define "what snapshot does verify anchor to".
**Undefined**: verify runs for 10 minutes, during which user modifies code; which version counts?
**Candidate solutions**:
- A. Use commit SHA at verify start (**preference**: run must commit before verify can start, verify anchors to SHA)
- B. Snapshot at verify end
- C. Detect changes and abort verify
**Resolution**: ✅ A. `verify` uses the commit SHA at start time.
- `run` must commit before handing off to `verify` (`Code Publication State: local_only → published` completed before verify)
- `verify` records `Verify Anchor SHA` = current HEAD at start
- `verify` only validates committed code; working directory changes during verify are independent and do not affect verify results
- If `verify` fails (retryable), next `run` restarts based on that SHA or newer committed code
- `workflow-state-contract.md` already added `Verify Anchor SHA` field

→ Written to `iron-tree-protocol.md` § Verify Anchor

---

## Execution Cadence

| Round | Scope | Entry criteria | Exit criteria |
|---|---|---|---|
| **Round 1** | Phase 1 (TC-1.x) | Repo reset to baseline tag | TC-1.1 all ✅ |
| **Round 2** | Phase 2 (TC-2.x ~ TC-6.x) | Round 1 passed | Most ✅, failed items have root cause recorded |
| **Round 3** | Phase 3 (TC-7.x) | Make protocol decisions first | Resolutions written to `references/` |

**Per-round convention**:
- Immediately fill "Actual/Root cause/Fix" back into this file after each TC
- Commit the same file, leaving iterative traces through git history
- Push TEST_PLAN.md to main repo after Round 1 passes — it is itself an artifact

---

## Success Criteria

| Achievement | Meaning | Action |
|---|---|---|
| Phase 1 all pass | Protocol **ready**, can begin small-scale real usage | Write a dogfood blog post |
| Phase 1 + Phase 2 mostly pass | Protocol **production-ready** | Add "Validated scenarios" section to README |
| Phase 3 all have resolutions | Protocol **v3 complete** | **Resolutions must be written back to `references/iron-tree-protocol.md`**, TEST_PLAN only serves as execution trace |

**Minimum acceptance line: TC-1.1 all ✅.**
Everything else is a bonus.

---

## Appendix: Current Known Protocol Gap Snapshot

Migrated from old TEST_PLAN, serving as Phase 3 topic sources:

- TC-7.1 ← old TC-5.2 (external close handling)
- TC-7.2 ← old TC-6.2 (parallel execution strategy)
- TC-7.3 ← old TC-6.3 (verify mid-flight code change determinism)

Old "fault injection" framework specific TC numbers have been merged/renamed into Phase 1-3; original file preserved in git history.
