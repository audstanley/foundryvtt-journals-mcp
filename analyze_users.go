package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

// UserSchema represents the schema analysis for user data
type UserSchema struct {
	Fields     map[string]FieldInfo
	TotalCount int
	Examples   map[string]interface{}
}

// FieldInfo contains type and example information for a field
type FieldInfo struct {
	Type    string
	Example interface{}
	Count   int
	IsArray bool
}

func main() {
	// Define the paths to the user databases (the directories containing .ldb files)
	dbPaths := []struct {
		name   string
		dbPath string
	}{
		{
			name:   "testworld",
			dbPath: "/home/steam/Documents/fvtt-journal-mcp/worlds/testworld/data/users/",
		},
		{
			name:   "dragons_guard",
			dbPath: "/home/steam/Documents/fvtt-journal-mcp/worlds/dragons_guard/data/users/",
		},
		{
			name:   "the-iron-kingdoms",
			dbPath: "/home/steam/Documents/fvtt-journal-mcp/worlds/the-iron-kingdoms/data/users/",
		},
	}

	// Collect all schema data across all databases
	allSchema := make(map[string][]map[string]interface{})
	allKeys := make(map[string][]string)

	// Open and analyze each database
	for _, dbConfig := range dbPaths {
		fmt.Printf("\nAnalyzing: %s\n", dbConfig.name)

		// Check if directory exists
		if _, err := os.Stat(dbConfig.dbPath); os.IsNotExist(err) {
			log.Printf("Database directory does not exist: %s\n", dbConfig.dbPath)
			continue
		}

		// Open the LevelDB database
		db, err := openLevelDB(dbConfig.dbPath)
		if err != nil {
			log.Printf("Error opening database: %v\n", err)
			continue
		}

		// Collect all keys and values
		var keys []string
		var values []map[string]interface{}
		iter := db.NewIterator(nil, nil)
		for iter.Next() {
			key := string(iter.Key())
			value := iter.Value()
			keys = append(keys, key)
			vals, err := parseJSON(value)
			if err != nil {
				// Try to salvage partial data
				vals = map[string]interface{}{
					"_parse_error": true,
					"raw_length":   len(value),
				}
			}
			values = append(values, vals)
		}
		iter.Release()

		allSchema[dbConfig.name] = values
		allKeys[dbConfig.name] = keys

		db.Close()
	}

	// Generate comprehensive schema analysis
	analysis := generateSchemaAnalysis(allSchema, allKeys)

	// Save to markdown file
	err := saveAnalysisToMarkdown(analysis, "foundry_vtt_user_schema_analysis.md")
	if err != nil {
		log.Fatalf("Error saving analysis: %v", err)
	}

	// Also print to console
	printAnalysis(analysis)

	fmt.Println("\n✓ Schema analysis complete!")
	fmt.Println("  Saved to: foundry_vtt_user_schema_analysis.md")
}

func openLevelDB(path string) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return db, nil
}

func parseJSON(data []byte) (map[string]interface{}, error) {
	// Clean up binary/control characters
	cleaned := cleanBinaryData(data)

	var result map[string]interface{}
	err := json.Unmarshal(cleaned, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func cleanBinaryData(data []byte) []byte {
	// Remove or replace problematic control characters
	cleaned := make([]byte, 0, len(data))
	for _, b := range data {
		// Keep printable ASCII and common JSON characters
		if b >= 32 || b == '\t' || b == '\n' || b == '\r' {
			cleaned = append(cleaned, b)
		} else {
			// Replace other control chars with space or hex representation
			cleaned = append(cleaned, '?')
		}
	}
	return cleaned
}

func generateSchemaAnalysis(allSchema map[string][]map[string]interface{}, allKeys map[string][]string) string {
	var sb strings.Builder

	sb.WriteString("# Foundry VTT User Schema Analysis\n\n")
	sb.WriteString("## Overview\n\n")

	totalUsers := 0
	for _, users := range allSchema {
		totalUsers += len(users)
	}
	sb.WriteString(fmt.Sprintf("\n**Total databases analyzed: 3**\n"))
	sb.WriteString(fmt.Sprintf("**Total users across all databases: %d**\n\n", totalUsers))

	sb.WriteString("## Key Format Analysis\n\n")
	sb.WriteString("The key format follows: `!users!{random_id}`\n\n")
	sb.WriteString("### Key Examples\n\n")
	for name, keys := range allKeys {
		if len(keys) > 0 {
			sb.WriteString(fmt.Sprintf("**%s**: %s\n", name, keys[0]))
		}
	}

	sb.WriteString("\n### User ID Format\n\n")
	sb.WriteString("User IDs are stored as random base64-like strings prefixed with `!users!`:\n\n")
	sb.WriteString("| Database | Example User ID | Format |\n")
	sb.WriteString("|----------|-----------------|--------|\n")
	for name, keys := range allKeys {
		if len(keys) > 0 {
			userID := extractUserID(keys[0])
			sb.WriteString(fmt.Sprintf("| %s | %s | base64-like string |\n", name, userID))
		}
	}

	sb.WriteString("\n## Field Schema Analysis\n\n")

	// Collect all unique fields across databases
	allFields := make(map[string][]FieldInfo)
	for _, users := range allSchema {
		for _, user := range users {
			for field, value := range user {
				if _, exists := allFields[field]; !exists {
					allFields[field] = []FieldInfo{}
				}
				fieldInfo := FieldInfo{
					Type:    detectType(value),
					Example: value,
					Count:   1,
				}
				allFields[field] = append(allFields[field], fieldInfo)
			}
		}
	}

	// Sort fields for consistent output
	var fieldNames []string
	for name := range allFields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	sb.WriteString("### Complete Field List\n\n")
	sb.WriteString("| Field | Type | Description | Example |\n")
	sb.WriteString("|-------|------|-------------|--------|\n")

	// Define field descriptions based on Foundry VTT user schema
	fieldDescriptions := map[string]string{
		"_id":           "Unique document identifier (internal ID)",
		"password":      "Hashed password for authentication",
		"passwordSalt":  "Salt used for password hashing",
		"role":          "Numeric role ID (1=owner, 2=gm, 3=player, etc.)",
		"permissions":   "Object mapping role IDs to permission settings",
		"flags":         "Additional metadata and settings",
		"avatar":        "User avatar URL or data (null if not set)",
		"color":         "User display color in hex format",
		"character":     "Link to associated character actor ID",
		"name":          "User's display name",
		"pronouns":      "User's pronouns (optional)",
		"hotbar":        "Hotbar configuration object",
		"_stats":        "Statistics object with various metrics",
		"_creationDate": "ISO 8601 creation timestamp",
		"_modifiedDate": "ISO 8601 last modified timestamp",
	}

	for _, fieldName := range fieldNames {
		fieldInfo := allFields[fieldName][0]
		desc := fieldDescriptions[fieldName]
		example := formatExample(fieldInfo.Example)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", fieldName, fieldInfo.Type, desc, example))
	}

	sb.WriteString("\n## Detailed Field Analysis\n\n")

	// Detailed analysis for each field category
	categories := map[string][]string{
		"Authentication": {"password", "passwordSalt"},
		"Identity":       {"_id", "name", "character", "pronouns", "avatar", "color"},
		"Access Control": {"role", "permissions"},
		"Metadata":       {"flags", "hotbar", "_stats"},
		"Timestamps":     {"_creationDate", "_modifiedDate"},
	}

	for category, fields := range categories {
		sb.WriteString(fmt.Sprintf("### %s\n\n", category))
		sb.WriteString("| Field | Type | Example |\n")
		sb.WriteString("|-------|------|--------|\n")

		for _, fieldName := range fields {
			if fieldInfo, exists := allFields[fieldName]; exists {
				example := formatExample(fieldInfo[0].Example)
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", fieldName, fieldInfo[0].Type, example))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Permission Fields Structure\n\n")
	sb.WriteString("Permission fields contain a nested object structure mapping role IDs to permission arrays:\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("\"permissions\": {\n")
	sb.WriteString("  \"role_id\": [\n")
	sb.WriteString("    \"permission_name\",\n")
	sb.WriteString("    \"another_permission\"\n")
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")
	sb.WriteString("### Common Permission Types\n\n")
	sb.WriteString("| Permission | Description |\n")
	sb.WriteString("|------------|-------------|\n")
	sb.WriteString("| `CREATE` | Create new documents |\n")
	sb.WriteString("| `UPDATE` | Modify existing documents |\n")
	sb.WriteString("| `DELETE` | Remove documents |\n")
	sb.WriteString("| `ACTOR` | Access to Actor documents |\n")
	sb.WriteString("| `ITEM` | Access to Item documents |\n")
	sb.WriteString("| `COMBAT` | Access to Combat documents |\n")
	sb.WriteString("| `SETTING` | Access to Game Settings |\n")
	sb.WriteString("| `RULES` | Access to Rule documents |\n")

	sb.WriteString("\n## User Identity Analysis\n\n")
	sb.WriteString("User identity fields include name, associated character, and display color:\n\n")

	for name, users := range allSchema {
		sb.WriteString(fmt.Sprintf("### %s\n\n", name))

		if len(users) > 0 {
			displayName := ""
			characterID := ""
			userColor := ""
			if n, ok := users[0]["name"].(string); ok {
				displayName = n
			}
			if c, ok := users[0]["character"].(string); ok {
				characterID = c
			}
			if col, ok := users[0]["color"].(string); ok {
				userColor = col
			}
			sb.WriteString(fmt.Sprintf("- **Display Name**: `%s`\n", displayName))
			sb.WriteString(fmt.Sprintf("- **Character ID**: `%s`\n", characterID))
			sb.WriteString(fmt.Sprintf("- **User Color**: `%s`\n", userColor))
		}
	}

	sb.WriteString("\n## Schema Summary\n\n")
	sb.WriteString("### Key Statistics\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Databases | 3 |\n"))
	sb.WriteString(fmt.Sprintf("| Total Fields Identified | %d |\n", len(allFields)))
	sb.WriteString(fmt.Sprintf("| Authentication Fields | password, passwordSalt |\n"))
	sb.WriteString(fmt.Sprintf("| Identity Fields | _id, name, character, color, avatar |\n"))
	sb.WriteString(fmt.Sprintf("| Access Control Fields | role, permissions |\n"))
	sb.WriteString(fmt.Sprintf("| Metadata Fields | flags, hotbar, _stats |\n"))

	sb.WriteString("\n### Field Type Distribution\n\n")
	typeDistribution := make(map[string]int)
	for _, fieldInfo := range allFields {
		typeDistribution[fieldInfo[0].Type]++
	}

	var types []string
	for t := range typeDistribution {
		types = append(types, t)
	}
	sort.Strings(types)

	sb.WriteString("| Type | Count | Examples |\n")
	sb.WriteString("|------|-------|----------|\n")
	for _, t := range types {
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", t, typeDistribution[t], getSampleTypes(t)))
	}

	sb.WriteString("\n---\n")
	sb.WriteString("*Generated by Foundry VTT User Schema Analyzer*\n")

	return sb.String()
}

func extractUserID(key string) string {
	// Extract user ID from "users.000359" format
	parts := strings.Split(key, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return key
}

func detectType(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Check if it's a timestamp
		if isTimestamp(v) {
			return "ISO 8601 Timestamp"
		}
		return "string"
	case bool:
		return "boolean"
	case float64:
		// JSON numbers are float64
		if v == float64(int64(v)) {
			return "integer"
		}
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func isTimestamp(s string) bool {
	// Check for ISO 8601 format timestamps
	timestampRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	return timestampRegex.MatchString(s)
}

func formatExample(value interface{}) string {
	if value == nil {
		return "null"
	}
	switch v := value.(type) {
	case string:
		if len(v) > 30 {
			return fmt.Sprintf("`%s...`", v[:30])
		}
		return fmt.Sprintf("`%s`", v)
	case bool:
		return fmt.Sprintf("`%v`", v)
	case float64:
		return fmt.Sprintf("`%v`", int64(v))
	case []interface{}:
		if len(v) > 3 {
			return fmt.Sprintf("`[%v... (%d items)]`", v[0], len(v))
		}
		return fmt.Sprintf("`%v`", v)
	case map[string]interface{}:
		return fmt.Sprintf("`{... (%d keys)}`", len(v))
	default:
		return fmt.Sprintf("`%v`", v)
	}
}

func getSampleTypes(typeName string) string {
	switch typeName {
	case "string":
		return "username, notes"
	case "array":
		return "roles, permissions"
	case "object":
		return "flags, permissions"
	case "boolean":
		return ""
	case "integer":
		return ""
	case "number":
		return ""
	case "null":
		return ""
	default:
		return ""
	}
}

func printAnalysis(analysis string) {
	fmt.Printf("\n%.*s\n", min(len(analysis), 50000), analysis)
}

func saveAnalysisToMarkdown(analysis, filename string) error {
	return os.WriteFile(filename, []byte(analysis), 0644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
