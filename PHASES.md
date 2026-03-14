# Foundry VTT Journal MCP - Implementation Phases

## Overview
This document tracks the implementation phases for the FVTT Journal MCP Server project.

---

## Phase 1: Core Infrastructure ✅ COMPLETE

- [x] Project setup and structure
- [x] LevelDB reader for Foundry VTT data
- [x] Journal schema parsing
- [x] Basic MCP server skeleton
- [x] Configuration system
- [x] CLI with Cobra (simplified - only --worlds flag)

**Status:** Complete
**Test Coverage:** leveldb 84.7%, journal 20.6%

---

## Phase 2: HTML to Markdown Converter ✅ COMPLETE

### Implementation
- [x] Fix existing converter bugs
- [x] Add golang.org/x/net/html dependency
- [x] Implement tree-based parser
- [x] Integrate tree walker as primary converter with regex fallback

**Status:** COMPLETE
**Next:** Phase 3 - NDJSON Support & Unified Search

---

## Phase 3: NDJSON Support & Unified Search 🔄 IN PROGRESS

### Core Features - COMPLETED ✅
- [x] **NDJSON Reader** - Parse .db files (Actors, Items, Journals)
- [x] **Unified Repository** - Combine LevelDB + NDJSON
- [x] **Search Results Structure** - SearchResult with source tagging
- [x] **MergeSearchResults** - Deduplication by UUID
- [x] **Repository Search Methods** - SearchAll, SearchNDJSON, GetNDJSONByID

### MCP Tools - COMPLETED ✅
- [x] **search_all** - Unified search across LevelDB + NDJSON
- [x] **search_compendium** - NDJSON-only search with entity_type filter
- [x] **search_journals** - LevelDB journal entries only
- [x] **search_journal_pages** - LevelDB page content only
- [x] **resolve_uuid** - Updated to auto-discover world
- [x] **resolve_uuids_from_content** - Updated to auto-discover world

### CLI Simplification - COMPLETED ✅
- [x] Removed `--name`/`-n` flags
- [x] Only `--worlds` flag required
- [x] Auto-discover all worlds in --worlds directory
- [x] All tools auto-select first available world

### Implementation Notes
- NDJSON files discovered in `worlds/{world}/data/`
- Whitelisted entity types: `actors`, `items`, `journals`
- Source tagging: "LevelDB" (front) vs "NDJSON" (back)
- Deduplication prioritizes LevelDB results
- All entity data searchable (name, type, system fields)

**Status:** 90% Complete
**Next:** Phase 4 - Testing & Validation

### Test Coverage
- leveldb: 84.7% ✅
- ndjson: New module (needs tests)
- journal: 20.6%

---

## Phase 4: Testing & Validation

### Planned Tasks
- [ ] Integration tests with real FVTT worlds
- [ ] Search accuracy validation
- [ ] Deduplication verification
- [ ] Performance testing with large datasets
- [ ] NDJSON reader unit tests
- [ ] Unified search integration tests

**Status:** Not started

---

## Phase 5: Export Enhancement (Optional)

### Future Considerations
- [ ] Export NDJSON entities as MDX files
- [ ] Include statistics in export
- [ ] Configurable output structure

**Status:** Not started - Lower priority

---

## Quick Start for New Users

### Running the MCP Server
```bash
# Start server - auto-discovers all worlds
./fjm serve --worlds ./worlds

# Server runs on stdio, configure your MCP client to connect
```

### Exporting Journals to MDX
```bash
# Export all worlds in --worlds directory
./fjm mdx --worlds ./worlds --output ./exports
```

### Using MCP Tools

#### Unified Search
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search_all","arguments":{"query":"goblin"}},"id":1}
```

#### Compendium Search
```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"search_compendium","arguments":{"query":"dragon","entity_type":"actors"}},"id":1}
```

---

## Implementation Summary

### Database Architecture
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
```json
{
  "id": "006L3VJldOkDOqrI",
  "name": "Gargoyle",
  "type": "Actor",
  "source": "NDJSON",
  "uuid": "Actor/006L3VJldOkDOqrI",
  "meta": { ... }
}
```

---

## Notes

- All phases should maintain 90%+ code coverage target
- Use golang.org/x/net/html for HTML parsing
- Custom @UUID{} format requires special handling in parser
- Permission model: Filter by username but include permission metadata
- NDJSON search covers all fields recursively (name, type, system.* )