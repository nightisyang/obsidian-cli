# obsidian-cli

Standalone CLI for Obsidian vaults, designed for headless usage without the Obsidian desktop app.

## Features (POC Milestones 1-5)

- Vault bootstrap and status
- Note lifecycle (create/get/list/append/prepend/delete/move)
- Frontmatter properties with unknown-field preservation
- Tags from frontmatter and inline `#tag`
- Search with ripgrep (`rg`) backend for text + native tag/property search
- Search-content alias (`search-content <query>`)
- Graph retrieval helpers (`graph context`, `graph neighborhood`)
- Wikilink parsing and backlink index
- Daily notes (`daily`, `daily path/read/append/prepend`)
- Templates (`templates`, `template read/insert`, `note create --template`)
- Tasks and task updates (`tasks`, `task`)
- Heading and block addressing (`note get --heading`, `block get/set`)
- Obsidian URI open (`open <path>`)
- Vault-aware file/folder listing (`list [path]`)
- Agent contracts and schema export (`help --agent`, `schema`)
- Batch operations with rollback (`ops apply`)
- Native plugin/sync introspection (`plugins`, `commands`, `sync status`)
- JSON output mode for every command

## Requirements

- Go 1.22+
- `rg` (ripgrep) for full-text search (`obsidian-cli search <query>`)

## Install

```bash
export PATH=/usr/local/go/bin:$PATH
git clone https://github.com/nightisyang/obsidian-cli.git
cd obsidian-cli
go build -o obsidian-cli .
```

## Quick Start

```bash
# Initialize vault config
./obsidian-cli --vault /path/to/vault vault init /path/to/vault

# Create and read notes
./obsidian-cli --vault /path/to/vault note create "Project Plan" --tag project --kind note
./obsidian-cli --vault /path/to/vault note get project-plan.md
./obsidian-cli --vault /path/to/vault note append project-plan.md "Next steps"
./obsidian-cli --vault /path/to/vault note prepend project-plan.md "Summary"
./obsidian-cli --vault /path/to/vault note list

# Properties
./obsidian-cli --vault /path/to/vault prop set project-plan.md status active
./obsidian-cli --vault /path/to/vault prop delete project-plan.md status
./obsidian-cli --vault /path/to/vault prop get project-plan.md status
./obsidian-cli --vault /path/to/vault prop list project-plan.md

# Tags
./obsidian-cli --vault /path/to/vault tag list
./obsidian-cli --vault /path/to/vault tag search project

# Search
./obsidian-cli --vault /path/to/vault search "meeting notes"
./obsidian-cli --vault /path/to/vault search-content "meeting notes"
./obsidian-cli --vault /path/to/vault search --tag project
./obsidian-cli --vault /path/to/vault search --prop status=active

# Graph context for agent retrieval
./obsidian-cli --vault /path/to/vault --json graph context "meeting notes" --seed-limit 8 --depth 1
./obsidian-cli --vault /path/to/vault --json graph neighborhood project-plan.md --depth 2

# Vault file/folder discovery
./obsidian-cli --vault /path/to/vault list
./obsidian-cli --vault /path/to/vault list "Projects"

# Daily notes
./obsidian-cli --vault /path/to/vault daily
./obsidian-cli --vault /path/to/vault daily append "- [ ] Follow up"

# Templates
./obsidian-cli --vault /path/to/vault templates
./obsidian-cli --vault /path/to/vault template read Daily --resolve --title "Today"
./obsidian-cli --vault /path/to/vault template insert project-plan.md Daily

# Tasks
./obsidian-cli --vault /path/to/vault tasks --todo
./obsidian-cli --vault /path/to/vault task --ref "project-plan.md:12" --toggle

# Heading/block targeting
./obsidian-cli --vault /path/to/vault note get project-plan.md --heading "Next Steps"
./obsidian-cli --vault /path/to/vault block get project-plan.md abc123

# Open in Obsidian app
./obsidian-cli --vault /path/to/vault open project-plan.md --launch

# Links and backlinks
./obsidian-cli --vault /path/to/vault links list project-plan.md
./obsidian-cli --vault /path/to/vault links backlinks project-plan.md --index

# Move note with link rewrites
./obsidian-cli --vault /path/to/vault note move project-plan.md archive/project-plan.md --dry-run
./obsidian-cli --vault /path/to/vault note move project-plan.md archive/project-plan.md

# Agent context + schema
./obsidian-cli help --agent --format json
./obsidian-cli schema --format json

# Batch apply with rollback
cat > ops.json <<'JSON'
{
  "ops": [
    {"id": "1", "args": ["note", "append", "project-plan.md", "- [ ] Follow up"]},
    {"id": "2", "args": ["prop", "set", "project-plan.md", "status", "active"], "expect": {"ok": "true"}}
  ]
}
JSON
./obsidian-cli --vault /path/to/vault --json ops apply ops.json --rollback
```

## Agent Workflow

Use these as agent primitives:

- `help --agent [--format json] [--skill <name>]`: skill-injection command contracts.
- `schema [command...] --format json`: machine-readable command/flag schema.
- `print-default --path-only`: resolved vault root path.
- `list [path]`: deterministic folder/file discovery inside vault.
- `search-content <query>`: explicit content search entry point.
- `search` / `search-content --with-meta`: include retrieval metadata + warnings.
- `graph context` / `graph neighborhood`: relationship context packs with metadata + warnings.
- `ops apply <spec.json> [--rollback]`: batch execute command arrays from JSON.

Mutation safety flags:

- `--dry-run`: preview write behavior (supported on mutating commands).
- `--if-hash <sha256>`: optimistic concurrency guard for note writes.
- `--strict`: fail when warnings are present (agent guardrail mode).

## Global Flags

- `--vault <path>`: explicit vault root
- `--config <path>`: explicit config file
- `--mode <native|api|auto>`: mode selector (POC uses native backend)
- `--json`: machine-readable output envelope
- `--quiet`: reduce human output labels
- `--timeout <duration>`: command timeout value

## Exit Codes

- `0` success
- `2` validation error
- `3` not found
- `4` config error
- `1` generic error

When `--json` is set, failures include:

- `error.code`
- `error.reason`
- `error.message`
- `error.actionable_hint`

## Config Resolution

Config file resolution order:

1. `--config` path if provided
2. `<vault>/.obsidian-cli.yaml`
3. `~/.obsidian-cli.yaml`
4. built-in defaults

Example `.obsidian-cli.yaml`:

```yaml
vault_path: "."
mode_default: "auto"
api_base_url: "https://127.0.0.1:27124"
api_timeout: "5s"
templates_dir: ".obsidian/templates"
index_dir: ".obsidian-cli-index"
```
