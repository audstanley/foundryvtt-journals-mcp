package mdx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/yaml.v3"
)

// Generator creates MDX files from journal entries
type Generator struct {
	outputPath string
	worldsPath string
	worldName  string
	converter  *Converter
}

// NewGenerator creates a new MDX generator
func NewGenerator(outputPath, worldsPath, worldName string) *Generator {
	return &Generator{
		outputPath: outputPath,
		worldsPath: worldsPath,
		worldName:  worldName,
		converter:  NewConverter(),
	}
}

// Export exports all journals from a repository to MDX files
func (g *Generator) Export(repo *journal.Repository, worldName string) error {
	// Create output directory structure: output/world-name/
	worldDir := filepath.Join(g.outputPath, sanitizeFilename(worldName))
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get all entries
	entries, err := repo.ListEntries()
	if err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}

	// Build parent-child relationships
	entryMap := make(map[string]*journal.JournalEntry)
	for i := range entries {
		entryMap[entries[i].ID] = &entries[i]
	}

	// Process each entry
	for _, entry := range entries {
		if err := g.exportEntry(repo, entry, worldDir, entryMap); err != nil {
			return fmt.Errorf("failed to export entry %s: %w", entry.ID, err)
		}
	}

	return nil
}

// exportEntry exports a single journal entry to MDX files
func (g *Generator) exportEntry(repo *journal.Repository, entry journal.JournalEntry, worldDir string, entryMap map[string]*journal.JournalEntry) error {
	// Get pages for this entry
	pages, err := repo.ListPages(entry.ID)
	if err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}

	// Create entry directory with folder support
	var entryDir string
	if entry.Folder != nil && *entry.Folder != "" {
		// Resolve folder ID to name
		folderName := resolveFolderID(*entry.Folder, g.worldsPath, g.worldName)
		sanitizedFolder := sanitizeFilename(folderName)
		sanitizedEntry := sanitizeFilename(entry.Name)
		entryDir = filepath.Join(worldDir, sanitizedFolder, sanitizedEntry)
	} else {
		// Flat structure: world-name/entry-name/
		entryDir = filepath.Join(worldDir, sanitizeFilename(entry.Name))
	}

	if err := os.MkdirAll(entryDir, 0755); err != nil {
		return fmt.Errorf("failed to create entry directory: %w", err)
	}

	// Export each page as a separate MDX file
	for _, page := range pages {
		if err := g.exportPage(entry, page, entryDir, entryMap); err != nil {
			return fmt.Errorf("failed to export page %s: %w", page.ID, err)
		}
	}

	return nil
}

// exportPage exports a single page to an MDX file
func (g *Generator) exportPage(entry journal.JournalEntry, page journal.JournalPage, entryDir string, entryMap map[string]*journal.JournalEntry) error {
	// Generate content based on page type
	var content string
	switch page.Type {
	case "text":
		if page.Text != nil {
			content = g.converter.Convert(page.Text.Content)
		}
	case "image":
		if page.Image != nil {
			if page.Image.Src != "" {
				content = "![Image](" + page.Image.Src + ")"
			}
		}
	case "video":
		if page.Video != nil {
			content = "<video controls>\n"
			if page.Video.Src != nil {
				content += "  <source src=\"" + *page.Video.Src + "\" type=\"video/mp4\">\n"
			}
			content += "</video>"
		}
	case "pdf":
		if page.Src != nil {
			content = "[PDF Document](" + *page.Src + ")"
		}
	default:
		content = fmt.Sprintf("Content type not supported: %s", page.Type)
	}

	// Extract UUIDs from text content for frontmatter
	var contentForUUIDs string
	if page.Text != nil {
		contentForUUIDs = page.Text.Content
	}

	// Generate frontmatter with UUIDs
	frontmatter, err := g.generateFrontmatter(entry, page, contentForUUIDs, entryMap)
	if err != nil {
		return fmt.Errorf("failed to generate frontmatter: %w", err)
	}

	// Combine frontmatter and content
	mdxContent := "---\n" + frontmatter + "\n---\n\n" + content

	// Write to file
	filename := filepath.Join(entryDir, sanitizeFilename(page.Name)+".mdx")
	if err := os.WriteFile(filename, []byte(mdxContent), 0644); err != nil {
		return fmt.Errorf("failed to write MDX file: %w", err)
	}

	return nil
}

// generateFrontmatter creates YAML frontmatter for an MDX file
func (g *Generator) generateFrontmatter(entry journal.JournalEntry, page journal.JournalPage, content string, entryMap map[string]*journal.JournalEntry) (string, error) {
	frontmatter := map[string]interface{}{
		"title": page.Name,
		"entry": entry.Name,
		"type":  page.Type,
		"page":  page.ID,
		"sort":  page.Sort,
	}

	// Add entry-level metadata
	frontmatter["entry_sort"] = entry.Sort
	if entry.Folder != nil && *entry.Folder != "" {
		frontmatter["folder"] = *entry.Folder
	}

	// Add UUID references from content
	uuidRefs := ExtractUUIDLinks(content)
	if len(uuidRefs) > 0 {
		uuidList := make([]map[string]interface{}, len(uuidRefs))
		for i, ref := range uuidRefs {
			uuidList[i] = map[string]interface{}{
				"type":    ref.Type,
				"id":      ref.ID,
				"display": ref.Display,
				"uuid":    fmt.Sprintf("%s/%s", ref.Type, ref.ID),
			}
		}
		frontmatter["uuid_references"] = uuidList
	}

	// Add permission metadata
	frontmatter["ownership"] = entry.Ownership

	// Add page ownership
	if len(page.Ownership) > 0 {
		frontmatter["page_ownership"] = page.Ownership
	}

	// Add title config if present
	if page.Title != nil {
		frontmatter["title_config"] = map[string]interface{}{
			"show":  page.Title.Show,
			"level": page.Title.Level,
		}
	}

	// Add parent/child relationships
	if entry.Folder != nil && *entry.Folder != "" {
		// Find sibling entries with same folder
		var siblings []string
		for _, e := range entryMap {
			if e.Folder != nil && *e.Folder == *entry.Folder && e.ID != entry.ID {
				siblings = append(siblings, e.Name)
			}
		}
		if len(siblings) > 0 {
			frontmatter["siblings"] = siblings
		}
	}

	// Convert to YAML
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(frontmatter); err != nil {
		return "", fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to close encoder: %w", err)
	}

	return buf.String(), nil
}

// resolveFolderID looks up the folder name from its ID
func resolveFolderID(folderID string, worldsPath, worldName string) string {
	if folderID == "" {
		return ""
	}

	foldersPath := fmt.Sprintf("%s/%s/data/folders", worldsPath, worldName)
	db, err := leveldb.OpenFile(foldersPath, nil)
	if err != nil {
		return folderID // Return ID if database not found
	}
	defer db.Close()

	key := fmt.Sprintf("!folders!%s", folderID)
	value, err := db.Get([]byte(key), nil)
	if err != nil {
		return folderID // Return ID if folder not found
	}

	var data map[string]interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return folderID
	}

	if name, ok := data["name"].(string); ok && name != "" {
		return name
	}

	return folderID
}

// sanitizeFilename creates a safe filename from user input
func sanitizeFilename(name string) string {
	// Replace problematic characters
	sanitized := name
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, "\\", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.ReplaceAll(sanitized, "*", "-")
	sanitized = strings.ReplaceAll(sanitized, "?", "-")
	sanitized = strings.ReplaceAll(sanitized, "\"", "-")
	sanitized = strings.ReplaceAll(sanitized, "<", "-")
	sanitized = strings.ReplaceAll(sanitized, ">", "-")
	sanitized = strings.ReplaceAll(sanitized, "|", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")

	// Limit length
	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}

	// Remove leading/trailing whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Ensure not empty
	if sanitized == "" {
		sanitized = "untitled"
	}

	return sanitized
}
