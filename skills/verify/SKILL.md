---
name: verify
description: |
  Validate executed work against acceptance criteria and repository-native
  checks (build, test, lint). Produces pass/fail verdict with actionable
  failure context or manual acceptance steps. Use after run or when a
  task needs validation.
---

# Verification Gatekeeper

Validate executed work against explicit acceptance criteria and repository-native checks, then produce a deterministic pass/fail verdict bound to a specific commit (`Verify Anchor SHA`).

This skill consumes `run`'s output and feeds `checkup`. All structured artifacts (events, brief sections) are rendered from templates in `templates/` so downstream skills can rely on schema stability.

## Workflow

### Step 1: Confirm Scope

- Identify the task being verified (issue number or task key from the user)
- Load `.mino/briefs/issue-{N}.md` and the latest valid event sequence from the issue
- Verify `Approved Revision == Spec Revision`; if not, halt and direct user to `/task` for re-approval

### Step 2: Anchor The Verification

Record `Verify Anchor SHA = git rev-parse HEAD` **before** running any check.

This SHA is bound into every event and the verification summary. It guarantees the verdict refers to a specific committed state, not the current working tree.

If `git status --porcelain` shows uncommitted changes at this point, halt and direct user back to `/run` — `verify` evaluates committed code only. (`run` is responsible for committing before handoff.)

### Step 3: Select Verification Commands

Resolve commands in this order; the first hit wins:

1. **`.mino/config.yml` override** — read `verify.commands` (a list); if present, use exactly this list and skip auto-detection. The list may include any shell-runnable command.
2. **Brief override** — if the brief's `Verification` section contains an explicit shell command list (lines starting with `$ `), use that.
3. **Auto-detect repo-native tooling** by file signature:

   | Signature | Default commands |
   |-----------|------------------|
   | `package.json` (with `pnpm-lock.yaml`) | `pnpm install --frozen-lockfile`, `pnpm build`, `pnpm test`, `pnpm lint` |
   | `package.json` (with `yarn.lock`) | `yarn install --frozen-lockfile`, `yarn build`, `yarn test`, `yarn lint` |
   | `package.json` (otherwise) | `npm ci`, `npm run build`, `npm test`, `npm run lint` |
   | `pyproject.toml` / `setup.py` | `pytest` |
   | `Cargo.toml` | `cargo build`, `cargo test`, `cargo clippy -- -D warnings` |
   | `Package.swift` | `swift build`, `swift test` |
   | `Mino.xcodeproj` / `*.xcodeproj` | `xcodebuild -scheme <name> test` |
   | `Makefile` | `make test` |

   Skip commands whose script is not present (e.g., `pnpm lint` if `package.json` has no `lint` script). Do not invent a missing build/test target.

4. **Nothing detected** → route to `pending_acceptance` (Step 6.D).

### Step 4: Run Checks

Execute the selected commands sequentially. Capture stdout + stderr per command. Stop on the first failure.

When tooling integration with PRs is available and a PR is known, also surface `gh pr checks` output in the failure context if it is more authoritative than the local result.

### Step 5: Compare To Acceptance Criteria

Walk each item in the brief's `Acceptance Criteria` section:

- If the criterion is satisfied by an automated check, mark it as covered
- If a criterion has no automated coverage, route the verdict to `pending_acceptance` (Step 6.D), even if all automated checks pass

### Step 6: Render Verdict

Choose exactly one of A–E. Each writes the brief, posts an issue comment with narrative + the rendered event template, and never stages or commits `.mino/briefs/`.

Brief updates use surgical section replacement; preserve `Open Questions / Warnings` and any other human-authored content.

#### 6.A All automated checks passed AND all acceptance criteria covered

1. **Publish code first** if relevant:
   - If code files changed during `run`:
     - The commit was already created by `run`. `verify` only needs to push.
     - Run `git push`. If push fails, go to 6.E (publication failed) instead.
     - Capture `Code Ref = git rev-parse HEAD` after push.
     - Set `Code Publication State: published`.
   - If no code files changed:
     - Set `Code Ref: not_applicable`, `Code Publication State: not_applicable`.

2. **Update brief sections** (surgical replace):
   - `Verification Summary` ← render `templates/brief-section-verification-summary.md.tmpl`
   - `Workflow State`:
     - `Current Stage: checkup`
     - `Next Stage: done`
     - `Workflow Entry State: ready_to_start`
     - `Code Publication State: published | not_applicable`
   - `Pass/Fail Outcome` ← `pass`
   - `Completion Handoff`:
     - `Completion Basis: verified`
     - `Code Ref: {sha or not_applicable}`

3. **Post comment to issue** — narrative + rendered `templates/event-verify-passed.yml.tmpl`:

   ```
   ✅ verify passed — issue-{N}
   - Build: success
   - Tests: {n} passed
   - Lint: clean
   - Code Ref: {sha or not_applicable}

   {render templates/event-verify-passed.yml.tmpl}
   ```

#### 6.B Checks failed AND attempt budget remains

A failure on attempt `N` is retryable when `N <= Max Retry Count`. With default `Max Retry Count: 3`, attempts 1/2/3 may yield `fail_retryable`; attempt 4 must be terminal.

1. **Update brief sections**:
   - `Failure Context` ← render `templates/brief-section-failure-context.md.tmpl` with `pass_fail_outcome: fail_retryable`
   - `Workflow State`:
     - `Current Stage: run`
     - `Next Stage: verify`
     - `Workflow Entry State: ready_to_start`
   - `Pass/Fail Outcome` ← `fail_retryable`

2. **Do NOT change `Attempt Count`.** Only `run` increments it.

3. **Post comment** — narrative + rendered `templates/event-verify-failed-retryable.yml.tmpl`:

   ```
   ❌ verify failed (retryable) — issue-{N} — attempt {n} / {max}
   Failed check: {command}
   {first 50 lines of error output}
   ... (truncated, full output in Failure Context)
   {last 20 lines of error output}

   {render templates/event-verify-failed-retryable.yml.tmpl}
   ```

#### 6.C Checks failed AND attempt budget exhausted (or unrecoverable)

1. **Update brief sections**:
   - `Failure Context` ← render `templates/brief-section-failure-context.md.tmpl` with `pass_fail_outcome: fail_terminal`
   - `Workflow State`:
     - `Current Stage: verify`
     - `Next Stage: none`
     - `Workflow Entry State: blocked`
   - `Pass/Fail Outcome` ← `fail_terminal`

2. **Post comment** — narrative + rendered `templates/event-verify-failed-terminal.yml.tmpl`:

   ```
   🚫 verify failed (terminal) — issue-{N}
   Reason: {budget exhausted | unrecoverable error class}
   Failed check: {command}
   {truncated output}

   {render templates/event-verify-failed-terminal.yml.tmpl}
   ```

#### 6.D Manual acceptance required

Triggers:
- All automated checks passed but at least one acceptance criterion has no automated coverage
- No verification tooling could be detected
- An acceptance criterion explicitly requires human review (UI screenshot, perceptual quality, etc.)

1. **Update brief sections**:
   - `Manual Acceptance` ← render `templates/brief-section-manual-acceptance.md.tmpl` (write the actionable checklist here)
   - `Workflow State`:
     - `Current Stage: verify`
     - `Next Stage: checkup`
     - `Workflow Entry State: pending_acceptance`
   - Leave `Pass/Fail Outcome` unset; do not write `Completion Handoff` yet.

2. **Tag the issue** with the `pending-acceptance` label. Skip gracefully if the label is missing on this repo.

3. **Post comment** — short summary + action + rendered `templates/event-verify-pending-acceptance.yml.tmpl`:

   ```
   ⏸️ manual acceptance required — issue-{N}
   Reason: {one line}
   Action: Run `/checkup accept issue-{N}` after completing the checklist (stored in the local brief).

   {render templates/event-verify-pending-acceptance.yml.tmpl}
   ```

#### 6.E Publication failed (push or commit refused after checks passed)

This is reachable only from 6.A when `git push` (or any equivalent publication step) fails after automated checks have already passed.

1. **Do NOT record success.** Leave `Pass/Fail Outcome` and `Completion Handoff` unset.
2. **Update brief sections**:
   - `Failure Context` ← render `templates/brief-section-failure-context.md.tmpl` with the publication error and `pass_fail_outcome: null`
   - `Workflow State`:
     - `Current Stage: verify`
     - `Next Stage: verify`
     - `Workflow Entry State: ready_to_start`
     - `Code Publication State: local_only`

3. **Do NOT change `Attempt Count`.** Publication failure must not consume retry budget.

4. **Post comment** — narrative + rendered `templates/event-verify-publication-failed.yml.tmpl`:

   ```
   ⚠️ verify publication failed — issue-{N}
   Checks passed at SHA {anchor}, but publication failed.
   Error: {short message}
   Action: resolve push/auth issue, then re-run `/verify issue-{N}` (no retry budget consumed).

   {render templates/event-verify-publication-failed.yml.tmpl}
   ```

## Templates

All artifact shapes are externalized; `verify` MUST NOT generate freehand variations.

- `templates/event-verify-passed.yml.tmpl`
- `templates/event-verify-failed-retryable.yml.tmpl`
- `templates/event-verify-failed-terminal.yml.tmpl`
- `templates/event-verify-publication-failed.yml.tmpl`
- `templates/event-verify-pending-acceptance.yml.tmpl`
- `templates/brief-section-verification-summary.md.tmpl`
- `templates/brief-section-failure-context.md.tmpl`
- `templates/brief-section-manual-acceptance.md.tmpl`

Variable syntax is `{{ variable_name }}`. Replace literally; do not introduce conditional logic in templates.

## Constraints

- Do NOT skip recording `Verify Anchor SHA` — it is required by the protocol (v1.4 § Verify Anchor) and required in every event.
- Do NOT modify `Attempt Count`. Only `run` increments it.
- Do NOT auto-pass a task when no verification tooling is found — route to 6.D instead.
- Do NOT record `verify_passed` before code publication has succeeded; on push failure go to 6.E.
- Do NOT fix the failure here — hand `Failure Context` back to `run` for self-correction.
- Do NOT stage or commit `.mino/briefs/` or `.mino/locks/` — these are local workflow cache.
- Do NOT invent fields in the YAML events; the schema is fixed by `workflow-state-contract.md`.
- Do NOT overwrite `Open Questions / Warnings` in the brief; replace target sections only.
- Keep narrative summaries compact; the structured event is the machine source of truth.
- Do NOT `push --force`, `reset --hard` past the remote tip, rebase or amend any pushed commit; use `git revert` to undo published work (see protocol § Multi-Agent Git Hygiene).

## References

- [references/workflow-state-contract.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/workflow-state-contract.md)
- [references/iron-tree-protocol.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/iron-tree-protocol.md)
- [references/brief-contract.md](https://github.com/robinv8/mino-skills/blob/main/skills/references/brief-contract.md)
