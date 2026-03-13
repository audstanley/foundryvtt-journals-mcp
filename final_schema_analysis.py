#!/usr/bin/env python3
"""
Final comprehensive LevelDB journal schema analyzer
"""

import struct
import zlib
import json
import re
from pathlib import Path
from collections import defaultdict

def clean_json_string(s):
    """Remove non-printable characters from a JSON string while preserving valid content"""
    # Replace control characters except common whitespace
    result = []
    for char in s:
        if ord(char) < 32 and char not in '\t\n\r':
            result.append('?')
        else:
            result.append(char)
    return ''.join(result)

def extract_json_objects(data):
    """Extract JSON objects from binary data"""
    objects = []
    brace_depth = 0
    start = None
    
    for i, byte in enumerate(data):
        if byte == ord('{'):
            if brace_depth == 0:
                start = i
            brace_depth += 1
        elif byte == ord('}'):
            brace_depth -= 1
            if brace_depth == 0 and start is not None:
                obj_bytes = data[start:i+1]
                # Clean and try to parse
                try:
                    cleaned = clean_json_string(obj_bytes.decode('utf-8', errors='replace'))
                    # Fix any broken strings
                    cleaned = re.sub(r'[^{}:\[\],\s"\'\w.-]+', '', cleaned)
                    parsed = json.loads(cleaned)
                    objects.append({
                        'start': start,
                        'end': i,
                        'cleaned': cleaned[:3000],
                        'parsed': parsed
                    })
                except Exception as e:
                    pass
                start = None
    
    return objects

def analyze_journal_schema(file_path):
    """Analyze and report the complete schema for journal data"""
    print(f"\n{'='*80}")
    print(f"SCHEMA ANALYSIS: {file_path}")
    print(f"{'='*80}")
    
    with open(file_path, 'rb') as f:
        data = f.read()
    
    print(f"File size: {len(data)} bytes")
    
    # Extract all JSON objects
    json_objects = extract_json_objects(data)
    print(f"\nFound {len(json_objects)} JSON objects")
    
    # Analyze structure
    if not json_objects:
        print("No JSON objects found!")
        return
    
    # Collect all field names and their types
    field_info = defaultdict(set)
    entry_types = set()
    page_ids_found = []
    
    for obj in json_objects:
        parsed = obj['parsed']
        
        if isinstance(parsed, dict):
            # Track entry types
            if 'type' in parsed:
                entry_types.add(parsed['type'])
            
            # Track fields
            for key, value in parsed.items():
                field_info[key].add(type(value).__name__)
                
                # Track pages
                if key == 'pages' and isinstance(value, list):
                    page_ids_found.extend(value)
    
    # Print schema
    print("\n" + "="*80)
    print("FIELD SCHEMA")
    print("="*80)
    
    for field in sorted(field_info.keys()):
        types = field_info[field]
        print(f"\n{field}:")
        for t in types:
            print(f"  - {t}")
    
    # Show example entries by type
    print("\n" + "="*80)
    print("ENTRY TYPE EXAMPLES")
    print("="*80)
    
    entries_by_type = defaultdict(list)
    for obj in json_objects:
        parsed = obj['parsed']
        if isinstance(parsed, dict) and 'type' in parsed:
            entry_type = parsed['type']
            if len(entries_by_type[entry_type]) < 2:
                entries_by_type[entry_type].append(parsed)
    
    for entry_type, examples in sorted(entries_by_type.items()):
        print(f"\n--- Entry Type: '{entry_type}' ---")
        for i, example in enumerate(examples):
            print(f"\n  Example {i+1}:")
            print(f"    _id: {example.get('_id', 'N/A')}")
            print(f"    name: {example.get('name', 'N/A')[:100]}")
            if 'folder' in example:
                print(f"    folder: {example.get('folder', 'N/A')}")
            if 'pages' in example:
                pages = example.get('pages', [])
                print(f"    pages: {len(pages)} pages")
                if pages:
                    print(f"      First page IDs: {pages[:3]}")
            if 'sort' in example:
                print(f"    sort: {example.get('sort', 'N/A')}")
            if 'ownership' in example:
                ownership = example.get('ownership', {})
                print(f"    ownership.default: {ownership.get('default', 'N/A')}")
            if 'flags' in example:
                flags = example.get('flags', {})
                print(f"    flags keys: {list(flags.keys())[:5]}")
            
            # Show page-specific fields
            if entry_type == 'page':
                print(f"\n    Page-specific fields:")
                if 'text' in example:
                    text = example['text']
                    print(f"      text.format: {text.get('format', 'N/A')}")
                    if 'content' in text:
                        content = text['content'][:500]
                        # Clean for display
                        content_clean = re.sub(r'[^\x20-\x7E\n]+', '?', content)
                        print(f"      text.content (cleaned): {content_clean[:200]}")
                if 'image' in example:
                    print(f"      image.caption: {example['image'].get('caption', 'N/A')[:100]}")
                if 'video' in example:
                    video = example['video']
                    print(f"      video.controls: {video.get('controls', 'N/A')}")
                    print(f"      video.volume: {video.get('volume', 'N/A')}")
                if 'type' in example:
                    print(f"      page type: {example['type']}")
    
    # Analyze page entries specifically
    print("\n" + "="*80)
    print("PAGE ENTRIES DETAILED ANALYSIS")
    print("="*80)
    
    page_entries = []
    for obj in json_objects:
        parsed = obj['parsed']
        if isinstance(parsed, dict) and parsed.get('type') == 'page':
            page_entries.append(parsed)
    
    print(f"Found {len(page_entries)} page entries")
    
    if page_entries:
        # Analyze first page entry in detail
        example_page = page_entries[0]
        print("\nDetailed Page Structure Example:")
        print(json.dumps(example_page, indent=2, ensure_ascii=False)[:5000])
    
    # Page ID analysis
    print("\n" + "="*80)
    print("PAGE ID ANALYSIS")
    print("="*80)
    print(f"Total unique page IDs found: {len(set(page_ids_found))}")
    
    # Show page ID patterns
    if page_ids_found:
        sample_ids = page_ids_found[:10]
        print(f"\nSample page IDs:")
        for pid in sample_ids:
            print(f"  - {pid} (length: {len(pid)})")
    
    # Key structure patterns
    print("\n" + "="*80)
    print("KEY STRUCTURE PATTERNS")
    print("="*80)
    
    # Look for key patterns in raw data
    key_prefixes = ['journal.', 'journal.pages.']
    
    for prefix in key_prefixes:
        count = data.count(prefix.encode('utf-8'))
        if count > 0:
            print(f"'{prefix}' appears {count} times in raw data")
    
    # Show raw key examples
    print("\nRaw key structure examples:")
    # Find "journal." occurrences
    pos = 0
    for _ in range(5):
        pos = data.find(b'journal.', pos)
        if pos == -1:
            break
        
        # Extract context
        start = max(0, pos - 20)
        end = min(len(data), pos + 100)
        context = data[start:end]
        
        # Try to decode
        try:
            decoded = context.decode('utf-8', errors='replace')
            print(f"\n  Offset {pos}: ...{decoded[:80]}...")
        except:
            print(f"\n  Offset {pos}: {context[:80]}")
        
        pos += 1
    
    # Summary statistics
    print("\n" + "="*80)
    print("SUMMARY STATISTICS")
    print("="*80)
    print(f"Total JSON objects: {len(json_objects)}")
    print(f"Entry types found: {sorted(entry_types)}")
    print(f"Total fields across all entries: {len(field_info)}")
    print(f"Unique page IDs: {len(set(page_ids_found))}")
    print(f"Page entries: {len(page_entries)}")

if __name__ == '__main__':
    files = [
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000376.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/journal/001167.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/journal/000171.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000361.ldb',
    ]
    
    for file_path in files:
        if Path(file_path).exists():
            analyze_journal_schema(file_path)
        else:
            print(f"File not found: {file_path}")