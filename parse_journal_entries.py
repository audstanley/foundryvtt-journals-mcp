#!/usr/bin/env python3
"""
Advanced LevelDB parser for Foundry VTT journal data
Handles compression and extracts complete JSON structures
"""

import struct
import zlib
import json
import re
from pathlib import Path
from collections import defaultdict

def find_json_objects(data):
    """Find all JSON objects in binary data by matching braces"""
    json_objects = []
    brace_stack = []
    start_pos = None
    
    for i, byte in enumerate(data):
        if byte == ord('{'):
            if not brace_stack:
                start_pos = i
            brace_stack.append('{')
        elif byte == ord('}'):
            if brace_stack:
                brace_stack.pop()
                if not brace_stack and start_pos is not None:
                    json_str = data[start_pos:i+1]
                    try:
                        # Try to decode and parse
                        decoded = json_str.decode('utf-8', errors='replace')
                        parsed = json.loads(decoded)
                        json_objects.append({
                            'offset': start_pos,
                            'length': i - start_pos + 1,
                            'string': decoded,
                            'parsed': parsed
                        })
                    except:
                        pass
                    start_pos = None
    
    return json_objects

def decompress_data(data):
    """Try various compression formats"""
    results = []
    
    # Try standard zlib
    for offset in range(0, min(100000, len(data) - 10), 100):
        for header in [(0x78, 0x9c), (0x78, 0xda), (0x78, 0x01)]:
            if data[offset:offset+2] == bytes(header):
                try:
                    decompressed = zlib.decompress(data[offset:])
                    results.append(('zlib', offset, decompressed))
                except:
                    pass
    
    # Try raw deflate
    for offset in range(0, min(100000, len(data) - 10), 100):
        try:
            decompressed = zlib.decompress(data[offset:], -zlib.MAX_WBITS)
            results.append(('raw_deflate', offset, decompressed))
        except:
            pass
    
    return results

def parse_journal_entries(file_path):
    """Parse journal entries from LevelDB file"""
    print(f"\n{'='*80}")
    print(f"File: {file_path}")
    print(f"{'='*80}")
    
    with open(file_path, 'rb') as f:
        data = f.read()
    
    print(f"File size: {len(data)} bytes")
    
    # First, try to decompress and find JSON
    print("\n--- Attempting decompression ---")
    decompressed_results = decompress_data(data)
    
    for comp_type, offset, decompressed in decompressed_results:
        print(f"\n  Found {comp_type} compressed data at offset {offset}")
        print(f"  Compressed: {len(data) - offset} bytes -> Decompressed: {len(decompressed)} bytes")
        
        # Try to extract all JSON objects
        json_objects = find_json_objects(decompressed)
        print(f"  Found {len(json_objects)} JSON objects")
        
        # Show first few entries
        for i, obj in enumerate(json_objects[:3]):
            print(f"\n  JSON object #{i+1} at offset {obj['offset']}:")
            print(f"    Length: {obj['length']} bytes")
            
            # Try to pretty-print
            try:
                pretty = json.dumps(obj['parsed'], indent=2, ensure_ascii=False)[:2000]
                print(f"    Content preview:\n{pretty}")
            except Exception as e:
                print(f"    Error: {e}")
    
    # Also search for journal keys in raw data
    print("\n--- Searching for keys in raw data ---")
    
    # Pattern for potential key values
    key_patterns = [
        b'journal:',
        b'selling:',
        b'loot:',
        b'folder:',
    ]
    
    for pattern in key_patterns:
        positions = []
        pos = 0
        while True:
            pos = data.find(pattern, pos)
            if pos == -1:
                break
            positions.append(pos)
            pos += 1
        
        print(f"  Found {len(positions)} occurrences of pattern {pattern}")
        
        # Show context around first 3
        for pos in positions[:3]:
            start = max(0, pos - 50)
            end = min(len(data), pos + 200)
            context = data[start:end]
            print(f"    Offset {pos}: {context[:150]}...")
    
    return data

def analyze_page_references(file_path):
    """Find and analyze page references in journal entries"""
    with open(file_path, 'rb') as f:
        data = f.read()
    
    print(f"\n{'='*80}")
    print(f"Analyzing page references in {file_path}")
    print(f"{'='*80}")
    
    # Look for page ID patterns (like ["abc123", "def456"])
    # These appear after "pages":[ in the JSON
    
    # Find all occurrences of "pages":[
    pattern = b'"pages":['
    positions = []
    pos = 0
    while True:
        pos = data.find(pattern, pos)
        if pos == -1:
            break
        positions.append(pos)
        pos += 1
    
    print(f"Found {len(positions)} page array declarations")
    
    # Extract page arrays
    for i, pos in enumerate(positions[:5]):
        print(f"\n  Page array #{i+1} at offset {pos}:")
        # Find the array end
        start = pos + len(b'"pages":[')
        # Find matching ]
        depth = 1
        end = start
        while end < len(data) and depth > 0:
            if data[end] == ord('['):
                depth += 1
            elif data[end] == ord(']'):
                depth -= 1
            end += 1
        
        array_data = data[start-2:end]
        print(f"  Array content: {array_data[:300]}")
        
        # Try to extract page IDs
        try:
            decoded = array_data.decode('utf-8', errors='replace')
            # Extract quoted strings (page IDs)
            page_ids = re.findall(r'"([^"]+)"', decoded)
            print(f"  Page IDs found: {page_ids}")
        except:
            pass

if __name__ == '__main__':
    files = [
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000376.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/journal/001167.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/journal/000171.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000361.ldb',
    ]
    
    for file_path in files:
        if Path(file_path).exists():
            parse_journal_entries(file_path)
            analyze_page_references(file_path)
        else:
            print(f"File not found: {file_path}")