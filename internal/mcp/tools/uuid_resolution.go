package tools

import (
	"encoding/json"
	"fmt"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
	"github.com/anomalyco/fvtt-journal-mcp/internal/mdx"
)

// UUIDResolutionTools contains UUID resolution-related MCP tools
type UUIDResolutionTools struct {
	WorldsPath string
}

// NewUUIDResolutionTools creates a new UUID resolution tools instance
func NewUUIDResolutionTools(worldsPath string) *UUIDResolutionTools {
	return &UUIDResolutionTools{
		WorldsPath: worldsPath,
	}
}

// ResolveUUID resolves a single @UUID{} reference to actual Foundry VTT data
func (t *UUIDResolutionTools) ResolveUUID(params json.RawMessage) (interface{}, error) {
	var uuidType, uuidID string
	var username string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if ut, ok := p["type"].(string); ok && ut != "" {
		uuidType = ut
	}
	if uid, ok := p["id"].(string); ok && uid != "" {
		uuidID = uid
	}
	if un, ok := p["user"].(string); ok && un != "" {
		username = un
	}

	if uuidType == "" {
		return nil, fmt.Errorf("type is required (e.g., Item, Actor, Compendium)")
	}
	if uuidID == "" {
		return nil, fmt.Errorf("id is required")
	}

	// Auto-discover world
	worlds, err := journal.ListWorlds(t.WorldsPath)
	if err != nil || len(worlds) == 0 {
		return nil, fmt.Errorf("no worlds available")
	}
	worldName := worlds[0]

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	// Try to resolve based on type
	var result map[string]interface{}

	switch uuidType {
	case "Item":
		result, err = t.resolveItem(repo, uuidID, username)
	case "Actor":
		result, err = t.resolveActor(repo, uuidID, username)
	case "Compendium":
		result, err = t.resolveCompendium(repo, uuidID, username)
	case "JournalEntry":
		result, err = t.resolveJournalEntry(repo, uuidID, username)
	case "Macro":
		result, err = t.resolveMacro(repo, uuidID, username)
	case "RollTable":
		result, err = t.resolveRollTable(repo, uuidID, username)
	default:
		// Generic resolution - try to find by ID in all compendia
		result, err = t.resolveGeneric(repo, uuidType, uuidID, username)
	}

	if err != nil {
		// Return partial info even if resolution fails
		return map[string]interface{}{
			"uuid":     fmt.Sprintf("%s/%s", uuidType, uuidID),
			"resolved": false,
			"type":     uuidType,
			"id":       uuidID,
			"error":    err.Error(),
		}, nil
	}

	result["uuid"] = fmt.Sprintf("%s/%s", uuidType, uuidID)
	result["resolved"] = true

	return result, nil
}

// ResolveUUIDsFromContent extracts and resolves all @UUID{} references from content
func (t *UUIDResolutionTools) ResolveUUIDsFromContent(params json.RawMessage) (interface{}, error) {
	var content string
	var username string

	var p map[string]interface{}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if c, ok := p["content"].(string); ok && c != "" {
		content = c
	}
	if un, ok := p["user"].(string); ok && un != "" {
		username = un
	}

	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Extract UUID references
	uuidRefs := mdx.ExtractUUIDLinks(content)

	if len(uuidRefs) == 0 {
		return map[string]interface{}{
			"content":      content,
			"references":   []map[string]interface{}{},
			"count":        0,
			"resolvedRefs": []map[string]interface{}{},
		}, nil
	}

	// Resolve each reference
	worlds, err := journal.ListWorlds(t.WorldsPath)
	if err != nil || len(worlds) == 0 {
		return nil, fmt.Errorf("no worlds available")
	}
	worldName := worlds[0]

	repo, err := journal.NewRepository(t.WorldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	var resolvedRefs []map[string]interface{}
	for _, ref := range uuidRefs {
		resolved, err := t.resolveGeneric(repo, ref.Type, ref.ID, username)
		if err != nil {
			resolved = map[string]interface{}{
				"type":     ref.Type,
				"id":       ref.ID,
				"display":  ref.Display,
				"uuid":     fmt.Sprintf("%s/%s", ref.Type, ref.ID),
				"resolved": false,
				"error":    err.Error(),
			}
		} else {
			resolved["uuid"] = fmt.Sprintf("%s/%s", ref.Type, ref.ID)
			resolved["resolved"] = true
			resolved["display"] = ref.Display
		}
		resolvedRefs = append(resolvedRefs, resolved)
	}

	return map[string]interface{}{
		"content":      content,
		"references":   uuidRefs,
		"count":        len(uuidRefs),
		"resolvedRefs": resolvedRefs,
	}, nil
}

// resolveItem resolves an Item UUID
func (t *UUIDResolutionTools) resolveItem(repo *journal.Repository, id string, username string) (map[string]interface{}, error) {
	// Try to find item in journal entries (items referenced in journals)
	// This is a simplified implementation - in real FVTT, items would be in compendia
	return map[string]interface{}{
		"type":   "Item",
		"id":     id,
		"status": "not_found_in_journal",
		"note":   "Items in journals are typically referenced, not stored directly",
	}, nil
}

// resolveActor resolves an Actor UUID
func (t *UUIDResolutionTools) resolveActor(repo *journal.Repository, id string, username string) (map[string]interface{}, error) {
	// Similar to items, actors in journals are referenced
	return map[string]interface{}{
		"type":   "Actor",
		"id":     id,
		"status": "not_found_in_journal",
		"note":   "Actors in journals are typically referenced, not stored directly",
	}, nil
}

// resolveCompendium resolves a Compendium UUID
func (t *UUIDResolutionTools) resolveCompendium(repo *journal.Repository, id string, username string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":   "Compendium",
		"id":     id,
		"status": "compendium_reference",
		"note":   "Compendium entries are external to journal system",
	}, nil
}

// resolveJournalEntry resolves a JournalEntry UUID
func (t *UUIDResolutionTools) resolveJournalEntry(repo *journal.Repository, id string, username string) (map[string]interface{}, error) {
	// Try to find journal entry by ID
	entries, err := repo.ListEntries()
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}

	for _, entry := range entries {
		if entry.ID == id {
			// Check permissions
			userID := ""
			if username != "" {
				userID, _ = repo.GetUserID(username)
			}

			if !journal.HasAccess(entry.Ownership, userID) {
				return map[string]interface{}{
					"type":    "JournalEntry",
					"id":      id,
					"status":  "no_access",
					"message": "User does not have access to this journal entry",
				}, nil
			}

			// Get full entry with pages
			entryWithPages, err := repo.GetEntryWithPages(id)
			if err != nil {
				return nil, fmt.Errorf("failed to get entry: %w", err)
			}

			// Convert content to Markdown
			mdxConverter := mdx.NewConverter()

			var pages []map[string]interface{}
			for _, page := range entryWithPages.Pages {
				pageData := map[string]interface{}{
					"id":   page.ID,
					"name": page.Name,
					"type": page.Type,
				}

				if page.Text != nil {
					mdxContent := mdxConverter.Convert(page.Text.Content)
					pageData["content"] = mdxContent
				}

				pages = append(pages, pageData)
			}

			return map[string]interface{}{
				"type":      "JournalEntry",
				"id":        id,
				"status":    "found",
				"entry":     entryWithPages.Entry,
				"pages":     pages,
				"pageCount": len(pages),
			}, nil
		}
	}

	return map[string]interface{}{
		"type":   "JournalEntry",
		"id":     id,
		"status": "not_found",
		"note":   "Journal entry not found in world",
	}, nil
}

// resolveMacro resolves a Macro UUID
func (t *UUIDResolutionTools) resolveMacro(repo *journal.Repository, id string, username string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":   "Macro",
		"id":     id,
		"status": "macro_reference",
		"note":   "Macros are external to journal system",
	}, nil
}

// resolveRollTable resolves a RollTable UUID
func (t *UUIDResolutionTools) resolveRollTable(repo *journal.Repository, id string, username string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":   "RollTable",
		"id":     id,
		"status": "rolltable_reference",
		"note":   "RollTables are external to journal system",
	}, nil
}

// resolveGeneric attempts to resolve a UUID generically
func (t *UUIDResolutionTools) resolveGeneric(repo *journal.Repository, uuidType string, uuidID string, username string) (map[string]interface{}, error) {
	// Try common types
	switch uuidType {
	case "JournalEntry":
		return t.resolveJournalEntry(repo, uuidID, username)
	case "Item", "Actor", "Compendium", "Macro", "RollTable":
		// These are typically external to journals
		return map[string]interface{}{
			"type":   uuidType,
			"id":     uuidID,
			"status": "external_reference",
			"note":   fmt.Sprintf("%ss are typically stored in compendia, not journals", uuidType),
		}, nil
	default:
		return map[string]interface{}{
			"type":   uuidType,
			"id":     uuidID,
			"status": "unknown_type",
			"note":   fmt.Sprintf("Unknown UUID type: %s", uuidType),
		}, nil
	}
}
