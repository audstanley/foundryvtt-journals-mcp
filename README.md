# Foundry VTT Journal MCP Server

An MCP (Model Context Protocol) server for reading and searching Foundry VTT journals.

## Overview

This tool provides:
- **MCP Server** - JSON-RPC 2.0 compliant server for journal data access
- **MDX Export** - Export journals to structured Markdown with comprehensive frontmatter
- **Permission-Aware** - Respects Foundry VTT user permissions and ownership
- **UUID Resolution** - Resolve `@UUID{Type.ID}{Text}` references to actual data
- **Nested Structure** - Preserve FVTT folder hierarchy in exports

## Requirements

- Go 1.21+
- Foundry VTT worlds with LevelDB journal data (`worlds/<world>/data/journal/`)

## Installation

```bash
# Clone repository
git clone https://github.com/anomalyco/fvtt-journal-mcp.git
cd fvtt-journal-mcp

# Build binary
go mod download
go build -o fjm ./cmd/server

# Or use mage for additional commands
go install github.com/magefile/mage@latest
mage build
```

## Quick Start

### Running the MCP Server

```bash
# Start server for a specific world
./fjm serve --world MyWorld

# Server runs on stdio, configure your MCP client to connect
```

### Exporting Journals to MDX

```bash
# Export all journals to Markdown
./fjm mdx --world MyWorld --output ./exports

# Output structure:
# ./exports/
#   MyWorld/
#     Entry Name/
#       Page 1.mdx
#       Page 2.mdx
```

With folder support:
```bash
./exports/
  MyWorld/
    Campaign/Session 1/
      Session Notes.mdx
    Campaign/Session 2/
      Session Notes.mdx
```

## MCP Tools

All tools accept JSON parameters via stdio:

### Traversal Tools
- `list_worlds` - List all available worlds
- `list_entries` - List journal entries (with optional `user` for filtering)
- `get_entry` - Get entry with all pages
- `list_pages` - List pages within an entry
- `get_page` - Get single page content

### Search Tools
- `search_entries` - Search entry names by query
- `search_pages` - Search page content (returns snippets)

### Stats Tools
- `get_entry_stats` - Get world/entry statistics

### UUID Tools
- `resolve_uuid` - Resolve single `@UUID{}` reference
- `resolve_uuids_from_content` - Extract and resolve all UUIDs in content

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FJM_WORLDS_PATH` | `./worlds` | Path to Foundry VTT worlds directory |
| `FJM_USER` | - | Username for permission filtering |
| `FJM_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### Config File (`config.yaml`)

```yaml
worlds_path: ./worlds
user: your_username
log_level: info
```

## MDX Frontmatter

Exported pages include comprehensive YAML frontmatter:

```yaml
title: Session Notes
entry: Campaign Journal
type: text
page: abc123def456
sort: 100
entry_sort: 1000
folder: Campaign/Season 1
ownership:
  default: 2
  player1: 3
page_ownership:
  default: 2
uuid_references:
  - type: Item
    id: "ItemUUID123"
    display: "Sword"
    uuid: "Item/ItemUUID123"
  - type: Actor
    id: "ActorUUID456"
    display: "Goblin"
    uuid: "Actor/ActorUUID456"
siblings:
  - "Session 2"
  - "Session 3"
title_config:
  show: true
  level: 2
created: 1709123456000
modified: 1709987654000
```

## HTML to Markdown Conversion

The server uses a robust tree-based HTML parser (not regex) that handles:
- Block elements: paragraphs, headings, lists, blockquotes, tables
- Inline formatting: bold, italic, code
- Media: images, videos
- Custom `@UUID{}` references for Foundry VTT items/actors

## Project Phases

| Phase | Status | Description |
|-------|--------|-------------|
| 1 | ✅ Complete | Core Infrastructure |
| 2 | ✅ Complete | HTML to Markdown Converter |
| 3 | ✅ Complete | MCP Tools Implementation |
| 4 | 🔄 In Progress | MDX Export Enhancement |
| 5 | ⏳ Pending | Advanced Features |
| 6 | ⏳ Pending | Testing & Optimization |
| 7 | ⏳ Pending | Production Deployment |

See [PHASES.md](PHASES.md) for detailed task breakdown.

## Examples

### Listing Worlds
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_worlds","arguments":{}},"id":1}
```

### Searching Entries
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search_entries","arguments":{"world":"MyWorld","query":"goblin"}},"id":1}
```

### Resolving UUID
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"resolve_uuid","arguments":{"world":"MyWorld","type":"Item","id":"ItemUUID123"}},"id":1}
```

## Troubleshooting

### Journal Not Found
Ensure your `worlds_path` points to the directory containing Foundry VTT worlds. Each world should have:
- `worlds/<name>/data/journal/` - Journal entries
- `worlds/<name>/data/users/` - User data

### Permission Errors
Set `FJM_USER` or pass `user` parameter to tools for permission filtering.

### Build Errors
```bash
# Clean and rebuild
go clean
go mod download
go build -o fjm ./cmd/server
```

## Contributing

1. Check PHASES.md for current priorities
2. Create a new branch for your changes
3. Ensure all tests pass (`go test ./...`)
4. Maintain 90%+ code coverage where possible
5. Update documentation as needed

## License

MIT License - See LICENSE file for details.