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
- [x] CLI with Cobra (single binary: serve + mdx commands)

**Status:** Complete
**Test Coverage:** leveldb 84.7%, journal 20.6%

---

## Phase 2: HTML to Markdown Converter ✅ COMPLETE

### Current Approach: Regex-based
**Issues identified:**
- Fragile with nested/complex HTML
- Doesn't handle malformed input well
- Custom @UUID{} links not supported
- Hard to maintain

### New Approach: Proper HTML Parsing
**Benefits:**
- Robust tree traversal
- Handles nested elements correctly
- Better error handling
- Easier to extend with custom formats

### Tasks Completed

- [x] Fix existing converter bugs:
  - [x] List processing preserves non-list content
  - [x] Inline formatting order (bold/italic before paragraphs)
  - [x] Line breaks handling
  - [x] Table separator format
- [x] **Add golang.org/x/net/html dependency**
- [x] **Implement tree-based parser**
  - [x] Parse HTML into DOM tree
  - [x] Recursive tree walker
  - [x] Block element handling (p, h1-h6, ul, blockquote)
  - [x] Inline element handling (b, i, strong, em, code, a)
  - [x] Image and video handling
  - [x] Implement @UUID{} link parsing and conversion
- [x] **Fix identified issues:**
  - [x] UUID link extraction - bounds error with multiple links
  - [x] Ordered list numbering - working correctly
  - [x] Table cell extraction - proper separators
  - [x] Complex content parsing - no slice bounds panic
- [x] Integrate tree walker as primary converter with regex fallback
- [x] All tests passing (52.8% mdx coverage, 84.7% leveldb, 20.6% journal)
- [x] Single binary builds successfully (fjm)

**Status:** COMPLETE - Tree walker successfully handles all test cases
**Next:** Phase 3 - MCP Tools Implementation

### Test Coverage
- leveldb: 84.7% ✅
- mdx: 52.8% (tree walker + regex fallback)
- journal: 20.6%

---

## Phase 3: MCP Tools Implementation 🔄 IN PROGRESS

### Core Tools - COMPLETED ✅
- [x] `list_worlds` - List all available worlds (traversal.go)
- [x] `list_entries` - List journal entries in a world (traversal.go)
- [x] `get_entry` - Get complete entry with all pages (traversal.go)
- [x] `list_pages` - List pages within an entry (traversal.go)
- [x] `get_page` - Get single page content (traversal.go)
- [x] `search_entries` - Search entry names by query (search.go)
- [x] `search_pages` - Search page content by query (search.go)
- [x] `get_entry_stats` - Get entry statistics (stats.go)
- [x] `resolve_uuid` - Resolve @UUID{} references (uuid_resolution.go)
- [x] Permission filtering by username (all tools support `user` param)

### Implementation Notes
- ✅ Tree walker integrated for HTML to Markdown conversion
- ✅ ExtractUUIDLinks leveraged for UUID reference extraction
- ✅ Full integration with journal repository for data access
- ✅ Permission model support in all traversal and search tools

**Status:** 100% COMPLETE - All core MCP tools implemented
**Next:** Phase 4 - MDX Export Enhancement

---

## Phase 4: MDX Export Enhancement

- [ ] Complete MDX export functionality
- [ ] Generate proper frontmatter
- [ ] Handle nested journal structure
- [ ] Export images/videos with proper paths
- [ ] Include statistics in export
- [ ] Configurable output structure

**Status:** Not started

---

## Phase 5: Advanced Features

- [ ] Full-text search index
- [ ] Caching layer for performance
- [ ] WebSocket support for real-time updates
- [ ] Batch operations
- [ ] Journal entry linking/referencing
- [ ] Custom field extraction

**Status:** Not started

---

## Phase 6: Testing & Optimization

- [ ] Integration tests with real FVTT worlds
- [ ] Load testing for large journals
- [ ] Performance optimization
- [ ] Memory profiling
- [ ] Documentation generation
- [ ] Error handling improvements

**Status:** Not started

---

## Phase 7: Production Deployment

- [ ] Docker containerization
- [ ] Systemd service setup
- [ ] Monitoring and logging
- [ ] Backup/restore functionality
- [ ] Version migration tools
- [ ] Release process automation

**Status:** Not started

---

## Notes

- All phases should maintain 90%+ code coverage target
- Use golang.org/x/net/html for HTML parsing (Phase 2)
- Custom @UUID{} format requires special handling in parser
- Permission model: Filter by username but include permission metadata