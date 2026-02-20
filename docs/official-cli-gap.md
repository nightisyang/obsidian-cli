# Official CLI Gap Analysis

Compared against: `docs/official-cli-commands.md` (115 official command/subcommand headings).

## Implemented In This Session (Official-Parity Areas)

- `daily`
- `daily:path`
- `daily:read`
- `daily:append`
- `daily:prepend`
- `open`
- `templates`
- `template:read` (implemented as `template read`)
- `template:insert` (implemented as `template insert`)
- `sync:status` (implemented as `sync status`)
- `tasks`
- `task`
- `plugins`
- `commands`
- `command`
- Heading navigation (`note get --heading`)
- Block reference read/write (`block get`, `block set`)
- `create template=<name>` parity via `note create --template`

## Existing Pre-Session Coverage

- Vault/status/init and core note CRUD/search/tag/link/property commands (POC command model)

## Remaining Gaps Versus Official Surface

The official Obsidian CLI includes many app-coupled command families not fully present in this native/headless POC yet, including:

- Bases (`bases`, `base:*`)
- Bookmarks (`bookmarks`, `bookmark`)
- File history (`diff`, `history*`)
- Full files/folders alias set (`file`, `files`, `folder`, `folders`, `read`, `append`, `prepend`, `rename`, etc. using official parameter grammar)
- Full links set (`unresolved`, `orphans`, `deadends`)
- Outline (`outline`)
- Full plugin lifecycle (`plugin:*`, restricted mode toggles)
- Full properties alias set (`aliases`, `properties`, `property:*` with official syntax)
- Publish (`publish:*`)
- Random notes (`random`, `random:read`)
- Full search aliases (`search:context`, `search:open`)
- Full sync operations (`sync`, `sync:history`, `sync:read`, `sync:restore`, `sync:open`, `sync:deleted`)
- Full tags alias set (`tags`, `tag` official syntax variants)
- Themes/snippets (`themes`, `theme:*`, `snippets*`, `snippet:*`)
- Unique notes (`unique`)
- Vault switching/list (`vaults`, `vault:open`)
- Web viewer (`web`)
- Wordcount (`wordcount`)
- Workspace/TUI controls (`workspace*`, `tabs`, `tab:open`, `recents`)
- Developer tool commands (`devtools`, `dev:*`, `eval`)
- Other app lifecycle commands (`help`, `version`, `reload`, `restart`) aligned to official semantics

## Notes

- This repo currently uses Cobra subcommand hierarchy (for example `sync status`) rather than official colon-form names (for example `sync:status`).
- App-only execution remains explicitly surfaced as requiring Obsidian app/API mode where native headless behavior is not feasible.
