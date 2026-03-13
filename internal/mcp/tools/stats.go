package tools

import (
	"encoding/json"
	"fmt"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
)

// StatsTools contains statistics-related MCP tools
type StatsTools struct {
	WorldsPath string
}

// NewStatsTools creates a new stats tools instance
func NewStatsTools(worldsPath string) *StatsTools {
	return &StatsTools{
		WorldsPath: worldsPath,
	}
}

// GetEntryStats returns statistics about journal entries in a world
func (t *StatsTools) GetEntryStats(params json.RawMessage) (interface{}, error) {
	var worldName string
	var username string
	var includeEntries bool

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(string); ok && wn != "" {
		worldName = wn
	}
	if un, ok := p["user"].(string); ok && un != "" {
		username = un
	}
	if ie, ok := p["include_entries"].(bool); ok {
		includeEntries = ie
	}

	if worldName == "" {
		return nil, fmt.Errorf("world is required")
	}

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	// Get all entries
	entries, err := repo.ListEntries()
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
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

	// Build statistics
	stats := journal.WorldStats{
		WorldName: worldName,
	}

	for _, entry := range entries {
		stats.TotalEntries++

		// Get pages for this entry
		pages, err := repo.ListPages(entry.ID)
		if err != nil {
			continue
		}

		stats.TotalPages += len(pages)

		// Update page type stats
		for _, page := range pages {
			stats.PageTypes.Add(page)
		}

		// Add entry-level stats
		if includeEntries {
			userID := ""
			if username != "" {
				userID, _ = repo.GetUserID(username)
			}

			entryStat := journal.EntryStats{
				ID:         entry.ID,
				Name:       entry.Name,
				PageCount:  len(pages),
				Permission: journal.CheckPermission(entry.Ownership, userID),
			}

			// Count page types for this entry
			for _, page := range pages {
				switch page.Type {
				case "text":
					entryStat.PageTypes.Text++
				case "image":
					entryStat.PageTypes.Image++
				case "video":
					entryStat.PageTypes.Video++
				case "pdf":
					entryStat.PageTypes.PDF++
				default:
					entryStat.PageTypes.Other++
				}
			}

			stats.Entries = append(stats.Entries, entryStat)
		}
	}

	// Build response
	response := map[string]interface{}{
		"world":        worldName,
		"totalEntries": stats.TotalEntries,
		"totalPages":   stats.TotalPages,
		"pageTypes": map[string]int{
			"text":  stats.PageTypes.Text,
			"image": stats.PageTypes.Image,
			"video": stats.PageTypes.Video,
			"pdf":   stats.PageTypes.PDF,
			"other": stats.PageTypes.Other,
			"total": stats.PageTypes.Total,
		},
	}

	if includeEntries && len(stats.Entries) > 0 {
		var entryList []map[string]interface{}
		for _, e := range stats.Entries {
			entryList = append(entryList, map[string]interface{}{
				"id":        e.ID,
				"name":      e.Name,
				"pageCount": e.PageCount,
				"pageTypes": map[string]int{
					"text":  e.PageTypes.Text,
					"image": e.PageTypes.Image,
					"video": e.PageTypes.Video,
					"pdf":   e.PageTypes.PDF,
					"other": e.PageTypes.Other,
				},
				"permission": int(e.Permission),
			})
		}
		response["entries"] = entryList
	}

	return response, nil
}
