---
name: run
description: |
  Advance an approved task DAG through execution. Reads local briefs,
  respects dependency order, runs code changes, commits them, and
  hands off a stable SHA to verify. Stops on blocked or
  pending-acceptance states.
  Use when a task DAG has been approved and you want to execute it.
---

# Execution Engine

Read an approved task from local briefs, run any pre-flight check, acquire the global execution lock, modify the codebase, commit the result, and hand off a stable commit SHA to `verify`.

This skill is the only writer of `.mino/run.lock` and the only place where `Attempt Count` increments. All structured artifacts (events, brief section, lock file) are rendered from templates in `templates/`.

## Workflow

### Step 1: Resolve Target

- Accept either an issue locator (e.g., `issue-42`) or a `Task Key`
- Resolve to the canonical `Task Key` before any other step
- Load `.mino/briefs/issue-{N}.md` and the latest valid event sequence from the issue

### Step 2: Eligibility Check

Refuse to proceed and halt with a clear message if any of these are not true:

- `Approval State: approved`
- `Approved Revision == Spec Revision` (otherwise direct user to `/task` for re-approval)
- `Executability: executable` (skip `container` tasks; they decompose, not execute)
- `Workflow Entry State: ready_to_start`
- All `Depends On` task keys are in `done` state

### Step 3: Pre-flight (Internal)

Invoke `/checkup pre-flight issue-{N}` as a sub-step. If pre-flight marks the task `blocked`, do **not** proceed; let `checkup` own the `checkup_preflight_blocked` event and halt.

This step exists per Iron Tree Protocol § Required Capabilities and is referenced in § Decision Function notes — pre-flight is `run`'s internal gate, not a separate Loop Mode step.

### Step 4: Acquire Global Run Lock

The protocol guarantees serial `run` execution per repository.

1. Check whether `.mino/run.lock` exists:
   - If absent → proceed
   - If present and **not stale** (acquired within the last 2 hours) → refuse execution. Print the active task key, issue number, and `acquired_at` from the lock. Suggest: "wait for the active run to complete, or remove `.mino/run.lock` if you are sure it is dead."
   - If present and **stale** (`acquired_at` older than 2 hours) → ask the user explicitly: "A stale `.mino/run.lock` from {acquired_at} for {task_key} (issue-{N}) was found. Override? (yes/no)". Only proceed on `yes`.

2. Write `.mino/run.lock` from `templates/lock.yml.tmpl`. Lock file content is local-only; do NOT stage or commit it.

3. From this point on, **release the lock on every exit path** (success, halt, failure). Use a guarded structure if your runtime supports it.

### Step 5: Begin Attempt

Capture `attempt_count_before = Attempt Count` for potential rollback in Step 7.E.

1. Increment `Attempt Count` by 1.
2. Update brief Workflow State section:
   - `Current Stage: run`
   - `Next Stage: verify`
   - `Workflow Entry State: ready_to_start`
   - `Code Publication State: local_only` (provisional; finalized in Step 7)
3. Post issue comment with narrative + rendered `templates/event-run-started.yml.tmpl`:

   ```
   🚀 run started — issue-{N} — attempt {n} / {max}
   Target: {task_key}
   Plan: {one-line plan}

   {render templates/event-run-started.yml.tmpl}
   ```

### Step 6: Execute

Perform the actual work:

1. Read each file in the brief's `Target Files` section.
2. Apply changes guided by the brief's `Acceptance Criteria` and any `Failure Context` from a previous attempt (if retrying, attempt a different approach — never repeat the same fix).
3. Run any code-generation, scaffold, or fix-up commands the task requires.
4. Track changed files and commands run; both feed into the Execution Summary.

Print compact progress lines, one per substep:

```
[run] issue-{N}: reading 3 target files
[run] issue-{N}: applying edits to AuthService.ts
[run] issue-{N}: regenerating types
[run] issue-{N}: done; 4 files changed
```

### Step 7: Commit (or fail-publish)

Decide the commit path based on file change status.

#### 7.A No file changes

Some tasks are research-only or update only `.mino/briefs/` content. If `git status --porcelain -- ':!.mino/briefs/' ':!.mino/locks/'` returns empty:

- Skip commit
- Set `commit_sha_or_not_applicable = not_applicable`
- Set `Code Publication State = not_applicable`
- Proceed to Step 8

#### 7.B Files changed → commit

```bash
git add -A -- ':!.mino/briefs/' ':!.mino/locks/'
git commit -m "[run] issue-{N}: {concise change summary}"
```

The commit message format is fixed: `[run] issue-{N}: {summary}` where `{summary}` is one short imperative sentence.

If commit succeeds:

- Capture `commit_sha = git rev-parse HEAD`
- Set `Code Publication State = local_only` (verify will push and flip to `published`)
- Proceed to Step 8

#### 7.E Commit fails (pre-commit hook rejects, missing identity, signing failure, etc.)

Per protocol § Phase 4 and contract § run, commit failure must not consume retry budget.

1. **Roll back the attempt counter**: set `Attempt Count = attempt_count_before` (the value captured in Step 5).
2. Update brief sections (surgical replace):
   - `Failure Context` ← record the exact commit error output, the command that failed, the failed `Verify Anchor SHA` placeholder (use `none — commit refused`), and ISO timestamp. Use the same structure as `verify`'s failure-context section but mark `Failed Check: git commit`.
   - `Workflow State`:
     - `Current Stage: run`
     - `Next Stage: verify`
     - `Workflow Entry State: ready_to_start`
     - `Code Publication State: local_only`
3. Post issue comment with narrative + rendered `templates/event-run-commit-failed.yml.tmpl`:

   ```
   ⚠️ run commit failed — issue-{N}
   Reason: {short error message}
   Action: resolve the commit issue (hook, identity, signing, …) and re-run `/run issue-{N}` (no retry budget consumed).

   {render templates/event-run-commit-failed.yml.tmpl}
   ```
4. Release `.mino/run.lock`. Halt.

### Step 8: Mark Run Complete

Only reachable from 7.A or 7.B success.

1. Update brief sections:
   - `Execution Summary` ← render `templates/brief-section-execution-summary.md.tmpl`
   - `Workflow State`:
     - `Current Stage: verify`
     - `Next Stage: verify`
     - `Workflow Entry State: ready_to_start`
     - `Code Publication State: local_only` (or `not_applicable` from 7.A)
2. Post issue comment with narrative + rendered `templates/event-run-completed.yml.tmpl`:

   ```
   ✅ run completed — issue-{N} — attempt {n} / {max}
   Commit: {sha or no changes}
   Files: {n} changed
   Next: /verify issue-{N}

   {render templates/event-run-completed.yml.tmpl}
   ```

After the local state is updated and before releasing the lock, sync the GitHub stage label so humans see progress:

```
gh issue edit {N} --remove-label "stage:run" --add-label "stage:verify"
```

Label sync failure is a warning, not an error: log `stage_label_sync_failed: <reason>` in the run report and proceed. The local yml remains authoritative.

### Step 9: Release Lock & Hand Off

1. Remove `.mino/run.lock`.
2. Print: `Run /verify issue-{N} to validate the commit.`

### Aggregate Handoff (special case)

If the target task is composite/container and all required children are `done`, do **not** execute. Instead:

- Skip Steps 4–8 (no lock, no commit, no run events)
- Print: `All children of issue-{N} are done. Run /checkup aggregate issue-{N} to finalize the parent.`

### Stop Conditions

Halt the loop and report rather than continue if the chosen task reaches:

- `Workflow Entry State: blocked`
- `Workflow Entry State: pending_acceptance`
- `Current Stage: done`

## Templates

All artifact shapes are externalized; `run` MUST NOT generate freehand variations.

- `templates/event-run-started.yml.tmpl`
- `templates/event-run-completed.yml.tmpl`
- `templates/event-run-commit-failed.yml.tmpl`
- `templates/brief-section-execution-summary.md.tmpl`
- `templates/lock.yml.tmpl`

Variable syntax is `{{ variable_name }}`. Replace literally; do not introduce conditional logic in templates.

## Constraints

- Do NOT execute unapproved tasks.
- Do NOT execute when `Approved Revision != Spec Revision`.
- Do NOT acquire `.mino/run.lock` without writing it from `lock.yml.tmpl`.
- Do NOT proceed past a stale lock without explicit user confirmation.
- Do NOT run sibling tasks in parallel (V1 is serial; see protocol § Execution Lock).
- Do NOT push commits — `verify` owns push.
- Do NOT bypass git hooks with `--no-verify` or equivalent flags.
- Do NOT increment `Attempt Count` anywhere except Step 5; do NOT change it on commit failure (roll back to `attempt_count_before`).
- Do NOT commit `.mino/briefs/` or `.mino/locks/` files; the commit pathspec excludes both.
- Do NOT write `Pass/Fail Outcome` or `Completion Basis`. Only verify and checkup may set these.
- Do NOT invent fields in the YAML events; the schema is fixed by `workflow-state-contract.md`.
- Do NOT overwrite `Open Questions / Warnings` in the brief; replace target sections only.
- Always release `.mino/run.lock` on exit, including failure paths.
- Do NOT `push --force`, `reset --hard` past the remote tip, rebase or amend any pushed commit; use `git revert` to undo published work (see protocol § Multi-Agent Git Hygiene).
- Do NOT treat `gh issue edit` label-sync failures as fatal; the local event yml is authoritative.

## References

- [references/workflow-state-contract.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/workflow-state-contract.md)
- [references/iron-tree-protocol.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/iron-tree-protocol.md)
- [references/brief-contract.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/brief-contract.md)
