#!/usr/bin/env python3
"""
Manual LevelDB parser for Foundry VTT journal data
LevelDB files use a specific binary format with SST tables
"""

import struct
import zlib
import json
from pathlib import Path

def read_varint(data, offset):
    """Read a varint from binary data"""
    result = 0
    shift = 0
    while True:
        if offset >= len(data):
            break
        byte = data[offset]
        offset += 1
        result |= (byte & 0x7f) << shift
        if not (byte & 0x80):
            break
        shift += 7
    return result, offset

def read_string(data, offset, length):
    """Read a string of given length"""
    return data[offset:offset+length].decode('utf-8', errors='replace')

def parse_sst_table(file_path):
    """Parse a LevelDB SST table file"""
    print(f"\n{'='*80}")
    print(f"File: {file_path}")
    print(f"{'='*80}")
    
    with open(file_path, 'rb') as f:
        data = f.read()
    
    print(f"File size: {len(data)} bytes")
    
    # Look for trailer (last 10 bytes of SST file)
    # Trailer format: type(1 byte) + offset(8 bytes) + reserved(1 byte)
    if len(data) >= 10:
        trailer_offset = len(data) - 10
        trailer_type = data[trailer_offset]
        trailer_offset_val = struct.unpack('<Q', data[trailer_offset+1:trailer_offset+9])[0]
        print(f"Trailer: type={trailer_type}, offset={trailer_offset_val}")
    
    # Try to find data blocks and parse them
    # LevelDB SST format has data blocks followed by meta blocks
    
    # Search for common patterns in the data
    print("\n--- Analyzing binary structure ---")
    
    # Look for "journal:" prefix which indicates journal entries
    journal_keys = []
    for i in range(len(data) - 10):
        if data[i:i+8] == b'journal:':
            journal_keys.append(i)
    
    print(f"Found {len(journal_keys)} potential 'journal:' key locations")
    
    for idx, pos in enumerate(journal_keys[:5]):  # Show first 5
        print(f"\n--- Journal key candidate #{idx+1} at offset {pos} ---")
        # Try to read the key
        # Key format: varint length + key data
        length, new_pos = read_varint(data, pos)
        print(f"  Varint length: {length}")
        if new_pos + length <= len(data):
            key = data[new_pos:new_pos+length].decode('utf-8', errors='replace')
            print(f"  Key: {key}")
            print(f"  Raw bytes around key: {data[pos:min(pos+30, len(data))]}")
    
    # Try to find compressed/decompressed data
    print("\n--- Looking for JSON-like structures ---")
    
    # Search for common JSON patterns
    json_patterns = []
    for pattern in [b'"_id"', b'"pages"', b'"text"', b'"type"']:
        pos = 0
        while True:
            pos = data.find(pattern, pos)
            if pos == -1:
                break
            json_patterns.append((pos, pattern))
            pos += 1
    
    print(f"Found {len(json_patterns)} JSON-like patterns")
    
    # Show some examples
    for pos, pattern in json_patterns[:10]:
        print(f"  Offset {pos}: {pattern}")
        # Show context
        start = max(0, pos - 20)
        end = min(len(data), pos + 100)
        print(f"  Context: {data[start:end][:100]}")
    
    return data

def analyze_compressed_data(file_path):
    """Try to decompress data and look for JSON"""
    print(f"\n--- Trying to decompress data from {file_path} ---")
    
    with open(file_path, 'rb') as f:
        data = f.read()
    
    # Try zlib decompression on different offsets
    for offset in [100, 500, 1000, 10000, 100000]:
        if offset < len(data) - 10:
            try:
                # Try to find zlib header (0x78 0x9c or 0x78 0xda)
                for header in [0x789c, 0x78da, 0x7801]:
                    if data[offset:offset+2] == struct.pack('<H', header):
                        try:
                            decompressed = zlib.decompress(data[offset:])
                            print(f"\n  Found compressed data at offset {offset}")
                            print(f"  Compressed size: {len(data) - offset}")
                            print(f"  Decompressed size: {len(decompressed)}")
                            # Try to parse as JSON
                            try:
                                # Find JSON object start
                                json_start = decompressed.find(b'{')
                                if json_start != -1:
                                    json_end = decompressed.rfind(b'}') + 1
                                    json_str = decompressed[json_start:json_end]
                                    print(f"  JSON found: {json_str[:500]}")
                            except Exception as e:
                                print(f"  Decompressed (not JSON): {decompressed[:200]}")
                        except Exception as e:
                            pass
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
            parse_sst_table(file_path)
            analyze_compressed_data(file_path)
        else:
            print(f"File not found: {file_path}")