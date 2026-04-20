---
name: verify
description: |
  Validate executed work against acceptance criteria and repository-native
  checks (build, test, lint). Produces pass/fail verdict with actionable
  failure context or manual acceptance steps. Use after run or when a
  task needs validation.
---

# Verification Gatekeeper

Validate executed work against explicit acceptance criteria and repository-native checks, then produce a clear pass/fail verdict.

## Workflow

1. **Confirm scope** — identify the task being verified and load its brief.
2. **Select verification command** — prefer an explicit override first, then auto-detect repository-native tooling:
   | Signature | Command |
   |-----------|---------|
   | `.mino/config.yml` | read `verify.command` override |
   | `project.yml` (XcodeGen) | `xcodebuild -scheme <name>` |
   | `Package.swift` | `swift test` / `swift build` |
   | `Mino.xcodeproj` / `*.xcodeproj` | `xcodebuild -scheme <name>` |
   | `package.json` | `npm test` / `npm run build` / `npm run lint` |
   | `Cargo.toml` | `cargo test` / `cargo build` / `cargo clippy` |
   | `pyproject.toml` / `setup.py` | `pytest` / `python -m pytest` |
   | `Makefile` | `make test` |
   - If multiple signatures exist, use the one matching the project's primary language unless `.mino/config.yml` overrides it.
   - If none exist, proceed to the manual-acceptance verdict.
3. **Run checks** — execute the detected verification commands:
   - Build: `xcodebuild`, `npm run build`, `cargo build`, etc.
   - Test: `xcodebuild test`, `npm test`, `cargo test`, etc.
   - Lint: `swiftlint`, `eslint`, `clippy`, etc.
4. **Compare to acceptance criteria** — check each criterion from the brief against observed results.
5. **Render verdict** — after rendering each verdict:
   - Update local brief state
   - Do NOT stage the brief file
   - Post a short issue summary plus a structured workflow event

   **If all checks pass**:
   > ✅ **verify passed** — issue-8
   > Build: success
   > Tests: 42 passed
   > Summary: {concise}

   Then:
   - Publish code first if relevant:
     - If code files changed, stage only non-brief code files, commit with `[run] issue-{N}: {concise change summary}`, `git push`, and capture the resulting `HEAD` SHA as `Code Ref`
     - If no code files changed, use `Code Publication State: not_applicable` and `Code Ref: not_applicable`
     - If commit or push fails, do NOT record success. Keep `Current Stage: verify`, `Next Stage: verify`, `Workflow Entry State: ready_to_start`, `Code Publication State: local_only`, leave `Pass/Fail Outcome` unset, persist the publication error in `Failure Context`, post a structured `verify_publication_failed` event, and do not change `Attempt Count`
   - Only after publication succeeds, update brief: `Current Stage: checkup`, `Next Stage: done`, `Workflow Entry State: ready_to_start`, `Code Publication State: published|not_applicable`, `Pass/Fail Outcome: pass`
   - Record `Completion Basis: verified` and `Code Ref`
   - Post a structured `verify_passed` event
   **If automated checks pass but human review is still required**:
   > ⏸️ **manual acceptance required** — issue-8
   > Steps:
   > 1. Launch the app
   > 2. Navigate to ...
   > 3. Verify ...
   > 4. Record acceptance with `/checkup accept issue-8`

   Then:
   - Write the detailed checklist to the brief `Manual Acceptance` section:
     - `Reason: verify passed, awaiting manual check`
     - `Checklist: ...`
     - `Action: Run /checkup accept issue-{N}`
   - Update brief: `Current Stage: verify`, `Next Stage: checkup`, `Workflow Entry State: pending_acceptance`
   - Leave `Pass/Fail Outcome` unset
   - Add the issue label `pending-acceptance` if available
   - Post a short issue summary comment:
     - Reason
     - Action to run `/checkup accept issue-{N}`
     - Note that the detailed checklist is stored in the local brief
   - Post a structured `verify_pending_acceptance` event

   **If no verification command is detected**:
   > 🟡 **manual acceptance required** — issue-8
   > Reason: no build/test/lint tooling found in repository
   > Action: add a `.mino/config.yml` with `verify.command`, or review manually and then run `/checkup accept issue-8`

   Then:
   - Write the detailed checklist to the brief `Manual Acceptance` section:
     - `Reason: no tooling detected`
     - `Checklist: ...`
     - `Action: Run /checkup accept issue-{N}`
   - Update brief: `Current Stage: verify`, `Next Stage: checkup`, `Workflow Entry State: pending_acceptance`
   - Leave `Pass/Fail Outcome` unset
   - Add the issue label `pending-acceptance` if available
   - Post a short issue summary comment:
     - Reason
     - Action to run `/checkup accept issue-{N}`
     - Note that the detailed checklist is stored in the local brief
   - Post a structured `verify_pending_acceptance` event

   **If checks fail and retryable**:
   > ❌ **verify failed (retryable)** — issue-8
   > Failure Context:
   > ```
   > {first 50 lines of error output}
   > ... (truncated, see full output above)
   > {last 20 lines of error output}
   > ```

   Then:
   - A failure on attempt `N` is retryable only if `N <= Max Retry Count`; with the default `Max Retry Count: 3`, failures on attempts `1`, `2`, and `3` remain retryable, and a failure on attempt `4` is terminal
   - Update brief: `Current Stage: run`, `Next Stage: verify`, `Workflow Entry State: ready_to_start`, `Pass/Fail Outcome: fail_retryable`
   - Persist `Failure Context`
   - Post a structured `verify_failed_retryable` event

   **If checks fail and terminal**:
   > 🚫 **verify failed (terminal)** — issue-8
   > Failure Context:
   > ```
   > {truncated exact output}
   > ```
   > Reason: {why it's unrecoverable or why attempt budget is exhausted}

   Then:
   - Update brief: `Current Stage: verify`, `Next Stage: none`, `Workflow Entry State: blocked`, `Pass/Fail Outcome: fail_terminal`
   - Persist `Failure Context`
   - Post a structured `verify_failed_terminal` event

6. **CI / PR integration (optional)** — if the repository uses GitHub Actions or PR checks and a relevant PR is known:
   - Run `gh pr checks` to surface CI status
   - Include CI failures in `Failure Context` if they are more authoritative than local checks

## Constraints

- Do NOT skip `Failure Context` on errors — capture exact output, truncated if necessary.
- Do NOT suggest manual acceptance without clear, actionable steps.
- Do NOT auto-pass a task when no verification tooling is found.
- Do NOT record `verify_passed` before code publication has succeeded.
- Do NOT fix the failure here — hand `Failure Context` back to `run` for self-correction.
- Keep success summaries compact. No prose for green checks.
- Use `Attempt Count` and `Max Retry Count` exactly as defined in the workflow contract.

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
