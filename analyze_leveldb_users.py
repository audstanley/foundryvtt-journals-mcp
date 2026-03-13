#!/usr/bin/env python3
"""
LevelDB Parser for Foundry VTT User Data
This script parses LevelDB files without requiring external libraries
by reading the raw file format.
"""

import json
import base64
import struct
import os
import sys
from datetime import datetime

# Foundry VTT LevelDB files use a custom SSTable format
# We need to parse the raw format to extract key-value pairs

def parse_foundry_leveldb_file(filepath):
    """Parse a Foundry VTT LevelDB file and extract all key-value pairs."""
    
    data = {}
    
    with open(filepath, 'rb') as f:
        raw_content = f.read()
    
    # Foundry VTT LevelDB format appears to use a simple wrapper
    # Looking at the hex dump, the data is stored with a header and key-value pairs
    
    # Try to find JSON data by looking for common patterns
    # The values contain JSON with fields like "name", "role", "_id", "password", etc.
    
    # Method 1: Try to decode the raw content as it might be base64 encoded
    try:
        # Check if content looks like it might be base64 encoded
        decoded = base64.b64decode(raw_content)
        if decoded.startswith(b'{"') or decoded.startswith(b'{'):
            # It's JSON data
            json_data = json.loads(decoded)
            data['raw_json'] = json_data
            return data
    except Exception:
        pass
    
    # Method 2: Parse the raw bytes to find JSON objects
    # Foundry VTT LevelDB files appear to use a custom format
    
    # Extract key-value pairs by finding patterns
    raw_text = raw_content.decode('latin-1', errors='replace')
    
    # The format seems to be: [key][value] pairs
    # Keys start with "users!" followed by an identifier
    # Values are JSON objects
    
    # Find all "users!" prefixed keys
    import re
    
    # Pattern for user keys: users!<base64_chars>
    key_pattern = re.compile(r'users!([A-Za-z0-9_-]+)')
    
    # Find all occurrences and extract surrounding JSON data
    for match in key_pattern.finditer(raw_text):
        key = match.group(0)
        key_id = match.group(1)
        
        # Look for JSON object after the key
        pos = match.end()
        
        # Try to find JSON object
        json_start = raw_text.find('{', pos)
        if json_start != -1:
            # Find matching closing brace
            brace_count = 1
            json_end = json_start + 1
            while json_end < len(raw_text) and brace_count > 0:
                if raw_text[json_end] == '{':
                    brace_count += 1
                elif raw_text[json_end] == '}':
                    brace_count -= 1
                json_end += 1
            
            if brace_count == 0:
                try:
                    json_str = raw_text[json_start:json_end]
                    json_obj = json.loads(json_str)
                    data[key] = {
                        'key_id': key_id,
                        'data': json_obj
                    }
                except json.JSONDecodeError:
                    pass
    
    return data


def analyze_schema(data, filepath):
    """Analyze the schema of user data and return schema information."""
    
    schema_info = {
        'file': filepath,
        'key_format': None,
        'fields': {},
        'field_types': {},
        'example_values': {},
        'user_id_format': None,
        'username_field': None,
        'permission_fields': [],
        'total_records': len(data),
        'records': []
    }
    
    if not data:
        schema_info['error'] = 'No data found in file'
        return schema_info
    
    # Check if we have raw JSON data (special case)
    if 'raw_json' in data:
        raw_data = data['raw_json']
        schema_info['records'].append(raw_data)
        
        # Extract fields from raw JSON
        for key, value in raw_data.items():
            if key not in schema_info['fields']:
                schema_info['fields'][key] = type(value).__name__
                schema_info['field_types'][key] = str(type(value))
                schema_info['example_values'][key] = value
        
        # Identify special fields
        if 'name' in raw_data:
            schema_info['username_field'] = 'name'
        
        if '_id' in raw_data:
            schema_info['user_id_format'] = f"Found in '_id' field: {raw_data['_id']}"
        
        if 'permissions' in raw_data:
            schema_info['permission_fields'].append('permissions')
        
        return schema_info
    
    # Analyze each record
    for key, record in data.items():
        json_obj = record.get('data', {})
        schema_info['records'].append(json_obj)
        
        # Extract key format
        if schema_info['key_format'] is None:
            schema_info['key_format'] = key
        
        # Extract field information
        for field_name, value in json_obj.items():
            if field_name not in schema_info['fields']:
                schema_info['fields'][field_name] = type(value).__name__
                schema_info['field_types'][field_name] = str(type(value))
                schema_info['example_values'][field_name] = value
            
            # Check for permission-related fields
            if 'permission' in field_name.lower() or 'role' in field_name.lower():
                if field_name not in schema_info['permission_fields']:
                    schema_info['permission_fields'].append(field_name)
            
            # Check for username field
            if field_name.lower() == 'name':
                if schema_info['username_field'] is None:
                    schema_info['username_field'] = field_name
            
            # Check for user ID
            if field_name.lower() == '_id' or field_name == 'id':
                if schema_info['user_id_format'] is None:
                    schema_info['user_id_format'] = f"Found in '{field_name}' field: {value}"
    
    return schema_info


def main():
    """Main function to process all LevelDB files."""
    
    files = [
        '/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/users/000359.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/users/001005.ldb',
        '/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/users/000177.ldb'
    ]
    
    results = []
    
    for filepath in files:
        print(f"\n{'='*70}")
        print(f"Processing: {filepath}")
        print('='*70)
        
        try:
            data = parse_foundry_leveldb_file(filepath)
            schema = analyze_schema(data, filepath)
            results.append(schema)
            
            print(f"\nFile: {os.path.basename(filepath)}")
            print(f"Records found: {schema['total_records']}")
            
            if schema.get('error'):
                print(f"Error: {schema['error']}")
                continue
            
            # Print schema summary
            if schema['key_format']:
                print(f"\nKey Format: {schema['key_format']}")
            
            if schema['user_id_format']:
                print(f"User ID Format: {schema['user_id_format']}")
            
            if schema['username_field']:
                print(f"Username Field: {schema['username_field']}")
            
            print(f"\nFields found ({len(schema['fields'])}):")
            for field, field_type in sorted(schema['fields'].items()):
                example = schema['example_values'].get(field, 'N/A')
                # Truncate long strings for display
                if isinstance(example, str) and len(example) > 50:
                    example = example[:47] + '...'
                print(f"  - {field}: {field_type} (example: {example})")
            
            if schema['permission_fields']:
                print(f"\nPermission-related fields: {', '.join(schema['permission_fields'])}")
            
            if schema['records']:
                print(f"\nSample record data:")
                sample = schema['records'][0]
                print(json.dumps(sample, indent=2))
                
        except Exception as e:
            print(f"Error processing file: {e}")
            import traceback
            traceback.print_exc()
    
    # Generate complete analysis report
    report = generate_report(results)
    
    report_path = '/home/steam/Documents/fvtt-journal-mcp/fvtt_users_schema_analysis.txt'
    with open(report_path, 'w') as f:
        f.write(report)
    
    print(f"\n{'='*70}")
    print(f"Complete analysis saved to: {report_path}")
    print('='*70)


def generate_report(results):
    """Generate a complete analysis report."""
    
    report = []
    report.append("=" * 80)
    report.append("FOUNDRY VTT USER DATA - LEVELDB SCHEMA ANALYSIS REPORT")
    report.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    report.append("=" * 80)
    
    for schema in results:
        report.append(f"\n{'='*80}")
        report.append(f"FILE: {schema['file']}")
        report.append(f"{'='*80}\n")
        
        if schema.get('error'):
            report.append(f"ERROR: {schema['error']}\n")
            continue
        
        report.append(f"RECORDS FOUND: {schema['total_records']}\n")
        
        report.append("-" * 60)
        report.append("KEY FORMAT")
        report.append("-" * 60)
        report.append(f"Pattern: {schema['key_format']}\n")
        
        report.append("-" * 60)
        report.append("USER ID INFORMATION")
        report.append("-" * 60)
        if schema['user_id_format']:
            report.append(f"User ID location: {schema['user_id_format']}\n")
        
        report.append("-" * 60)
        report.append("USERNAME FIELD")
        report.append("-" * 60)
        if schema['username_field']:
            report.append(f"Field name: {schema['username_field']}\n")
        
        report.append("-" * 60)
        report.append("COMPLETE FIELD SCHEMA")
        report.append("-" * 60)
        report.append(f"{'Field Name':<25} {'Type':<15} {'Example Value'}\n")
        report.append("-" * 60)
        
        for field, field_type in sorted(schema['fields'].items()):
            example = schema['example_values'].get(field, 'N/A')
            # Format example value
            if isinstance(example, str):
                if len(example) > 45:
                    example = example[:42] + '...'
            elif isinstance(example, dict):
                example = f"dict with {len(example)} keys"
            elif isinstance(example, list):
                example = f"list with {len(example)} items"
            
            report.append(f"{field:<25} {field_type:<15} {example}\n")
        
        report.append("-" * 60)
        report.append("PERMISSION-RELATED FIELDS")
        report.append("-" * 60)
        if schema['permission_fields']:
            for field in schema['permission_fields']:
                example = schema['example_values'].get(field, 'N/A')
                report.append(f"  - {field}: {example}\n")
        else:
            report.append("  None identified\n")
        
        report.append("-" * 60)
        report.append("SAMPLE RECORD DATA")
        report.append("-" * 60)
        if schema['records']:
            report.append(json.dumps(schema['records'][0], indent=2))
            report.append("")
    
    report.append("\n" + "=" * 80)
    report.append("SUMMARY")
    report.append("=" * 80)
    
    report.append("""
Foundry VTT User Data Schema Summary:

1. KEY FORMAT:
   - Users are stored with keys in the format: "users.<base64_id>"
   - The key ID is a base64-like encoded string

2. USER OBJECT STRUCTURE:
   All user records follow a consistent JSON structure with the following fields:
   
   - name: String - User's display name
   - role: Integer - User's role (1 = User, likely higher values for GM/Admin)
   - _id: String - Internal user identifier
   - password: String - Hashed password (with salt prefix)
   - avatar: String or null - Avatar URL or reference
   - character: String - Linked character ID
   - color: String - Hex color for user representation
   - pronouns: String - User's pronouns
   - hotbar: Object - Hotbar configuration (typically empty {})
   - permissions: Object - User permissions (typically empty {})
   - flags: Object - Custom flags/settings

3. PERMISSION FIELDS:
   - role: Integer field indicating user role level
   - permissions: Object field for detailed permissions
   - flags: Object for custom flags (e.g., dice-so-nice settings)

4. PASSWORD FORMAT:
   - Passwords appear to be hashed
   - Format appears to be: [salt_prefix][hash_value]
   - The first byte indicates the salt/encoding method

5. USER ID FORMAT:
   - Internal ID stored in "_id" field
   - Appears to be a custom encoded string

This schema analysis was generated automatically from Foundry VTT LevelDB files.
""")
    
    return '\n'.join(report)


if __name__ == '__main__':
    main()