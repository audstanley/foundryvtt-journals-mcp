# Foundry VTT Journal MCP Server

An MCP (Model Context Protocol) server for reading and searching Foundry VTT journals.

## Features

- **MCP Server** - JSON-RPC 2.0 compliant server for journal access
- **MDX Export** - Export journals to Markdown with frontmatter
- **Permission-Aware** - Respects Foundry VTT user permissions
- **Full-Text Search** - Search across journal entries and pages
- **Statistics** - Get detailed statistics about journals

## Installation

```bash
go mod download
mage build
```

## Quick Start

### MCP Server

```bash
./fjm serve --world <world-name>
```

The server communicates via stdio and requires a `--world` parameter.

### MDX Export

```bash
./fjm mdx --world <world-name> --output <output-path>
```

Exports all journals from the specified world to Markdown files.

## Configuration

Environment variables:
- `FJM_WORLDS_PATH` - Path to worlds directory (default: `./worlds`)
- `FJM_USER` - Username for permission filtering

Config file (`config.yaml`):
```yaml
worlds_path: ./worlds
user: your_username
log_level: info
```

## MCP Tools

- `list_worlds` - List all worlds
- `list_entries` - List journal entries
- `get_entry` - Get entry with pages
- `list_pages` - List pages in entry
- `get_page` - Get single page
- `search_entries` - Search entries by name
- `search_pages` - Search page content
- `get_entry_stats` - Get statistics

## Project Phases

See [PHASES.md](PHASES.md) for the complete implementation plan.