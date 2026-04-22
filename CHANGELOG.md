# Changelog

## v0.5.1 (2026-04-22)

- **Renamed all 4 skills with `mino-` prefix** to prevent slash-command collisions in shared host palettes:
  - `task` → `mino-task` (slash command: `/mino-task`)
  - `run` → `mino-run` (slash command: `/mino-run`)
  - `verify` → `mino-verify` (slash command: `/mino-verify`)
  - `checkup` → `mino-checkup` (slash command: `/mino-checkup`)
- Renamed `skills/{task,run,verify,checkup}/` directories accordingly and updated SKILL.md `name:` frontmatter.
- Updated all documentation (README, TEST_PLAN, protocol references, SKILL.md cross-refs) to use the new slash command names.
- **Protocol field names unchanged**: commit prefixes (`[task]`, `[run]`, `[verify]`, `[checkup]`), event types (`task_published`, `run_*`, `verify_*`, `checkup_*`), and stage labels (`stage:task`, `stage:run`, `stage:verify`, `stage:done`) keep their bare names — they identify protocol roles, not user-facing commands.
- Existing v0.5.0 users: run `claude plugin update mino@mino-skills` (or `copilot plugin update mino`) and use the new `/mino-*` commands.

## v0.5.0 (2026-04-22)

- Plugin marketplace support for Copilot CLI, Claude Code, and Cursor.
- Added `plugin.json`, `.claude-plugin/plugin.json`, `.cursor-plugin/plugin.json`, and `.claude-plugin/marketplace.json`.
- Retained backward compatibility with vercel-labs `npx skills add` installation path.
- Restructured README install section: Plugin Marketplace commands listed first, vercel-labs CLI as alternative fallback.
