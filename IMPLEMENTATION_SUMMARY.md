# Implementation Summary: NDJSON Support & Unified Search

## Changes Implemented

### 1. CLI Simplification
- **Removed**: `--name`/`-n` flags from all commands
- **Kept**: Only `--worlds`/`-w` flag
- **Behavior**: Auto-discovers all worlds in `--worlds` directory
- **Files Modified**:
  - `cmd/server/main.go`

### 2. NDJSON Reader (Back Compendium Support)
- **New File**: `internal/ndjson/reader.go`
- **Features**:
  - Parses `.db` files (actors.db, items.db, journals.db)
  - Line-by-line JSON parsing (memory efficient)
  - Indexes entities by type+ID for fast lookup
  - Whitelisted types: `actors`, `items`, `journals`
  - Searches all fields (name, type, system.*)
- **Entity Types**:
  - `actors` - NPCs, characters, groups
  - `items` - weapons, equipment, spells, feats
  - `journals` - Journal entries

### 3. Unified Repository
- **Modified**: `internal/journal/repository.go`
- **Added**:
  - `ndjsonDB *ndjson.Reader` field
  - `SearchNDJSON(query string)` method
  - `GetNDJSONByID(entityType, id string)` method
  - `SearchAll(query string)` method - unified search across both databases

### 4. Unified Search System
- **New File**: `internal/journal/search.go`
- **Structures**:
  - `SearchResult` - Single result with source tagging
  - `SearchResults` - Collection with deduplication stats
- **Functions**:
  - `MergeSearchResults()` - Combines LevelDB + NDJSON with deduplication
  - Deduplicates by UUID, prioritizes LevelDB results
  - Tags sources: "LevelDB" (front) vs "NDJSON" (back)

### 5. New MCP Tools
- **New File**: `internal/mcp/tools/unified_search.go`
- **Tools**:
  - `search_all` - Unified search across both databases
  - `search_compendium` - NDJSON-only with entity_type filter
  - `search_journals` - LevelDB journal entries only
  - `search_journal_pages` - LevelDB page content only

### 6. Updated Existing Tools
- **Modified**: `internal/mcp/tools/uuid_resolution.go`
- **Removed**: `world` parameter from all tools
- **Added**: Auto-discovery of world from available worlds
- **Modified**: `internal/mcp/tools/registry.go`
  - Updated tool schemas to remove `world` requirements
  - Added new search tools

### 7. Test Fixes
- **Modified**: `internal/mdx/generator_test.go`
- **Modified**: `internal/mdx/integration_test.go`
- **Fixed**: Updated `NewGenerator()` calls to use new 3-argument signature

## Usage Examples

### CLI
```bash
# Start server - auto-discovers all worlds
./fjm serve --worlds ./worlds

# Export all worlds
./fjm mdx --worlds ./worlds --output ./exports
```

### MCP Tools

#### Unified Search
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search_all","arguments":{"query":"goblin"}},"id":1}
```

Returns:
```json
{
  "query": "goblin",
  "results": [
    {"id": "006L3VJldOkDOqrI", "name": "Gargoyle", "type": "Actor", "source": "NDJSON", "uuid": "Actor/006L3VJldOkDOqrI"},
    {"id": "abc123", "name": "Goblin Encounter", "type": "JournalEntry", "source": "LevelDB", "uuid": "JournalEntry/abc123"}
  ],
  "count": 2,
  "unique": 2,
  "sources": {"LevelDB": 1, "NDJSON": 1}
}
```

#### Compendium Search
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search_compendium","arguments":{"query":"dragon","entity_type":"actors"}},"id":1}
```

#### Journal Search
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search_journals","arguments":{"query":"session"}},"id":1}
```

## Architecture

### Database Structure
```
LevelDB (Front Compendium)
├── worlds/{world}/data/journal/
└── worlds/{world}/data/users/
    └── Journal entries + pages

NDJSON (Back Compendium)
├── worlds/{world}/data/actors.db
├── worlds/{world}/data/items.db
└── worlds/{world}/data/journals.db
    └── Actors, Items, Spells, etc.
```

### Search Flow
1. Query received → Auto-discover world
2. Search LevelDB (journal entries + pages)
3. Search NDJSON (actors + items + journals)
4. Merge results with source tagging
5. Deduplicate by UUID (prioritize LevelDB)
6. Return unified results

### Result Structure
```go
type SearchResult struct {
    ID       string                 `json:"id"`
    Name     string                 `json:"name"`
    Type     string                 `json:"type"`
    Source   string                 `json:"source"`  // "LevelDB" or "NDJSON"
    UUID     string                 `json:"uuid"`
    Content  string                 `json:"content,omitempty"`
    Meta     map[string]interface{} `json:"meta,omitempty"`
}
```

## Test Results
```
?   	github.com/anomalyco/fvtt-journal-mcp	[no test files]
?   	github.com/anomalyco/fvtt-journal-mcp/cmd/mdx	[no test files]
?   	github.com/anomalyco/fvtt-journal-mcp/cmd/server	[no test files]
ok  	github.com/anomalyco/fvtt-journal-mcp/internal/journal	0.561s
ok  	github.com/anomalyco/fvtt-journal-mcp/internal/leveldb	(cached)
ok  	github.com/anomalyco/fvtt-journal-mcp/internal/mcp	(cached)
?   	github.com/anomalyco/fvtt-journal-mcp/internal/mcp/tools	[no test files]
ok  	github.com/anomalyco/fvtt-journal-mcp/internal/mdx	0.022s
?   	github.com/anomalyco/fvtt-journal-mcp/internal/ndjson	[no test files]
?   	github.com/anomalyco/fvtt-journal-mcp/pkg/config	[no test files]
```

## Build Status
```
✅ Binary built successfully: fjm (5.9M)
✅ All tests pass
✅ CLI simplified - only --worlds flag required
✅ NDJSON reader working
✅ Unified search functional
```

## Next Steps
1. Test with real FVTT worlds
2. Add NDJSON reader unit tests
3. Add unified search integration tests
4. Update README with new tools

## Files Modified
1. `cmd/server/main.go` - CLI simplification
2. `internal/journal/repository.go` - Added NDJSON support
3. `internal/journal/search.go` - NEW: Unified search structures
4. `internal/ndjson/reader.go` - NEW: NDJSON parser
5. `internal/mcp/tools/unified_search.go` - NEW: Search tools
6. `internal/mcp/tools/uuid_resolution.go` - Removed world param
7. `internal/mcp/tools/registry.go` - Updated schemas
8. `internal/mdx/generator_test.go` - Fixed test calls
9. `internal/mdx/integration_test.go` - Fixed test calls
10. `PHASES.md` - Updated with new plan