package tools

import (
	"encoding/json"

	"github.com/anomalyco/fvtt-journal-mcp/internal/mcp"
)

// Registry holds all MCP tools for the server
type Registry struct {
	worldsPath string
	tools      []mcp.Tool
}

// NewRegistry creates a new tool registry
func NewRegistry(worldsPath string) *Registry {
	r := &Registry{
		worldsPath: worldsPath,
	}
	r.registerAllTools()
	return r
}

// registerAllTools initializes and registers all tool categories
func (r *Registry) registerAllTools() {
	traversal := NewTraversalTools(r.worldsPath)
	unifiedSearch := NewUnifiedSearchTools(r.worldsPath)
	uuid := NewUUIDResolutionTools(r.worldsPath)

	r.tools = []mcp.Tool{
		// Traversal tools
		{
			Name:        "list_worlds",
			Description: "List all available Foundry VTT worlds",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"world_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to worlds directory (optional, defaults to server config)",
					},
				},
			},
			Handler: traversal.ListWorlds,
		},
		// Search tools - Unified across LevelDB and NDJSON
		{
			Name:        "search_all",
			Description: "Search across both LevelDB (journals) and NDJSON (back compendium) - returns unified results with source tagging",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{
					"query",
				},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (searches all databases)",
					},
				},
			},
			Handler: unifiedSearch.SearchAll,
		},
		{
			Name:        "search_compendium",
			Description: "Search only the NDJSON back compendium (Actors, Items, Journals)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{
					"query",
				},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for compendium data",
					},
					"entity_type": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter by entity type (actors, items, journals)",
					},
				},
			},
			Handler: unifiedSearch.SearchCompendium,
		},
		{
			Name:        "search_journals",
			Description: "Search only LevelDB journal entries",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{
					"query",
				},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for journal entry names",
					},
				},
			},
			Handler: unifiedSearch.SearchJournals,
		},
		{
			Name:        "search_journal_pages",
			Description: "Search only LevelDB journal page content",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{
					"query",
				},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for journal page content",
					},
				},
			},
			Handler: unifiedSearch.SearchJournalPages,
		},
		// UUID tools - Updated to remove world parameter
		{
			Name:        "resolve_uuid",
			Description: "Resolve single @UUID{} reference to Foundry VTT data",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{
					"type",
					"id",
				},
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"description": "UUID type (Item, Actor, Compendium, JournalEntry, Macro, RollTable)",
					},
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID ID to resolve",
					},
				},
			},
			Handler: uuid.ResolveUUID,
		},
		{
			Name:        "resolve_uuids_from_content",
			Description: "Extract and resolve all UUIDs in content",
			InputSchema: map[string]interface{}{
				"type": "object",
				"required": []string{
					"content",
				},
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to extract UUIDs from",
					},
				},
			},
			Handler: uuid.ResolveUUIDsFromContent,
		},
	}
}

// RegisterAll registers all tools with the server
func (r *Registry) RegisterAll(server *mcp.Server) {
	for _, tool := range r.tools {
		server.RegisterTool(tool)
	}
}

// GetToolsList returns all tools for the tools/list MCP method
func (r *Registry) GetToolsList() []map[string]interface{} {
	toolsList := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		toolsList = append(toolsList, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}
	return toolsList
}

// GetTools returns the list of registered tools
func (r *Registry) GetTools() []mcp.Tool {
	return r.tools
}

// GetToolByName returns a specific tool by name
func (r *Registry) GetToolByName(name string) (*mcp.Tool, bool) {
	for _, tool := range r.tools {
		if tool.Name == name {
			return &tool, true
		}
	}
	return nil, false
}

// GetToolSchema returns the JSON schema for a tool
func (r *Registry) GetToolSchema(name string) (json.RawMessage, bool) {
	tool, exists := r.GetToolByName(name)
	if !exists {
		return nil, false
	}

	schema, err := json.Marshal(tool.InputSchema)
	if err != nil {
		return nil, false
	}

	return schema, true
}
