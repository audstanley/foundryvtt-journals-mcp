#!/usr/bin/env python3
"""
Deep LevelDB SST file analyzer for Foundry VTT journal data
"""

import struct
import zlib
import json
import re
from pathlib import Path

def read_varint(data, offset):
    """Read a varint from binary data"""
    result = 0
    shift = 0
    while offset < len(data):
        byte = data[offset]
        offset += 1
        result |= (byte & 0x7f) << shift
        if not (byte & 0x80):
            break
        shift += 7
    return result, offset

def parse_leveldb_sst(data):
    """
    Parse LevelDB SST file format
    SST format:
    - Data blocks (key-value pairs, compressed)
    - Meta block (contains indexes)
    - Footer (contains meta block offset)
    """
    
    print(f"SST File Analysis")
    print(f"Total size: {len(data)} bytes")
    
    # The footer is at the end - 10 bytes
    # Format: data(8 bytes) + checksum(4 bytes) + type(1 byte)
    # Actually: data(offset) + type(1) = 9 bytes, but there's a checksum too
    # Standard LevelDB SST footer: 10 bytes total
    # [data (8 bytes)][reserved (1 byte)][type (1 byte)]
    
    if len(data) < 10:
        return None
    
    # Read the footer
    footer_offset = len(data) - 10
    footer_data = data[footer_offset:footer_offset+10]
    print(f"Footer at offset {footer_offset}: {footer_data.hex()}")
    
    # Parse footer
    # First 8 bytes = meta block offset (little-endian uint64)
    meta_offset = struct.unpack('<Q', footer_data[:8])[0]
    reserved = footer_data[8]
    entry_type = footer_data[9]
    
    print(f"Meta block offset: {meta_offset}")
    print(f"Reserved byte: {reserved}")
    print(f"Entry type: {entry_type}")
    
    # Meta block contains index blocks which point to data blocks
    if meta_offset > 0 and meta_offset < len(data):
        print(f"\nAnalyzing meta block at offset {meta_offset}...")
        meta_block = data[meta_offset:]
        
        # Meta block format: varint(block_data_len) + block_data + varint(block_data_len) + block_data
        # Actually it's: [block1_len(8 bytes)] [block1_data] [block2_len(8 bytes)] [block2_data] ...
        
        pos = 0
        block_count = 0
        while pos < len(meta_block) - 8:
            try:
                block_len = struct.unpack('<Q', meta_block[pos:pos+8])[0]
                pos += 8
                
                if pos + block_len > len(meta_block):
                    break
                
                block_data = meta_block[pos:pos+block_len]
                print(f"\n  Block #{block_count} at {pos}: {block_len} bytes")
                print(f"    Raw: {block_data[:100]}")
                
                # Try to decompress
                try:
                    decompressed = zlib.decompress(block_data)
                    print(f"    Decompressed ({len(decompressed)} bytes): {decompressed[:200]}")
                except Exception as e:
                    print(f"    Decompression failed: {e}")
                
                pos += block_len
                block_count += 1
                
            except Exception as e:
                print(f"  Error reading block at {pos}: {e}")
                break
    
    # Look for compressed data blocks in the file
    print(f"\n--- Searching for compressed data blocks ---")
    
    # LevelDB data blocks are varint-prefixed
    # Format: [key_length_varint][key][value_length_varint][value_type][value]
    
    # Search for typical block markers
    # Block type 1 = data block, 2 = meta block, 3 = index block
    
    block_types = {1: 'data', 2: 'meta', 3: 'index'}
    
    # Try to find blocks by looking at the structure
    # Data blocks start with varint of block length
    
    pos = 0
    block_num = 0
    
    while pos < len(data) - 20:
        # Try to read a varint (block header)
        try:
            block_size, new_pos = read_varint(data, pos)
            
            if block_size > 100000 or block_size < 10:
                pos += 1
                continue
            
            block_data = data[pos:new_pos + block_size]
            
            print(f"\nBlock {block_num} at {pos}: size {block_size}")
            
            # Try decompressing
            for header in [(0x78, 0x9c), (0x78, 0xda), (0x78, 0x01)]:
                if block_data[:2] == bytes(header):
                    try:
                        decompressed = zlib.decompress(block_data)
                        print(f"  ZLIB compressed: {len(decompressed)} bytes")
                        
                        # Try to find JSON
                        json_starts = []
                        for match in re.finditer(rb'\{', block_data):
                            json_starts.append(match.start())
                        
                        for json_pos in json_starts[:5]:
                            try:
                                # Find matching brace
                                depth = 0
                                start = json_pos
                                for i in range(json_pos, len(block_data)):
                                    if block_data[i] == ord('{'):
                                        depth += 1
                                    elif block_data[i] == ord('}'):
                                        depth -= 1
                                        if depth == 0:
                                            json_obj = block_data[start:i+1]
                                            try:
                                                decoded = json_obj.decode('utf-8', errors='replace')
                                                parsed = json.loads(decoded)
                                                print(f"  JSON at {json_pos}: {decoded[:500]}")
                                            except:
                                                pass
                                            break
                            except:
                                pass
                        
                    except Exception as e:
                        print(f"  Decompression error: {e}")
                    break
            
            pos = new_pos + block_size
            block_num += 1
            
            if block_num > 50:  # Limit output
                print(f"  ... (stopping after 50 blocks)")
                break
                
        except Exception as e:
            pos += 1
            if pos > 1000000:  # Safety limit
                break

def find_all_json_candidates(data):
    """Find all potential JSON objects by scanning for brace patterns"""
    print(f"\n--- Scanning for JSON candidates ---")
    
    candidates = []
    brace_pos = []
    
    for i, byte in enumerate(data):
        if byte == ord('{'):
            if not brace_pos:
                brace_pos.append(i)
        elif byte == ord('}'):
            if brace_pos:
                start = brace_pos.pop()
                if not brace_pos:
                    # Found a complete object
                    obj_data = data[start:i+1]
                    try:
                        decoded = obj_data.decode('utf-8', errors='replace')
                        # Check if it looks like JSON
                        if '":{' in decoded or '":"' in decoded or '":[' in decoded:
                            candidates.append((start, decoded[:1000]))
                    except:
                        pass
    
    print(f"Found {len(candidates)} JSON-like objects")
    
    # Show first 5 candidates
    for i, (pos, content) in enumerate(candidates[:5]):
        print(f"\nCandidate #{i+1} at offset {pos}:")
        # Try to parse as JSON
        try:
            # Clean up the data (replace non-UTF8 bytes)
            cleaned = re.sub(rb'[\x80-\xff]+', b'?', content.encode() if isinstance(content, str) else content)
            decoded = cleaned.decode('utf-8', errors='replace')
            parsed = json.loads(decoded)
            print(f"  Parsed successfully!")
            print(f"  Keys: {list(parsed.keys()) if isinstance(parsed, dict) else 'array'}")
            
            # Show structure
            if isinstance(parsed, dict):
                for key, value in list(parsed.items())[:5]:
                    if isinstance(value, (dict, list)):
                        print(f"    {key}: {type(value).__name__} with {len(value)} items")
                    else:
                        print(f"    {key}: {type(value).__name__} = {value}")
        except Exception as e:
            print(f"  Error parsing: {e}")
            print(f"  Preview: {content[:200]}")

def analyze_raw_structure(file_path):
    """Analyze the raw structure without compression assumptions"""
    print(f"\n{'='*80}")
    print(f"Raw Structure Analysis: {file_path}")
    print(f"{'='*80}")
    
    with open(file_path, 'rb') as f:
        data = f.read()
    
    print(f"File size: {len(data)} bytes")
    
    # Find all occurrences of specific patterns
    patterns = [
        (b'"_id"', 'ID field'),
        (b'"pages"', 'Pages field'),
        (b'"name"', 'Name field'),
        (b'"title"', 'Title field'),
        (b'"type"', 'Type field'),
        (b'"text"', 'Text field'),
        (b'"markdown"', 'Markdown field'),
        (b'"ownership"', 'Ownership field'),
        (b'"folder"', 'Folder field'),
        (b'"sort"', 'Sort field'),
        (b'"show"', 'Show field'),
        (b'"level"', 'Level field'),
        (b'"image"', 'Image field'),
        (b'"video"', 'Video field'),
    ]
    
    for pattern, description in patterns:
        positions = []
        pos = 0
        while True:
            pos = data.find(pattern, pos)
            if pos == -1:
                break
            positions.append(pos)
            pos += 1
        
        if positions:
            print(f"\n{description} ('{pattern.decode()}'): {len(positions)} occurrences")
            
            # Show context around first occurrence
            if positions:
                start = max(0, positions[0] - 100)
                end = min(len(data), positions[0] + 200)
                context = data[start:end]
                # Highlight the pattern
                highlight_pos = positions[0] - start
                highlighted = context[:highlight_pos].decode('utf-8', errors='replace')
                highlighted += context[highlight_pos:highlight_pos+10].decode('utf-8', errors='replace')
                highlighted += context[highlight_pos+10:].decode('utf-8', errors='replace')
                print(f"  Context: ...{highlighted[:150]}...")

if __name__ == '__main__':
    files = [
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000376.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/journal/001167.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/journal/000171.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/journal/000361.ldb',
    ]
    
    for file_path in files:
        if Path(file_path).exists():
            with open(file_path, 'rb') as f:
                data = f.read()
            
            parse_leveldb_sst(data)
            find_all_json_candidates(data)
            analyze_raw_structure(file_path)
        else:
            print(f"File not found: {file_path}")