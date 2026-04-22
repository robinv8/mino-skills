# Spec: Adopt Brief Quality — Regression

Validates that `/task adopt issue-N` produces a brief whose field-filling pattern matches a native `/task <spec>` publish brief (protocol v1.11).

## TC-11.1: High-quality bug issue

Given:
- OPEN issue #501 with title `NullPointerException in UserService.validateEmail on null input`
- Body contains:
  - Reproduction steps (3 steps)
  - Expected behavior: `should throw IllegalArgumentException with message "email must not be null"`
  - Actual behavior: `throws NullPointerException at UserService.java:142`
  - Mentioned files: `src/main/java/com/example/UserService.java`, `src/test/java/com/example/UserServiceTest.java`
- Labels: `bug`
- No qualifying comments needed

When operator runs `/task adopt issue-501` and approves.

Then the rendered brief at `.mino/briefs/issue-501.md` has:
- `type`: `bug`
- `acceptance_criteria_checklist`: ≥3 testable `- [ ] ...` items, each a verifiable statement (e.g. `- [ ] UserService.validateEmail(null) throws IllegalArgumentException with message "email must not be null"`), no feelings paraphrases
- `verification_steps`: contains reproduction steps from the issue, followed by `- Expected: ...` and `- Actual after fix: ...` lines
- `target_files_list`: lists `src/main/java/com/example/UserService.java` and `src/test/java/com/example/UserServiceTest.java`
- No placeholder text in `acceptance_criteria_checklist`, `verification_steps`, or `target_files_list`

## TC-11.2: Vague feature request

Given:
- OPEN issue #502 with title `Add dark mode support`
- Body: `It would be great if the app supported dark mode.`
- Labels: `enhancement`
- No qualifying comments

When operator runs `/task adopt issue-502`.

Then:
- Adopt-Step 6 produces a brief with `acceptance_criteria_checklist` containing exactly one line: `- [ ] _(insufficient detail — see Open Questions)_`
- `Open Questions / Warnings` section lists ≥3 specific gaps (e.g. `Q: Which UI surfaces should support dark mode?`, `Q: Should it follow system preference or require a manual toggle?`, `Q: What color palette / contrast requirements apply?`)
- Adopt-Step 5 halts and re-prompts for approval citing the questions (does not proceed to Step 6 until user resolves)

## TC-11.3: Bug with stack trace

Given:
- OPEN issue #503 with title `Crash on startup in DataLoader`
- Body contains a Java stack trace with frames:
  - `at com.example.DataLoader.load(DataLoader.java:45)`
  - `at com.example.App.main(App.java:12)`
  - `at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)`
- Labels: `bug`

When operator runs `/task adopt issue-503` and approves.

Then the brief's `target_files_list` includes:
- `src/main/java/com/example/DataLoader.java` (or the repo-relative path equivalent)
- Does NOT include `sun.reflect.NativeMethodAccessorImpl` (JDK internal, not project code)

## TC-11.4: Issue with maintainer comment refining scope

Given:
- OPEN issue #504 with title `Support OAuth2 login`
- Body proposes OAuth2 for GitHub, Google, and Twitter
- One comment by an OWNER: `Let's scope this to GitHub only for now. Google and Twitter will be tracked in follow-up issues.` (👍 by another MEMBER)

When operator runs `/task adopt issue-504` and approves.

Then the brief's `acceptance_criteria_checklist`:
- Contains testable items related to GitHub OAuth2 only
- Does NOT contain acceptance items for Google or Twitter OAuth2

## TC-11.5: Issue with non-collaborator noise comment

Given:
- OPEN issue #505 with title `Fix memory leak in CacheManager`
- Body describes a clear leak in `CacheManager.java`
- One comment by a NONE user: `This is not a leak, the JVM handles it. Just increase heap size.` (no endorsement from OWNER/MEMBER/COLLABORATOR)

When operator runs `/task adopt issue-505` and approves.

Then the brief's `acceptance_criteria_checklist` and `verification_steps` reflect the issue body's description of a memory leak in `CacheManager.java`, ignoring the NONE user's comment.

## TC-11.6: Issue body unchanged

Given any adopted issue from TC-11.1 ~ TC-11.5.

When the adoption completes.

Then `gh issue view {N} --json body` returns body text identical to the pre-adopt state. The adopt process NEVER calls `gh issue edit {N} --body` or otherwise mutates the issue body.

## TC-11.7: stage:task label removed on approval

Given:
- OPEN issue #506, not previously adopted

When operator runs `/task adopt issue-506` and approves.

Then the issue's label set is exactly `{iron-tree:adopted, stage:run}`. The label `stage:task` is NOT present.
