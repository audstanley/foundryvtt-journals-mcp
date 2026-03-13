package tools

import (
	"encoding/json"
	"fmt"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
)

// TraversalTools contains all traversal-related MCP tools
type TraversalTools struct {
	WorldsPath string
}

// NewTraversalTools creates a new traversal tools instance
func NewTraversalTools(worldsPath string) *TraversalTools {
	return &TraversalTools{
		WorldsPath: worldsPath,
	}
}

// ListWorlds returns all available worlds
func (t *TraversalTools) ListWorlds(params json.RawMessage) (interface{}, error) {
	// Parse params (optional world_path)
	var worldPath string
	if len(params) > 0 {
		var p map[string]interface{}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if wp, ok := p["world_path"].(string); ok && wp != "" {
			worldPath = wp
		}
	}
	if worldPath == "" {
		worldPath = t.WorldsPath
	}

	worlds, err := journal.ListWorlds(worldPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list worlds: %w", err)
	}

	return map[string]interface{}{
		"worlds": worlds,
		"count":  len(worlds),
	}, nil
}

// ListEntries returns all journal entries for a world
func (t *TraversalTools) ListEntries(params json.RawMessage) (interface{}, error) {
	var worldName string
	var username string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(*json.Number); ok {
		worldName = wn.String()
	} else if wnStr, ok := p["world"].(string); ok {
		worldName = wnStr
	}

	if un, ok := p["user"].(string); ok && un != "" {
		username = un
	}

	// Open repository
	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

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

	// Build response with permission info
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
			"type":       entry.Type,
			"pageCount":  len(entry.Pages),
			"permission": int(permission),
			"hasAccess":  journal.HasAccess(entry.Ownership, username),
		})
	}

	return map[string]interface{}{
		"world":   worldName,
		"entries": result,
		"count":   len(result),
	}, nil
}

// GetEntry returns a specific journal entry
func (t *TraversalTools) GetEntry(params json.RawMessage) (interface{}, error) {
	var worldName, entryID string
	var username string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(string); ok && wn != "" {
		worldName = wn
	}
	if eid, ok := p["entry_id"].(string); ok && eid != "" {
		entryID = eid
	}
	if un, ok := p["user"].(string); ok && un != "" {
		username = un
	}

	if worldName == "" {
		return nil, fmt.Errorf("world is required")
	}
	if entryID == "" {
		return nil, fmt.Errorf("entry_id is required")
	}

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	entryWithPages, err := repo.GetEntryWithPages(entryID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	userID := ""
	if username != "" {
		userID, _ = repo.GetUserID(username)
	}

	permission := journal.CheckPermission(entryWithPages.Entry.Ownership, userID)

	// Build page list
	var pages []map[string]interface{}
	for _, page := range entryWithPages.Pages {
		pageData := map[string]interface{}{
			"id":   page.ID,
			"name": page.Name,
			"type": page.Type,
		}

		// Add content based on type
		if page.Type == "text" && page.Text != nil {
			pageData["content"] = page.Text.Content
		} else if page.Type == "image" && page.Image != nil {
			pageData["content"] = page.Image.Caption
			pageData["src"] = page.Image.Src
		} else if page.Type == "video" && page.Video != nil {
			pageData["controls"] = page.Video.Controls
			pageData["volume"] = page.Video.Volume
			if page.Video.Src != nil {
				pageData["src"] = *page.Video.Src
			}
		}

		pages = append(pages, pageData)
	}

	return map[string]interface{}{
		"entry": map[string]interface{}{
			"id":         entryWithPages.Entry.ID,
			"name":       entryWithPages.Entry.Name,
			"type":       entryWithPages.Entry.Type,
			"pageCount":  len(entryWithPages.Pages),
			"permission": int(permission),
			"hasAccess":  journal.HasAccess(entryWithPages.Entry.Ownership, userID),
		},
		"pages": pages,
	}, nil
}

// ListPages returns all pages for a journal entry
func (t *TraversalTools) ListPages(params json.RawMessage) (interface{}, error) {
	var worldName, entryID string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(string); ok && wn != "" {
		worldName = wn
	}
	if eid, ok := p["entry_id"].(string); ok && eid != "" {
		entryID = eid
	}

	if worldName == "" {
		return nil, fmt.Errorf("world is required")
	}
	if entryID == "" {
		return nil, fmt.Errorf("entry_id is required")
	}

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	pages, err := repo.ListPages(entryID)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, page := range pages {
		result = append(result, map[string]interface{}{
			"id":   page.ID,
			"name": page.Name,
			"type": page.Type,
		})
	}

	return map[string]interface{}{
		"entry_id": entryID,
		"pages":    result,
		"count":    len(result),
	}, nil
}

// GetPage returns a specific page
func (t *TraversalTools) GetPage(params json.RawMessage) (interface{}, error) {
	var worldName, pageID string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if wn, ok := p["world"].(string); ok && wn != "" {
		worldName = wn
	}
	if pid, ok := p["page_id"].(string); ok && pid != "" {
		pageID = pid
	}

	if worldName == "" {
		return nil, fmt.Errorf("world is required")
	}
	if pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	page, err := repo.GetPage(pageID)
	if err != nil {
		return nil, err
	}

	pageData := map[string]interface{}{
		"id":   page.ID,
		"name": page.Name,
		"type": page.Type,
	}

	// Add content based on type
	switch page.Type {
	case "text":
		if page.Text != nil {
			pageData["content"] = page.Text.Content
			pageData["format"] = page.Text.Format
		}
	case "image":
		if page.Image != nil {
			pageData["caption"] = page.Image.Caption
			pageData["src"] = page.Image.Src
		}
	case "video":
		if page.Video != nil {
			pageData["controls"] = page.Video.Controls
			pageData["volume"] = page.Video.Volume
			if page.Video.Src != nil {
				pageData["src"] = *page.Video.Src
			}
		}
	case "pdf":
		pageData["src"] = page.Src
	}

	if page.Title != nil {
		pageData["title"] = map[string]interface{}{
			"show":  page.Title.Show,
			"level": page.Title.Level,
		}
	}

	return pageData, nil
}
