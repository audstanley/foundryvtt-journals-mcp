package journal

import (
	"fmt"
	"strings"
)

// SearchResult represents a unified search result from any source
type SearchResult struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Source  string                 `json:"source"` // "LevelDB" or "NDJSON"
	UUID    string                 `json:"uuid"`
	Content string                 `json:"content,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// SearchResults represents a collection of search results with deduplication
type SearchResults struct {
	Results   []SearchResult `json:"results"`
	Count     int            `json:"count"`
	UniqueIDs int            `json:"unique_ids"`
	Query     string         `json:"query"`
}

// MergeSearchResults combines LevelDB and NDJSON results with deduplication
func MergeSearchResults(
	journalEntries []JournalEntry,
	journalPages []JournalPage,
	ndjsonEntities []map[string]interface{},
	query string,
) *SearchResults {
	seenUUIDs := make(map[string]bool)
	results := make([]SearchResult, 0)

	// Process LevelDB journal entries
	for _, entry := range journalEntries {
		uuid := fmt.Sprintf("JournalEntry/%s", entry.ID)
		if !seenUUIDs[uuid] {
			seenUUIDs[uuid] = true
			results = append(results, SearchResult{
				ID:     entry.ID,
				Name:   entry.Name,
				Type:   "JournalEntry",
				Source: "LevelDB",
				UUID:   uuid,
				Meta: map[string]interface{}{
					"pageCount": len(entry.Pages),
				},
			})
		}
	}

	// Process LevelDB journal pages
	for _, page := range journalPages {
		uuid := fmt.Sprintf("JournalPage/%s", page.ID)
		if !seenUUIDs[uuid] {
			seenUUIDs[uuid] = true
			result := SearchResult{
				ID:     page.ID,
				Name:   page.Name,
				Type:   page.Type,
				Source: "LevelDB",
				UUID:   uuid,
			}
			if page.Text != nil && page.Text.Content != "" {
				content := page.Text.Content
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				result.Content = cleanContent(content)
			}
			results = append(results, result)
		}
	}

	// Process NDJSON entities
	for _, entity := range ndjsonEntities {
		entityID, _ := entity["_id"].(string)
		entityName, _ := entity["name"].(string)
		entityType, _ := entity["type"].(string)

		uuid := fmt.Sprintf("%s/%s", strings.Title(entityType), entityID)
		if !seenUUIDs[uuid] {
			seenUUIDs[uuid] = true
			results = append(results, SearchResult{
				ID:     entityID,
				Name:   entityName,
				Type:   strings.Title(entityType),
				Source: "NDJSON",
				UUID:   uuid,
				Meta:   entity,
			})
		}
	}

	return &SearchResults{
		Results:   results,
		Count:     len(results),
		UniqueIDs: len(seenUUIDs),
		Query:     query,
	}
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
