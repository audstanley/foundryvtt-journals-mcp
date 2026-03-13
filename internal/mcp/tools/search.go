package tools

import (
	"encoding/json"
	"fmt"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
)

// SearchTools contains all search-related MCP tools
type SearchTools struct {
	WorldsPath string
}

// NewSearchTools creates a new search tools instance
func NewSearchTools(worldsPath string) *SearchTools {
	return &SearchTools{
		WorldsPath: worldsPath,
	}
}

// SearchEntries searches for journal entries by name (case-sensitive, partial match)
func (t *SearchTools) SearchEntries(params json.RawMessage) (interface{}, error) {
	var worldName, query string
	var username string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(string); ok && wn != "" {
		worldName = wn
	}
	if q, ok := p["query"].(string); ok && q != "" {
		query = q
	}
	if un, ok := p["user"].(string); ok && un != "" {
		username = un
	}

	if worldName == "" {
		return nil, fmt.Errorf("world is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	entries, err := repo.SearchEntries(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search entries: %w", err)
	}

	// Filter by permissions if user specified
	if username != "" {
		userID, err := repo.GetUserID(username)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		var filtered []journal.JournalEntry
		for _, entry := range entries {
			if journal.HasAccess(entry.Ownership, userID) {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}

	var result []map[string]interface{}
	for _, entry := range entries {
		permission := journal.PermissionLevel(0)
		if username != "" {
			userID, _ := repo.GetUserID(username)
			permission = journal.CheckPermission(entry.Ownership, userID)
		}

		result = append(result, map[string]interface{}{
			"id":         entry.ID,
			"name":       entry.Name,
			"pageCount":  len(entry.Pages),
			"permission": int(permission),
			"hasAccess":  journal.HasAccess(entry.Ownership, username),
		})
	}

	return map[string]interface{}{
		"world":   worldName,
		"query":   query,
		"results": result,
		"count":   len(result),
	}, nil
}

// SearchPages searches for pages within content (case-sensitive, partial match)
func (t *SearchTools) SearchPages(params json.RawMessage) (interface{}, error) {
	var worldName, query string
	var entryID *string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(string); ok && wn != "" {
		worldName = wn
	}
	if q, ok := p["query"].(string); ok && q != "" {
		query = q
	}
	if eid, ok := p["entry_id"].(string); ok && eid != "" {
		entryID = &eid
	}

	if worldName == "" {
		return nil, fmt.Errorf("world is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	pages, err := repo.SearchPages(query, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to search pages: %w", err)
	}

	var result []map[string]interface{}
	for _, page := range pages {
		pageData := map[string]interface{}{
			"id":   page.ID,
			"name": page.Name,
			"type": page.Type,
		}

		// Include content snippet
		if page.Text != nil && page.Text.Content != "" {
			// Extract snippet (first 200 chars)
			content := page.Text.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			pageData["content"] = content
		}

		result = append(result, pageData)
	}

	return map[string]interface{}{
		"world":    worldName,
		"query":    query,
		"entry_id": entryID,
		"results":  result,
		"count":    len(result),
	}, nil
}
