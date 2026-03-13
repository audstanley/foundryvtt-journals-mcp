#!/usr/bin/env python3
"""
Working LevelDB Journal Reader for Foundry VTT
This module can read and parse journal data from LevelDB files
"""

import zlib
import re
import json
from pathlib import Path
from typing import Dict, List, Any, Optional, Tuple

# Control character cleanup function
def clean_json_string(s: str) -> str:
    """Remove problematic control characters while preserving valid JSON"""
    result = []
    in_string = False
    escape = False
    
    for char in s:
        code = ord(char)
        
        if escape:
            result.append(char)
            escape = False
            continue
        
        if char == '\\' and in_string:
            result.append(char)
            escape = True
            continue
        
        if char == '"':
            in_string = not in_string
            result.append(char)
            continue
        
        # Keep printable ASCII and common whitespace
        if 32 <= code < 127 or code in (9, 10, 13):  # tab, newline, carriage return
            result.append(char)
        elif code == 0 or code == 127:  # null and delete - replace
            result.append('?')
        elif code < 32:  # other control chars - replace
            result.append(' ')
        else:
            result.append(char)
    
    return ''.join(result)

def decompress_value(data: bytes) -> bytes:
    """Try to decompress a value using various methods"""
    # Try standard zlib
    try:
        return zlib.decompress(data)
    except:
        pass
    
    # Try raw deflate
    try:
        return zlib.decompress(data, -zlib.MAX_WBITS)
    except:
        pass
    
    # Return as-is if decompression fails
    return data

def parse_key_value(raw_key: bytes, raw_value: bytes) -> Tuple[Optional[str], Optional[Dict[str, Any]]]:
    """
    Parse a key-value pair from LevelDB
    
    Returns:
        Tuple of (cleaned_key, parsed_value_dict)
    """
    # Clean the key
    cleaned_key = raw_key.decode('utf-8', errors='replace')
    # Remove control characters from key
    cleaned_key = re.sub(r'[^\x20-\x7E]+', '', cleaned_key)
    
    # Try to decompress the value
    decompressed = decompress_value(raw_value)
    
    # Clean and parse JSON
    cleaned_json = clean_json_string(decompressed.decode('utf-8', errors='replace'))
    
    try:
        parsed = json.loads(cleaned_json)
        return cleaned_key, parsed if isinstance(parsed, dict) else None
    except json.JSONDecodeError:
        return cleaned_key, None

def read_journal_files(journal_dir: str) -> Dict[str, Dict[str, Any]]:
    """
    Read all journal data from a LevelDB journal directory
    
    Args:
        journal_dir: Path to the journal directory containing .ldb files
        
    Returns:
        Dictionary mapping keys to parsed JSON objects
    """
    journal_data = {}
    journal_path = Path(journal_dir)
    
    if not journal_path.exists():
        raise FileNotFoundError(f"Journal directory not found: {journal_dir}")
    
    # Find all .ldb files
    ldb_files = list(journal_path.glob('*.ldb'))
    
    for ldb_file in ldb_files:
        print(f"Processing {ldb_file.name}...")
        
        try:
            with open(ldb_file, 'rb') as f:
                data = f.read()
            
            # Simple LevelDB SST parsing
            # The data is a series of compressed blocks
            # We need to find and decompress the compressed sections
            
            # Try multiple decompression points
            decompression_points = [0, 1000, 5000, 10000, 50000, 100000]
            
            for offset in decompression_points:
                if offset < len(data) - 10:
                    try:
                        # Try to find compressed data
                        for header in [(0x78, 0x9c), (0x78, 0xda), (0x78, 0x01)]:
                            if data[offset:offset+2] == bytes(header):
                                try:
                                    decompressed = zlib.decompress(data[offset:])
                                    
                                    # Try to extract JSON objects
                                    json_objects = extract_json_objects(decompressed)
                                    for obj in json_objects:
                                        key, value = parse_key_value(
                                            obj.get('key', b''), 
                                            obj.get('value', b'')
                                        )
                                        if key and value:
                                            journal_data[key] = value
                                    
                                except:
                                    pass
                    except:
                        pass
            
            # Alternative: scan for patterns in raw data
            raw_journal_data = scan_for_journal_entries(data)
            for key, value in raw_journal_data.items():
                if key and value:
                    journal_data[key] = value
                    
        except Exception as e:
            print(f"Error processing {ldb_file}: {e}")
    
    return journal_data

def extract_json_objects(data: bytes) -> List[Dict]:
    """Extract key-value pairs from decompressed data"""
    results = []
    
    # Look for patterns: key followed by value
    # This is a simplified parser - actual SST format is more complex
    
    pos = 0
    while pos < len(data) - 10:
        # Try to find a string pattern that could be a key
        if data[pos:pos+8] == b'journal.':
            # Found a potential key start
            key_end = data.find(b'"', pos + 8)
            if key_end != -1:
                # Find the key value
                value_start = data.find(b'":', key_end)
                if value_start != -1:
                    value_start += 2
                    
                    # Try to find JSON object start
                    obj_start = data.find(b'{', value_start)
                    if obj_start != -1:
                        # Find matching brace
                        depth = 0
                        obj_end = obj_start
                        for i in range(obj_start, min(obj_start + 10000, len(data))):
                            if data[i] == ord('{'):
                                depth += 1
                            elif data[i] == ord('}'):
                                depth -= 1
                                if depth == 0:
                                    obj_end = i + 1
                                    break
                        
                        if depth == 0:
                            key = data[pos:key_end].decode('utf-8', errors='replace')
                            value = data[obj_start:obj_end]
                            
                            results.append({
                                'key': key.encode(),
                                'value': value
                            })
                            pos = obj_end
                            continue
        
        pos += 1
    
    return results

def scan_for_journal_entries(data: bytes) -> Dict[str, Dict[str, Any]]:
    """Scan raw data for journal entry patterns"""
    results = {}
    
    # Look for "journal." patterns
    pos = 0
    while pos < len(data) - 10:
        pos = data.find(b'journal.', pos)
        if pos == -1:
            break
        
        # Extract context
        start = max(0, pos - 20)
        end = min(len(data), pos + 500)
        context = data[start:end]
        
        # Try to extract JSON
        try:
            # Find JSON object
            json_start = context.find(b'{')
            if json_start != -1:
                depth = 0
                json_end = json_start
                for i in range(json_start, min(json_start + 5000, len(context))):
                    if context[i] == ord('{'):
                        depth += 1
                    elif context[i] == ord('}'):
                        depth -= 1
                        if depth == 0:
                            json_end = i + 1
                            break
                
                if depth == 0:
                    json_bytes = context[json_start:json_end]
                    cleaned = clean_json_string(json_bytes.decode('utf-8', errors='replace'))
                    
                    # Extract key from context
                    key_start = context.find(b'journal.')
                    key_end = context.find(b'"', key_start)
                    if key_start != -1 and key_end != -1:
                        key = context[key_start:key_end].decode('utf-8', errors='replace')
                        
                        try:
                            parsed = json.loads(cleaned)
                            if isinstance(parsed, dict):
                                results[key] = parsed
                        except:
                            pass
        except:
            pass
        
        pos = end
    
    return results

def get_journal_entries(journal_data: Dict[str, Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Extract journal entries from parsed data"""
    entries = []
    
    for key, value in journal_data.items():
        if isinstance(value, dict) and 'type' in value:
            entry_type = value.get('type', '')
            
            # Journal entries typically have 'pages' array
            if 'pages' in value or entry_type == 'journal':
                entries.append({
                    'key': key,
                    'data': value
                })
    
    return entries

def get_pages(journal_data: Dict[str, Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Extract page data from parsed data"""
    pages = []
    
    for key, value in journal_data.items():
        if isinstance(value, dict) and 'type' in value:
            entry_type = value.get('type', '')
            
            # Pages have specific content types
            if entry_type in ('text', 'image', 'video', 'pdf'):
                pages.append({
                    'key': key,
                    'data': value
                })
    
    return pages

def get_page_by_id(journal_data: Dict[str, Dict[str, Any]], page_id: str) -> Optional[Dict[str, Any]]:
    """Find a page by its ID"""
    for key, value in journal_data.items():
        if isinstance(value, dict) and value.get('_id') == page_id:
            return value
    return None

# Main execution for testing
if __name__ == '__main__':
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python reader.py <journal_directory>")
        sys.exit(1)
    
    journal_dir = sys.argv[1]
    
    try:
        journal_data = read_journal_files(journal_dir)
        
        print(f"\nFound {len(journal_data)} journal entries/pages")
        
        # Show journal entries
        entries = get_journal_entries(journal_data)
        print(f"\nJournal Entries: {len(entries)}")
        for entry in entries[:5]:
            print(f"  - {entry['data'].get('name', 'Unnamed')}")
        
        # Show pages
        pages = get_pages(journal_data)
        print(f"\nPages: {len(pages)}")
        for page in pages[:5]:
            print(f"  - {page['data'].get('name', 'Unnamed')} ({page['data'].get('type', 'unknown')})")
        
        print(f"\nTotal keys in data: {len(journal_data)}")
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()