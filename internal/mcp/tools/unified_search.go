package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
)

// UnifiedSearchTools contains unified search tools for both LevelDB and NDJSON
type UnifiedSearchTools struct {
	WorldsPath string
}

// NewUnifiedSearchTools creates a new unified search tools instance
func NewUnifiedSearchTools(worldsPath string) *UnifiedSearchTools {
	return &UnifiedSearchTools{
		WorldsPath: worldsPath,
	}
}

// SearchAll searches across both LevelDB (journals) and NDJSON (back compendium)
func (t *UnifiedSearchTools) SearchAll(params json.RawMessage) (interface{}, error) {
	var query string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if q, ok := p["query"].(string); ok && q != "" {
		query = q
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Get world name from available worlds
	worlds, err := journal.ListWorlds(t.WorldsPath)
	if err != nil || len(worlds) == 0 {
		return nil, fmt.Errorf("no worlds available")
	}
	// Use first world for search
	worldName := worlds[0]

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	defer repo.Close()

	results, err := repo.SearchAll(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	var formattedResults []map[string]interface{}
	for _, result := range results.Results {
		resultMap := map[string]interface{}{
			"id":     result.ID,
			"name":   result.Name,
			"type":   result.Type,
			"source": result.Source,
			"uuid":   result.UUID,
		}

		if result.Content != "" {
			resultMap["content"] = result.Content
		}

		if result.Meta != nil {
			resultMap["meta"] = result.Meta
		}

		formattedResults = append(formattedResults, resultMap)
	}

	levelDBCount := 0
	ndjsonCount := 0
	for _, r := range results.Results {
		if r.Source == "LevelDB" {
			levelDBCount++
		} else if r.Source == "NDJSON" {
			ndjsonCount++
		}
	}

	return map[string]interface{}{
		"query":   results.Query,
		"results": formattedResults,
		"count":   results.Count,
		"unique":  results.UniqueIDs,
		"sources": map[string]int{
			"LevelDB": levelDBCount,
			"NDJSON":  ndjsonCount,
		},
	}, nil
}

// SearchCompendium searches only the NDJSON back compendium
func (t *UnifiedSearchTools) SearchCompendium(params json.RawMessage) (interface{}, error) {
	var query string
	var entityType string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if q, ok := p["query"].(string); ok && q != "" {
		query = q
	}
	if et, ok := p["entity_type"].(string); ok && et != "" {
		entityType = et
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Get world name from available worlds
	worlds, err := journal.ListWorlds(t.WorldsPath)
	if err != nil || len(worlds) == 0 {
		return nil, fmt.Errorf("no worlds available")
	}
	// Use first world for search
	worldName := worlds[0]

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	defer repo.Close()

	ndjsonResults, err := repo.SearchNDJSON(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search ndjson: %w", err)
	}

	var formattedResults []map[string]interface{}
	for _, entity := range ndjsonResults {
		if entityType != "" {
			entityTypeMatch, _ := entity["type"].(string)
			if entityTypeMatch != entityType {
				continue
			}
		}

		entityID, _ := entity["_id"].(string)
		entityName, _ := entity["name"].(string)
		entityTypeVal, _ := entity["type"].(string)
		uuid := fmt.Sprintf("%s/%s", entityTypeVal, entityID)

		resultMap := map[string]interface{}{
			"id":     entityID,
			"name":   entityName,
			"type":   entityTypeVal,
			"source": "NDJSON",
			"uuid":   uuid,
		}

		resultMap["meta"] = entity

		formattedResults = append(formattedResults, resultMap)
	}

	return map[string]interface{}{
		"query":       query,
		"entity_type": entityType,
		"results":     formattedResults,
		"count":       len(formattedResults),
	}, nil
}

// SearchJournals searches only LevelDB journal entries
func (t *UnifiedSearchTools) SearchJournals(params json.RawMessage) (interface{}, error) {
	var query string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if q, ok := p["query"].(string); ok && q != "" {
		query = q
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Get world name from available worlds
	worlds, err := journal.ListWorlds(t.WorldsPath)
	if err != nil || len(worlds) == 0 {
		return nil, fmt.Errorf("no worlds available")
	}
	// Use first world for search
	worldName := worlds[0]

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	defer repo.Close()

	entries, err := repo.SearchEntries(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search journals: %w", err)
	}

	var formattedResults []map[string]interface{}
	for _, entry := range entries {
		resultMap := map[string]interface{}{
			"id":     entry.ID,
			"name":   entry.Name,
			"type":   "JournalEntry",
			"source": "LevelDB",
			"uuid":   fmt.Sprintf("JournalEntry/%s", entry.ID),
			"meta": map[string]interface{}{
				"pageCount": len(entry.Pages),
			},
		}
		formattedResults = append(formattedResults, resultMap)
	}

	return map[string]interface{}{
		"query":   query,
		"results": formattedResults,
		"count":   len(formattedResults),
	}, nil
}

// SearchJournalPages searches only LevelDB journal page content
func (t *UnifiedSearchTools) SearchJournalPages(params json.RawMessage) (interface{}, error) {
	var query string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if q, ok := p["query"].(string); ok && q != "" {
		query = q
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Get world name from available worlds
	worlds, err := journal.ListWorlds(t.WorldsPath)
	if err != nil || len(worlds) == 0 {
		return nil, fmt.Errorf("no worlds available")
	}
	// Use first world for search
	worldName := worlds[0]

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	defer repo.Close()

	pages, err := repo.SearchPages(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search journal pages: %w", err)
	}

	var formattedResults []map[string]interface{}
	for _, page := range pages {
		resultMap := map[string]interface{}{
			"id":     page.ID,
			"name":   page.Name,
			"type":   page.Type,
			"source": "LevelDB",
			"uuid":   fmt.Sprintf("JournalPage/%s", page.ID),
		}

		if page.Text != nil && page.Text.Content != "" {
			content := page.Text.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			resultMap["content"] = cleanContent(content)
		}

		formattedResults = append(formattedResults, resultMap)
	}

	return map[string]interface{}{
		"query":   query,
		"results": formattedResults,
		"count":   len(formattedResults),
	}, nil
}

func cleanContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	return content
}
