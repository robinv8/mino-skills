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
2. **Run checks** — execute repository-native verification commands:
   - Build: `xcodebuild`, `npm run build`, `cargo build`, etc.
   - Test: `xcodebuild test`, `npm test`, `cargo test`, etc.
   - Lint: `swiftlint`, `eslint`, `clippy`, etc.
   Use the project's standard tooling.
3. **Compare to acceptance criteria** — check each criterion from the brief against observed results.
4. **Render verdict**:

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

   **If checks fail and retryable**:
   > ❌ **verify failed (retryable)** — issue-8
   > Failure Context:
   > ```
   > {exact compiler error or test failure output}
   > ```
   > Retries so far: {N}/3

   **If checks fail and terminal**:
   > 🚫 **verify failed (terminal)** — issue-8
   > Failure Context:
   > ```
   > {exact output}
   > ```
   > Reason: {why it's unrecoverable}

## Constraints

- Do NOT skip `Failure Context` on errors — capture exact output.
- Do NOT suggest manual acceptance without clear, actionable steps.
- Do NOT fix the failure here — hand `Failure Context` back to `run` for self-correction.
- Keep success summaries compact. No prose for green checks.

## References

- [references/workflow-state-contract.md](references/workflow-state-contract.md)
- [references/iron-tree-protocol.md](references/iron-tree-protocol.md)
