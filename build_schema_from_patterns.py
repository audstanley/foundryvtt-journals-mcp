#!/usr/bin/env python3
"""
Build schema from raw field pattern analysis
Uses the data we already extracted about field occurrences
"""

import re
from pathlib import Path

def extract_fields_from_raw_data(file_path):
    """Extract field information from raw binary data patterns"""
    print(f"\n{'='*80}")
    print(f"RAW PATTERN ANALYSIS: {file_path}")
    print(f"{'='*80}")
    
    with open(file_path, 'rb') as f:
        data = f.read()
    
    print(f"File size: {len(data)} bytes")
    
    # Define field patterns to look for
    field_patterns = {
        '_id': (b'"_id"', 'Unique identifier'),
        'name': (b'"name"', 'Entry/page name/title'),
        'type': (b'"type"', 'Entry type (text, image, video, etc.)'),
        'folder': (b'"folder"', 'Parent folder ID'),
        'pages': (b'"pages"', 'Array of page IDs'),
        'sort': (b'"sort"', 'Sort order'),
        'ownership': (b'"ownership"', 'Permission settings'),
        'flags': (b'"flags"', 'Module/system flags'),
        'title': (b'"title"', 'Title display settings'),
        'text': (b'"text"', 'Text content data'),
        'image': (b'"image"', 'Image metadata'),
        'video': (b'"video"', 'Video metadata'),
        'markdown': (b'"markdown"', 'Markdown content'),
        'show': (b'"show"', 'Title visibility flag'),
        'level': (b'"level"', 'Title visibility level'),
        'content': (b'"content"', 'Text/video content'),
        'format': (b'"format"', 'Content format'),
        'caption': (b'"caption"', 'Image caption'),
        'controls': (b'"controls"', 'Video controls'),
        'volume': (b'"volume"', 'Video volume'),
        'src': (b'"src"', 'Source URL'),
        'system': (b'"system"', 'System-specific data'),
        'default': (b'"default"', 'Default permission value'),
        'sheetClass': (b'"sheetClass"', 'Sheet class for UI'),
        'compendiumSource': (b'"compendiumSource"', 'Compendium source'),
        'coreVersion': (b'"coreVersion"', 'Foundry core version'),
        'systemId': (b'"systemId"', 'System identifier'),
        'createdTime': (b'"createdTime"', 'Creation timestamp'),
        'modifiedTime': (b'"modifiedTime"', 'Last modified timestamp'),
        'lastModifiedBy': (b'"lastModifiedBy"', 'Last modifier user ID'),
    }
    
    # Analyze each pattern
    pattern_results = {}
    
    for field_name, (pattern_bytes, description) in field_patterns.items():
        count = data.count(pattern_bytes)
        if count > 0:
            # Find positions and analyze context
            positions = []
            pos = 0
            while True:
                pos = data.find(pattern_bytes, pos)
                if pos == -1:
                    break
                positions.append(pos)
                pos += 1
            
            if positions:
                # Get sample values
                samples = []
                for p in positions[:3]:
                    # Extract context around the pattern
                    start = max(0, p - 50)
                    end = min(len(data), p + 200)
                    context = data[start:end]
                    
                    # Try to extract the value
                    # Look for pattern after the field name
                    pattern_pos = context.find(pattern_bytes)
                    if pattern_pos != -1:
                        value_start = pattern_pos + len(pattern_bytes)
                        
                        # Skip past colon
                        while value_start < len(context) and context[value_start] not in [ord(':'), ord('['), ord('{')]:
                            value_start += 1
                        
                        # Get value
                        if value_start < len(context):
                            if context[value_start] == ord('['):
                                # Array value
                                depth = 1
                                array_end = value_start + 1
                                while array_end < len(context) and depth > 0:
                                    if context[array_end] == ord('['):
                                        depth += 1
                                    elif context[array_end] == ord(']'):
                                        depth -= 1
                                    array_end += 1
                                
                                array_data = context[value_start:array_end]
                                # Extract quoted strings
                                strings = re.findall(b'"([^"]*)"', array_data)
                                samples.append(f'Array with {len(strings)} strings: {strings[:3]}')
                            elif context[value_start] == ord('{'):
                                # Object value
                                depth = 1
                                obj_end = value_start + 1
                                while obj_end < len(context) and depth > 0:
                                    if context[obj_end] == ord('{'):
                                        depth += 1
                                    elif context[obj_end] == ord('}'):
                                        depth -= 1
                                    obj_end += 1
                                
                                obj_data = context[value_start:obj_end]
                                samples.append(f'Object ({len(obj_data)} bytes)')
                            elif context[value_start] == ord('"'):
                                # String value
                                end_quote = context.find(b'"', value_start + 1)
                                if end_quote != -1:
                                    samples.append(f'"{context[value_start+1:end_quote].decode("utf-8", errors="replace")}"')
                            elif context[value_start] in [ord('0'), ord('1'), ord('2'), ord('3'), ord('4'), 
                                                           ord('5'), ord('6'), ord('7'), ord('8'), ord('9')]:
                                # Numeric value
                                num_end = value_start
                                while num_end < len(context) and context[num_end] in [ord('0'), ord('1'), ord('2'), 
                                                                                       ord('3'), ord('4'), ord('5'), 
                                                                                       ord('6'), ord('7'), ord('8'), 
                                                                                       ord('9'), ord('-'), ord('.')]:
                                    num_end += 1
                                samples.append(f'Number: {context[value_start:num_end].decode("ascii")}')
                            elif context[value_start:value_start+5] in [b'true', b'false']:
                                samples.append(f'Boolean: {context[value_start:value_start+5]}')
                            elif context[value_start:6] == b'null':
                                samples.append('null')
                
                pattern_results[field_name] = {
                    'count': count,
                    'description': description,
                    'sample_values': samples
                }
                print(f"\n{field_name}: {count} occurrences")
                print(f"  Description: {description}")
                if samples:
                    print(f"  Sample values: {samples}")
    
    # Find entry types
    print("\n" + "="*80)
    print("ENTRY TYPE ANALYSIS")
    print("="*80)
    
    # Look for "type":" followed by values
    type_pattern = b'"type":"'
    pos = 0
    types_found = {}
    while pos < len(data):
        pos = data.find(type_pattern, pos)
        if pos == -1:
            break
        
        # Extract the type value
        start = pos + len(type_pattern)
        end = data.find(b'"', start)
        if end != -1:
            type_value = data[start:end].decode('utf-8', errors='replace')
            types_found[type_value] = types_found.get(type_value, 0) + 1
        
        pos = end + 1
    
    for type_name, count in sorted(types_found.items()):
        print(f"  {type_name}: {count} occurrences")
    
    # Find page references
    print("\n" + "="*80)
    print("PAGE REFERENCE ANALYSIS")
    print("="*80)
    
    # Look for "pages":[" patterns
    pages_pattern = b'"pages":['
    pos = 0
    page_arrays = []
    while pos < len(data):
        pos = data.find(pages_pattern, pos)
        if pos == -1:
            break
        
        # Extract array content
        start = pos + len(pages_pattern)
        depth = 1
        end = start
        while end < len(data) and depth > 0:
            if data[end] == ord('['):
                depth += 1
            elif data[end] == ord(']'):
                depth -= 1
            end += 1
        
        array_data = data[start:end-1]
        
        # Extract page IDs
        page_ids = re.findall(b'"([^"]*)"', array_data)
        
        if page_ids:
            page_arrays.append(page_ids)
            print(f"\nPage array with {len(page_ids)} pages:")
            if len(page_ids) <= 5:
                for pid in page_ids:
                    print(f"  - {pid.decode('utf-8', errors='replace')}")
            else:
                print(f"  First 5: {[pid.decode('utf-8', errors='replace') for pid in page_ids[:5]]}")
                print(f"  Last 2: {[pid.decode('utf-8', errors='replace') for pid in page_ids[-2:]]}")
        
        pos = end
    
    # Find key prefixes in raw data
    print("\n" + "="*80)
    print("RAW KEY STRUCTURE")
    print("="*80)
    
    key_patterns = [
        (b'journal.', 'Journal entry key'),
        (b'journal.pages.', 'Page data key'),
    ]
    
    for pattern, desc in key_patterns:
        positions = []
        pos = 0
        while True:
            pos = data.find(pattern, pos)
            if pos == -1:
                break
            positions.append(pos)
            pos += 1
        
        if positions:
            print(f"\n{desc} ('{pattern.decode()}'): {len(positions)} occurrences")
            
            # Show first 3 examples
            for p in positions[:3]:
                # Extract key portion
                start = max(0, p - 10)
                end = min(len(data), p + 80)
                key_data = data[start:end]
                
                # Clean up control chars for display
                key_display = ''.join(chr(b) if 32 <= b < 127 else '.' for b in key_data)
                print(f"  Offset {p}: {key_display[:70]}...")
    
    # Build comprehensive schema
    print("\n" + "="*80)
    print("COMPREHENSIVE SCHEMA")
    print("="*80)
    
    print("""
JOURNAL DATA STRUCTURE (based on field analysis):

1. JOURNAL ENTRIES (parent objects with pages)
   =============================================
   
   Required Fields:
   - _id: string (unique identifier, ~16 chars)
   - name: string (display name/title)
   - type: string (entry type, likely "journal" or similar)
   - pages: array of strings (page IDs)
   - sort: number (sort order)
   - ownership: object (permission settings)
   
   Optional Fields:
   - folder: string (parent folder ID, null if none)
   - flags: object (module/system flags)
   - system: object (system-specific data)
   
   Ownership Structure:
   - ownership.default: number (default permission level)
   - ownership.{user_id}: number (specific user permission)

2. PAGE ENTRIES (individual page content)
   =======================================
   
   Required Fields:
   - _id: string (unique identifier, ~16 chars)
   - name: string (page title)
   - type: string (page content type)
     - "text": text content
     - "image": image content
     - "video": video content
   
   Optional Fields:
   - folder: string (parent folder ID)
   - sort: number
   - ownership: object
   - flags: object
   - system: object
   
   Title Display Settings:
   - title.show: boolean
   - title.level: number
   
   Content by Type:
   
   TEXT PAGES:
   - text.format: number (content format)
   - text.content: string (HTML/text content)
   - markdown: string (markdown content, empty if using HTML)
   
   IMAGE PAGES:
   - image.caption: string
   - image.src: string (URL)
   
   VIDEO PAGES:
   - video.controls: boolean
   - video.volume: number (0.0 - 1.0)
   - video.src: string (URL)
   
3. DATA ENCODING
   ===============
   
   The data is stored in LevelDB SST files with:
   - Binary compression (zlib/deflate)
   - Embedded control characters in JSON strings
   - Varint-encoded key lengths
   - Key format: "journal." or "journal.pages." prefix
   - Values are compressed JSON objects

4. PAGE ID FORMAT
   ================
   
   Page IDs appear to be:
   - Base64-like encoding
   - ~16 character length
   - Alphanumeric with possible special chars
   - Examples: "r6Y3buhLpmY0lQel", "ZNk6nVzTlEobjGQ9"

5. HIERARCHY STRUCTURE
   ====================
   
   Journal Entry (parent)
   ├─ _id: unique ID
   ├─ name: journal title
   ├─ pages: [page_id1, page_id2, ...]
   └─ ownership: permissions
   
   Page (leaf)
   ├─ _id: unique ID
   ├─ name: page title
   ├─ type: "text", "image", or "video"
   ├─ text/image/video: content data
   └─ title: display settings
""")

if __name__ == '__main__':
    files = [
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000376.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/journal/001167.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/journal/000171.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000361.ldb',
    ]
    
    for file_path in files:
        if Path(file_path).exists():
            extract_fields_from_raw_data(file_path)
        else:
            print(f"File not found: {file_path}")