# obsidian-cli

Standalone CLI for Obsidian vaults, designed for headless usage without the Obsidian desktop app.

## Features (POC Milestones 1-5)

- Vault bootstrap and status
- Note lifecycle (create/get/list/append/prepend/delete/move)
- Frontmatter properties with unknown-field preservation
- Tags from frontmatter and inline `#tag`
- Search with ripgrep (`rg`) backend for text + native tag/property search
- Wikilink parsing and backlink index
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
./obsidian-cli --vault /path/to/vault prop get project-plan.md status
./obsidian-cli --vault /path/to/vault prop list project-plan.md

# Tags
./obsidian-cli --vault /path/to/vault tag list
./obsidian-cli --vault /path/to/vault tag search project

# Search
./obsidian-cli --vault /path/to/vault search "meeting notes"
./obsidian-cli --vault /path/to/vault search --tag project
./obsidian-cli --vault /path/to/vault search --prop status=active

# Links and backlinks
./obsidian-cli --vault /path/to/vault links list project-plan.md
./obsidian-cli --vault /path/to/vault links backlinks project-plan.md --index

# Move note with link rewrites
./obsidian-cli --vault /path/to/vault note move project-plan.md archive/project-plan.md --dry-run
./obsidian-cli --vault /path/to/vault note move project-plan.md archive/project-plan.md
```

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
