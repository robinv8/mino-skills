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
2. **Auto-detect verification command** — infer the correct build/test/lint tool from project signatures (in priority order):
   | Signature | Command |
   |-----------|---------|
   | `project.yml` (XcodeGen) | `xcodebuild -scheme <name>` |
   | `Package.swift` | `swift test` / `swift build` |
   | `Mino.xcodeproj` / `*.xcodeproj` | `xcodebuild -scheme <name>` |
   | `package.json` | `npm test` / `npm run build` / `npm run lint` |
   | `Cargo.toml` | `cargo test` / `cargo build` / `cargo clippy` |
   | `pyproject.toml` / `setup.py` | `pytest` / `python -m pytest` |
   | `Makefile` | `make test` |
   | `.mino/config.yml` | read `verify.command` override |
   - If multiple signatures exist, use the one matching the project's primary language.
   - If none exist, proceed to "No command" verdict (see step 5).
3. **Run checks** — execute the detected verification commands:
   - Build: `xcodebuild`, `npm run build`, `cargo build`, etc.
   - Test: `xcodebuild test`, `npm test`, `cargo test`, etc.
   - Lint: `swiftlint`, `eslint`, `clippy`, etc.
4. **Compare to acceptance criteria** — check each criterion from the brief against observed results.
5. **Render verdict**:

   **If all checks pass**:
   > ✅ **verify passed** — issue-8
   > Build: success
   > Tests: 42 passed
   > Summary: {concise}

   **If automated checks pass but human review needed**:
   > ⏸️ **manual acceptance required** — issue-8
   > Steps:
   > 1. Launch the app
   > 2. Navigate to ...
   > 3. Verify ...

   **If no verification command detected**:
   > 🟡 **verify skipped** — issue-8
   > Reason: no build/test/lint tooling found in repository
   > Action: add a `.mino/config.yml` with `verify.command`, or proceed to manual acceptance

   **If checks fail and retryable**:
   > ❌ **verify failed (retryable)** — issue-8
   > Failure Context:
   > ```
   > {first 50 lines of error output}
   > ... (truncated, see full output above)
   > {last 20 lines of error output}
   > ```
   > Retry state: invoked by run (max 3 retries controlled by run)

   **If checks fail and terminal**:
   > 🚫 **verify failed (terminal)** — issue-8
   > Failure Context:
   > ```
   > {truncated exact output}
   > ```
   > Reason: {why it's unrecoverable}

6. **CI / PR integration (optional)** — if the repository uses GitHub Actions or PR checks:
   - Run `gh pr checks` to surface CI status
   - Include CI failures in Failure Context if they are more authoritative than local checks

## Constraints

- Do NOT skip `Failure Context` on errors — capture exact output, truncated if necessary.
- Do NOT suggest manual acceptance without clear, actionable steps.
- Do NOT fix the failure here — hand `Failure Context` back to `run` for self-correction.
- Keep success summaries compact. No prose for green checks.
- Do NOT count retries — `run` owns the retry budget (max 3). `verify` only reports its own output.

## References

- [../references/workflow-state-contract.md](../references/workflow-state-contract.md)
- [../references/iron-tree-protocol.md](../references/iron-tree-protocol.md)
