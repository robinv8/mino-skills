---
name: checkup
description: |
  Diagnose and repair workflow health, finalize tasks, record manual acceptance,
  aggregate composite parents, and reconcile out-of-band activity. Drives every
  task to a final `done` state. Supports pre-flight, check, repair, reconcile,
  accept, aggregate, and finalize modes.
---

# Workflow Health Mechanic

Diagnose environment readiness, skill wiring, brief freshness, and source-task alignment. Record manual acceptance, aggregate composite parents, and finalize verified tasks. Repair safely, report what cannot be repaired, never silently mutate human content.

## Modes

The user may specify a mode. Default is `check`.

- `pre-flight <issue>` — validate environment for one task before `/run` executes. Halts on broken environment, never advances state.
- `check` — read-only inspection of `.mino/`, skill ecosystem, briefs. No mutation.
- `repair` — fix missing wiring and auto-repairable issues (re-create missing dirs, refresh stale brief metadata).
- `reconcile [<issue>]` — compare local briefs against source tasks; replay valid events to rebuild brief state; detect external close.
- `accept <issue>` — record human acceptance for a task currently in `pending_acceptance`. Publishes any local-only code first.
- `aggregate <issue>` — finalize a `composite` / `container` task once all required children are `done`.
- `finalize <issue>` — bind verified-completion evidence and transition to `done`. Invoked by the Loop Mode Decision Function (see `iron-tree-protocol.md`) when verify produced `pass + verified` but the task is not yet `done`.

## Pre-flight gate

`pre-flight` is a stable interface called by `/run` before it acquires the lock. It is the only mode that must never mutate state beyond writing one structured event when blocking.

Steps:

1. Confirm `.mino/` and `.mino/briefs/` exist; `gh` is authenticated; the brief for `<issue>` exists and is well-formed.
2. Run task-specific dependency checks (e.g. `node_modules` present, toolchain available, secrets present).
3. If everything is healthy, print `pre-flight ok issue-{N}` and exit 0. Do NOT post any comment or modify any file.
4. If something is broken:
   - Set `Workflow Entry State: blocked` in the brief.
   - Render `templates/event-checkup-preflight-blocked.yml.tmpl` with the next sequence number and post it as an issue comment.
   - Exit non-zero so `/run` halts before acquiring the lock.

## Check / Repair

Both modes do the inspection below. `check` is read-only; `repair` may auto-repair the items explicitly marked safe.

1. **Core checks**:
   - `.mino/` and `.mino/briefs/` exist (`repair` may `mkdir -p`).
   - `gh` authenticated.
   - Project- or global-scope skills directory accessible.
2. **Skill ecosystem scan**:
   - Discover installed skills across scopes: `.claude/skills/`, `.agents/skills/`, `~/.claude/skills/`, `~/.agents/skills/`.
   - Verify the four core skills are present: `task`, `run`, `verify`, `checkup`.
   - Report complementary skill coverage and gaps.
3. **Brief freshness** (`repair` only):
   - For each brief, ensure required sections exist per `brief-contract.md`. Missing sections may be appended with empty placeholders. Existing human content is never overwritten.

Neither mode mutates any workflow event or transitions any task. Print a concise health report at the end.

## Reconcile

`reconcile` (or `reconcile <issue>` for one task) refreshes brief state from canonical evidence and detects out-of-band activity.

1. List local briefs: `ls .mino/briefs/issue-*.md`.
2. List source tasks: `gh issue list --state all --limit 200`.
3. For each brief in scope:
   - Pull all issue comments and parse YAML events whose `task_key` and `approved_revision` match the brief.
   - Replay events in ascending `sequence` order to rebuild the canonical workflow state. Surgically replace `Workflow State`, `Pass/Fail Outcome`, and `Completion Handoff` sections from the resulting state. Never overwrite `Open Questions / Warnings`.
   - **External close detection**: if the source issue is `closed` but no `checkup_done` event exists for the active approved revision:
     - Set `Workflow Entry State: blocked` in the brief.
     - Render `templates/brief-section-external-event.md.tmpl` (event=`issue_closed`, source=`github`, action=`Investigate the close reason and either re-open the issue with /task or accept the close as final`) and surgically replace the `External Event` section.
     - Render `templates/event-checkup-reconcile-external-close.yml.tmpl` and post as comment.
     - Do NOT mark the task `done`. Do NOT auto-sync brief to a completed state.
4. Emit a `Pending Acceptance` subsection in the report listing every task still in `pending_acceptance` with the next manual action.

## Accept

`accept <issue>` records human acceptance for a task currently in `pending_acceptance`.

1. Confirm `Workflow Entry State: pending_acceptance`. If not, halt and report.
2. **Publish any local-only code first**:
   - Stage all changes except workflow-local cache:
     ```bash
     git add -A -- ':!.mino/briefs/' ':!.mino/locks/' ':!.mino/run.lock'
     ```
   - If nothing is staged, skip to step 3 with `Code Ref: not_applicable`, `Code Publication State: not_applicable`.
   - Otherwise commit with `[run] issue-{N}: {concise change summary}` and `git push`.
   - Capture the resulting `HEAD` SHA as `Code Ref`. Set `Code Publication State: published`.
   - **On commit or push failure**: do NOT record acceptance. Keep `Current Stage: verify`, `Next Stage: checkup`, `Workflow Entry State: pending_acceptance`, `Code Publication State: local_only`. Render `templates/brief-section-publication-failure.md.tmpl` into the `Failure Context` section. Render `templates/event-checkup-accept-publication-failed.yml.tmpl` and post as comment. Stop.
3. Render `templates/brief-section-acceptance-summary.md.tmpl` with reviewer name, ISO timestamp, code ref, and notes; surgically replace the `Verification Summary` section.
4. Surgically update brief metadata:
   - `Current Stage: checkup`
   - `Next Stage: done`
   - `Workflow Entry State: ready_to_start`
   - `Pass/Fail Outcome: pass`
   - `Completion Basis: accepted`
   - `Code Ref:` and `Code Publication State:` per step 2
5. Render `templates/event-checkup-accept-recorded.yml.tmpl` and post as comment.
6. Remove the `pending-acceptance` label from the issue.
7. Proceed to **Finalize** (below).

## Aggregate

`aggregate <issue>` finalizes a `composite` or `container` parent.

1. Confirm `Classification` is `composite` or `container`.
2. Resolve required child task keys from the brief `Work Breakdown` and `Dependencies` sections. Confirm every child brief has `Pass/Fail Outcome: pass` and `Current Stage: done`.
3. Render `templates/brief-section-aggregate-summary.md.tmpl` with the children evidence list (each line: `- {child_task_key} (issue-{N}): {completion_basis} @ {code_ref_or_not_applicable}`); surgically replace the `Verification Summary` section.
4. Surgically update brief metadata:
   - `Current Stage: checkup`
   - `Next Stage: done`
   - `Workflow Entry State: ready_to_start`
   - `Pass/Fail Outcome: pass`
   - `Completion Basis: aggregated`
   - `Code Publication State: not_applicable`
   - `Code Ref: not_applicable`
5. Render `templates/event-checkup-aggregate-recorded.yml.tmpl` and post as comment.
6. Proceed to **Finalize** (below).

## Finalize

`finalize <issue>` (also entered automatically by `accept` and `aggregate`) binds completion evidence and transitions the task to `done`.

1. Read the brief. Refuse to proceed unless ALL of:
   - `Pass/Fail Outcome: pass`
   - `Completion Basis` ∈ {`verified`, `accepted`, `aggregated`}
   - `Code Publication State` ∈ {`published`, `not_applicable`}
   - `Code Ref` is a SHA when `Code Publication State: published`, else `not_applicable`
   If any check fails, halt and report which precondition is missing. Do NOT mutate state.
2. Surgically update brief metadata:
   - `Current Stage: done`
   - `Next Stage: none`
3. Render `templates/event-checkup-done.yml.tmpl` with the bound `completion_basis`, `code_ref`, and `code_publication_state`; post as the next sequenced comment.
4. Issue closure:
   - If brief `Close On Done: auto` and the issue is still open: `gh issue close {N} --reason completed`.
   - If `Close On Done: manual`: leave the issue open and post:
     > [checkup] issue-{N}: Task done — awaiting manual verification
     >
     > To close this issue after verification, run:
     > ```
     > gh issue close {N} --reason completed
     > ```
   - If the issue was already closed externally before finalize ran, do NOT re-open and do NOT re-close. Note the pre-existing closure in the report.

## Sequence numbers

Events posted by checkup share the same sequence space as task / run / verify events. Before posting any event, fetch the current max sequence for the active `approved_revision` from the issue comments and use `max + 1`.

## Templates

All event YAML and brief sections come from `templates/`. Render via literal `{{ var }}` substitution; no conditionals.

- Events: `event-checkup-preflight-blocked.yml.tmpl`, `event-checkup-accept-recorded.yml.tmpl`, `event-checkup-accept-publication-failed.yml.tmpl`, `event-checkup-aggregate-recorded.yml.tmpl`, `event-checkup-done.yml.tmpl`, `event-checkup-reconcile-external-close.yml.tmpl`.
- Brief sections: `brief-section-acceptance-summary.md.tmpl`, `brief-section-aggregate-summary.md.tmpl`, `brief-section-external-event.md.tmpl`, `brief-section-publication-failure.md.tmpl`.

Brief edits are always surgical: replace only the named section header and its body up to the next `## ` header. Never touch `Open Questions / Warnings`, `Source`, or any other section not listed above.

## Constraints

- Do NOT create issues or re-open closed issues.
- Do NOT mark a task `done` outside `finalize` (or via `accept` / `aggregate` which delegate to `finalize`).
- Do NOT mark a task `done` during `pre-flight`, `check`, `repair`, or `reconcile`.
- Do NOT record manual acceptance against an unpublished code state.
- Do NOT bypass pre-flight failures — if the environment is broken, block execution.
- Do NOT close issues when `Close On Done: manual`.
- Do NOT stage or commit `.mino/briefs/`, `.mino/locks/`, or `.mino/run.lock` — these are local workflow cache and must not enter git history.
- Do NOT overwrite human-authored sections (`Open Questions / Warnings`, free-form notes inside `Acceptance Criteria`, etc.).
- Do NOT auto-sync a brief to `done` after detecting an external close — record `External Event` and stop.

## References

- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
- [../references/brief-contract.md](../references/brief-contract.md)
- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
