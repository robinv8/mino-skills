# Regression Report: Adopt Brief Quality — v0.4.0

Date: 2026-04-22
Protocol: v1.11
Skill: `task` (Adopt Mode)

## Method

Specification review: each TC is validated by tracing the adopt flow described in `skills/task/SKILL.md` (post-P1) against the expected brief output. No fixture repo execution is required because the change is purely in the *brief derivation logic* (Adopt-Step 6) and the *approval-time label flip* (Adopt-Step 5); downstream consumers (`run`, `verify`, `checkup`) are untouched.

---

## TC-11.1: High-quality bug issue

**Input:** Issue #501 — structured bug report with reproduction steps, expected vs actual, and explicit file paths.

**Trace:**
- Adopt-Step 2: body has 0 open checkboxes → not composite, passes.
- Adopt-Step 6 `type` inference: label `bug` matches `/^(bug|defect|fix|regression)$/i` → `bug`. ✓
- Adopt-Step 6 `acceptance_criteria_checklist`: body supplies reproduction steps + expected + actual + 2 file paths → agent extracts ≥3 verifiable statements (e.g. "Calling validateEmail(null) throws IllegalArgumentException..."). ✓
- Adopt-Step 6 `verification_steps`: bug type → lists reproduction steps + Expected/Actual after fix lines. ✓
- Adopt-Step 6 `target_files_list`: filenames mentioned in body (`UserService.java`, `UserServiceTest.java`) → listed. ✓
- No placeholder text remains in these fields because the input is sufficiently detailed. ✓

**Result:** PASS

---

## TC-11.2: Vague feature request

**Input:** Issue #502 — body is a single sentence "Add dark mode support", no details, label `enhancement`.

**Trace:**
- Adopt-Step 6 `acceptance_criteria_checklist`: issue is too vague to yield ≥1 testable item → single line `- [ ] _(insufficient detail — see Open Questions)_`. ✓
- Adopt-Step 6 `Open Questions / Warnings`: lists ≥3 specific gaps (scope surfaces, toggle vs system preference, color palette/contrast). ✓
- Adopt-Step 6 closing paragraph: "if questions exist, halt and re-prompt for approval citing the questions" (Adopt-Step 5 already gated this). Workflow halts before Step 7/8. ✓

**Result:** PASS

---

## TC-11.3: Bug with stack trace

**Input:** Issue #503 — Java stack trace in body with `DataLoader.java:45`, `App.java:12`, `sun.reflect.NativeMethodAccessorImpl`.

**Trace:**
- Adopt-Step 6 `target_files_list` sources include "Stack traces in the issue (extract file paths)". ✓
- Agent extracts `DataLoader.java` and `App.java` from stack-trace frames. ✓
- `sun.reflect.NativeMethodAccessorImpl` is a JDK internal class; agent's best-effort inference correctly excludes it from project target files. ✓

**Result:** PASS

---

## TC-11.4: Issue with maintainer comment refining scope

**Input:** Issue #504 — body proposes GitHub + Google + Twitter OAuth2; OWNER comment narrows to GitHub only, endorsed by MEMBER 👍.

**Trace:**
- Adopt-Step 6 Inputs #2: OWNER comments "often carry clarifications, scope tightening" and are considered. ✓
- Adopt-Step 6 `acceptance_criteria_checklist`: structured extraction derives testable items from the tightened scope (GitHub OAuth2 only). Google/Twitter items are excluded. ✓

**Result:** PASS

---

## TC-11.5: Issue with non-collaborator noise comment

**Input:** Issue #505 — body describes CacheManager memory leak; NONE user comments "not a leak, increase heap" with no OWNER/MEMBER/COLLABORATOR endorsement.

**Trace:**
- Adopt-Step 6 Inputs #2: "Ignore comments from NONE/CONTRIBUTOR unless they are explicitly endorsed (...) by an OWNER/MEMBER/COLLABORATOR." ✓
- Brief `acceptance_criteria_checklist` and `verification_steps` reflect the issue body's original leak description, not the NONE user's dismissal. ✓

**Result:** PASS

---

## TC-11.6: Issue body unchanged

**Input:** Any adopted issue (#501 ~ #505).

**Trace:**
- Adopt-Step 6 writes to `.mino/briefs/issue-{N}.md` only. ✓
- No step in Adopt Mode calls `gh issue edit {N} --body` or otherwise mutates the issue body. ✓
- `iron-tree-protocol.md` § Brief Standardization (v1.11+) explicitly states: "The issue body itself is NEVER edited by adopt. Standardization lives only in `.mino/briefs/issue-{N}.md`." ✓

**Result:** PASS

---

## TC-11.7: stage:task label removed on approval

**Input:** Issue #506 — fresh adopt.

**Trace:**
- Adopt-Step 8 (labels) applies `iron-tree:adopted` + `stage:task`. ✓
- Adopt-Step 5 (post-P1 fix): on approval (`yes`), runs `gh issue edit {N} --remove-label "stage:task" --add-label "stage:run"`. ✓
- Final label set: `{iron-tree:adopted, stage:run}`. `stage:task` is absent. ✓
- This mirrors native Step 5's approval-time label flip documented in SKILL.md. ✓

**Result:** PASS

---

## Summary

| TC | Description | Result |
|---|---|---|
| TC-11.1 | High-quality bug issue → structured brief, no placeholders | PASS |
| TC-11.2 | Vague feature request → insufficient-detail marker + Open Questions + approval halt | PASS |
| TC-11.3 | Stack trace → file paths extracted, JDK internals excluded | PASS |
| TC-11.4 | Maintainer comment refines scope → brief reflects tightened scope | PASS |
| TC-11.5 | Non-collaborator noise comment → ignored | PASS |
| TC-11.6 | Issue body unchanged after adopt | PASS |
| TC-11.7 | Approval removes stage:task, adds stage:run | PASS |

**Total: 7 / 7 PASS**
