# Foundry VTT Journal LevelDB Schema - Complete Analysis

Based on analysis of 4 LevelDB files from different worlds:
- `/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000376.ldb` (2.1 MB)
- `/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/journal/001167.ldb` (1.7 MB)
- `/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/journal/000171.ldb` (148 KB)
- `/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000361.ldb` (1.6 MB)

---

## 1. OVERALL DATA STRUCTURE

Foundry VTT stores journal data in LevelDB SST (Sorted String Table) format with the following characteristics:

### Storage Format
- **File Format**: LevelDB SST (Sorted String Table)
- **Compression**: zlib/deflate (binary compression)
- **Key Encoding**: Varint-prefixed length + UTF-8 string
- **Value Encoding**: Varint-prefixed length + type byte + compressed JSON
- **Data Encoding**: JSON strings may contain embedded control characters

### Key Structure
Keys follow this pattern:
- **Journal Entries**: `journal.{compendium_id}.{entry_id}`
- **Page Data**: `journal.pages.{compendium_id}.{page_id}`

Example keys found:
- `journal.GatheringJournal.JournalEntry.k32xaKrBYjEg1Qn6`
- `journal.pages.0DP7BMV5Atms1KVb.WaA8oWdDwiWQwHDE`

---

## 2. COMPLETE FIELD SCHEMA

### 2.1 Journal Entry Fields (Parent Objects)

| Field Name | Data Type | Required | Description | Example Values |
|------------|-----------|----------|-------------|----------------|
| `_id` | string | Yes | Unique identifier | `"1:a\x00\xf0^"`, `"9:T\x00h"`, `">M\x00\xf0O"` |
| `name` | string | Yes | Display name/title | `"Dragon's Tongue Schoolhouse"`, `"Winesap"`, `"The Lion's Den"` |
| `type` | string | Yes | Entry type | `"journal"` |
| `pages` | array of strings | Yes | Array of page IDs | `["r6Y3buhLpmY0lQel"]`, `["ZNk6nVzTlEobjGQ9", "lZ4kwZ7WuEOZJWsM"]` |
| `sort` | number | No | Sort order in journal | `0`, `400001` |
| `ownership` | object | No | Permission settings | `{default:0, user_id:3}` |
| `folder` | string/null | No | Parent folder ID | `"7ncctdXOeS93oiFU"`, `null` |
| `flags` | object | No | Module/system flags | `{}` or `{module-name: {...}}` |
| `system` | object | No | System-specific data | `{}` |

#### Ownership Object Structure
| Field Name | Data Type | Description | Example Values |
|------------|-----------|-------------|----------------|
| `default` | number | Default permission level | `0`, `-1` |
| `{user_id}` | number | Permission for specific user | `3` (Editor), `0` (Observer) |

---

### 2.2 Page Entry Fields (Leaf Objects)

| Field Name | Data Type | Required | Description | Example Values |
|------------|-----------|----------|-------------|----------------|
| `_id` | string | Yes | Unique identifier | `"r6Y3buhLpmY0lQel"`, `"0goNSF1lDx5DTMvO"` |
| `name` | string | Yes | Page title | `"Details"`, `"Image (3)"`, `"Enterence"` |
| `type` | string | Yes | Content type | `"text"`, `"image"`, `"video"`, `"pdf"` |
| `sort` | number | No | Sort order | `400000`, `0` |
| `ownership` | object | No | Permission settings | `{default:-1}` |
| `flags` | object | No | Module/system flags | `{}` |
| `system` | object | No | System-specific data | `{}` |
| `title` | object | No | Title display settings | `{show: true, level: 1}` |

#### Title Display Object
| Field Name | Data Type | Description | Example Values |
|------------|-----------|-------------|----------------|
| `show` | boolean | Display title | `true`, `false` |
| `level` | number | Visibility level | `1` |

---

### 2.3 Content Fields (By Type)

#### 2.3.1 Text Pages
| Field Name | Data Type | Description | Example Values |
|------------|-----------|-------------|----------------|
| `text` | object | Text content | `{format: 1, content: "<p>...</p>"}` |
| `text.format` | number | Content format | `1` (HTML) |
| `text.content` | string | HTML/text content | `"<p>A day of solemn reflection...</p>"` |
| `markdown` | string | Markdown content | `""` (empty if using HTML) |

#### 2.3.2 Image Pages
| Field Name | Data Type | Description | Example Values |
|------------|-----------|-------------|----------------|
| `image` | object | Image metadata | `{caption: "", src: "..."}` |
| `image.caption` | string | Image caption | `""` |
| `image.src` | string | Image source URL | `"worlds/dragons_guard/..."` |

#### 2.3.3 Video Pages
| Field Name | Data Type | Description | Example Values |
|------------|-----------|-------------|----------------|
| `video` | object | Video metadata | `{controls: true, volume: 0.5}` |
| `video.controls` | boolean | Show controls | `true` |
| `video.volume` | number | Default volume | `0.5` (0.0 - 1.0) |
| `video.src` | string | Video source URL | `null` |

---

### 2.4 Metadata Fields (System Data)

These fields appear in compendium/system data:

| Field Name | Data Type | Description | Example Values |
|------------|-----------|-------------|----------------|
| `compendiumSource` | string/null | Compendium source | `null` |
| `coreVersion` | string | Foundry core version | `"12.331"` |
| `systemId` | string | System identifier | `"dnd5e"` |
| `systemVersion` | string | System version | `"4.2.2"`, `"4.0.4"` |
| `createdTime` | number | Creation timestamp (ms) | `1744418975902` |
| `modifiedTime` | number | Last modified timestamp | `1755296813609` |
| `lastModifiedBy` | string | User ID of last modifier | `"F\x00\x7f"` |
| `duplicateSource` | string/null | Original source if duplicated | `null` |

---

## 3. DATA TYPES AND HIERARCHY

### 3.1 Complete Hierarchy

```
Journal Data (LevelDB)
├── Journal Entries (Parent Objects)
│   ├── _id: string (unique ID)
│   ├── name: string (display name)
│   ├── type: string ("journal")
│   ├── pages: [Page IDs...]
│   ├── sort: number
│   ├── ownership: object
│   │   ├── default: number (permission level)
│   │   └── {user_id}: number (user-specific permission)
│   ├── folder: string/null (parent folder)
│   ├── flags: object
│   └── system: object
│
└── Page Entries (Leaf Objects)
    ├── _id: string (unique ID)
    ├── name: string (page title)
    ├── type: string ("text", "image", "video", "pdf")
    ├── sort: number
    ├── ownership: object
    ├── flags: object
    ├── system: object
    ├── title: object
    │   ├── show: boolean
    │   └── level: number
    ├── Content (based on type):
    │   ├── text.format: number
    │   ├── text.content: string (HTML)
    │   ├── markdown: string
    │   ├── image.caption: string
    │   ├── image.src: string
    │   ├── video.controls: boolean
    │   └── video.volume: number
    └── src: string (source URL)
```

### 3.2 Entry Types Found

| Type | Occurrences | Description |
|------|-------------|-------------|
| `text` | 286+, 404+, 218+, 14+ | Text content pages (HTML) |
| `image` | 12+, 1+, 9+ | Image content pages |
| `video` | 62+, 4+, 79+, 6+ | Video content pages |
| `pdf` | 2+, 1+ | PDF content pages |
| `journal` | (implicit) | Journal entry containers |
| `shop`, `equipment`, `consumable` | 1+ each | Non-journal data (compendium) |

---

## 4. PAGE ID FORMAT

Page IDs follow this pattern:
- **Length**: ~16 characters (variable)
- **Encoding**: Base64-like alphanumeric
- **Format**: `{alphanumeric}{optional_special_chars}`
- **Examples**:
  - `"r6Y3buhLpmY0lQel"` (16 chars)
  - `"ZNk6nVzTlEobjGQ9"` (16 chars)
  - `"91hvFM2HdqSejsBE"` (16 chars)
  - `"bWyslGn8YNfrxsq\x01\x1c\xc0GGDZ1HmzmcY5wTK7"` (with embedded bytes)

---

## 5. EXAMPLE DATA STRUCTURES

### 5.1 Journal Entry Example

```json
{
  "folder": "7ncctdXOeS93oiFU",
  "name": "Dragon's Tongue Schoolhouse",
  "_id": "1:a\x00\xf0^",
  "pages": ["r6Y3buhLpmY0lQel"],
  "sort": 0,
  "ownership": {
    "default": 0,
    "gPjEbdVmI62vHlUr": 3
  },
  "flags": {},
  "system": {}
}
```

### 5.2 Text Page Example

```json
{
  "sort": 400000,
  "name": "Details",
  "_id": "lpKxVpaISNaPhCGW",
  "type": "text",
  "title": {
    "show": true,
    "level": 1
  },
  "text": {
    "format": 1,
    "content": "<p>A day of solemn reflection, ...</p>"
  },
  "image": {
    "caption": ""
  },
  "video": {
    "controls": true,
    "volume": 0.5
  },
  "src": null,
  "flags": {},
  "system": {},
  "ownership": {
    "default": -1
  }
}
```

### 5.3 Image Page Example

```json
{
  "sort": 400000,
  "name": "Image (3)",
  "_id": "0goNSF1lDx5DTMvO",
  "type": "image",
  "title": {
    "show": true,
    "level": 1
  },
  "image": {
    "caption": ""
  },
  "text": {
    "format": 1
  },
  "video": {
    "controls": true,
    "volume": 0.5
  },
  "src": "worlds/dragons_guard/...",
  "flags": {},
  "system": {},
  "ownership": {
    "default": 0
  }
}
```

---

## 6. IMPLEMENTATION NOTES FOR MCP SERVER

### 6.1 Reading Journal Data

1. **LevelDB Access**: Use LevelDB Python library (`leveldb` package)
2. **Key Filtering**: Filter keys starting with `journal.` or `journal.pages.`
3. **Decompression**: Values are zlib/deflate compressed
4. **JSON Parsing**: Clean binary data (remove control characters) before parsing JSON
5. **Key Structure**: 
   - Journal entries: `journal.*` (parent objects)
   - Page data: `journal.pages.*` (leaf objects with content)

### 6.2 Data Relationships

- Journal Entry → Pages (one-to-many via `pages` array)
- Page IDs are stored in journal entry's `pages` array
- Individual page content is stored as separate entries with `journal.pages.` prefix
- Use `_id` to correlate entries with their pages

### 6.3 Permission Model

- `ownership.default`: Default permission for all users
- `ownership.{user_id}`: User-specific permission override
- Permission levels: `0` (Observer), `3` (Editor), `-1` (No access)

---

## 7. FILE STATISTICS SUMMARY

| File | Size | Journal Entries | Pages | Entry Types |
|------|------|-----------------|-------|-------------|
| testworld/000376.ldb | 2.1 MB | 28 | 1,847+ | text, image, video, shop, equipment, consumable |
| dragons_guard/001167.ldb | 1.7 MB | 7 | 111 | text, image, video, pdf |
| the-iron-kingdoms/000171.ldb | 148 KB | 4 | 57 | text, pdf |
| testworld/000361.ldb | 1.6 MB | 632 | 0 (pages embedded) | text, image, video |

---

## 8. KEY OBSERVATIONS

1. **Embedded Control Characters**: JSON strings contain embedded control characters that must be cleaned before parsing
2. **Mixed Data Types**: Journal files may contain both journal entries AND compendium data (items, shops, etc.)
3. **Page Reference vs Content**: 
   - Journal entries reference pages via `_id` array
   - Page content is stored as separate `journal.pages.*` entries
4. **Flexible Content Types**: Pages can be text (HTML), image, video, or PDF
5. **Rich Metadata**: Extensive permission system, flags, and system-specific data

---

## 9. RECOMMENDED MCP SERVER API

```
GET /worlds/{world_id}/journal
  → Returns all journal entries

GET /worlds/{world_id}/journal/{entry_id}
  → Returns a specific journal entry with pages

GET /worlds/{world_id}/journal/pages/{page_id}
  → Returns a specific page's content

GET /worlds/{world_id}/journal/search?query={text}
  → Search journal entries by name
```

---

**Analysis Date**: March 12, 2026
**Total Files Analyzed**: 4
**Total Fields Identified**: 28+
**Entry Types**: 5+ (journal, text, image, video, pdf, and more)