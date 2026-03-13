package journal

// JournalEntry represents a Foundry VTT journal entry
type JournalEntry struct {
	ID        string                 `json:"_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Pages     []string               `json:"pages"`
	Sort      int                    `json:"sort"`
	Ownership map[string]int         `json:"ownership"`
	Folder    *string                `json:"folder"`
	Flags     map[string]interface{} `json:"flags"`
	System    map[string]interface{} `json:"system"`
}

// JournalPage represents a page within a journal entry
type JournalPage struct {
	ID        string                 `json:"_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // text, image, video, pdf
	Sort      int                    `json:"sort"`
	Ownership map[string]int         `json:"ownership"`
	Title     *TitleConfig           `json:"title"`
	Text      *TextContent           `json:"text"`
	Image     *ImageContent          `json:"image"`
	Video     *VideoContent          `json:"video"`
	Src       *string                `json:"src"`
	Markdown  string                 `json:"markdown"`
	Flags     map[string]interface{} `json:"flags"`
	System    map[string]interface{} `json:"system"`
}

// TitleConfig represents title display settings
type TitleConfig struct {
	Show  bool `json:"show"`
	Level int  `json:"level"`
}

// TextContent represents text/HTML content
type TextContent struct {
	Format  int    `json:"format"` // 1 = HTML
	Content string `json:"content"`
}

// ImageContent represents image metadata
type ImageContent struct {
	Caption string `json:"caption"`
	Src     string `json:"src"`
}

// VideoContent represents video metadata
type VideoContent struct {
	Controls bool    `json:"controls"`
	Volume   float64 `json:"volume"`
	Src      *string `json:"src"`
}

// User represents a Foundry VTT user
type User struct {
	ID           string                 `json:"_id"`
	Name         string                 `json:"name"`
	Role         int                    `json:"role"`
	Character    *string                `json:"character"`
	Pronouns     *string                `json:"pronouns"`
	Avatar       *string                `json:"avatar"`
	Color        *string                `json:"color"`
	Permissions  map[string][]string    `json:"permissions"`
	Flags        map[string]interface{} `json:"flags"`
	Hotbar       map[string]interface{} `json:"hotbar"`
	Stats        map[string]interface{} `json:"_stats"`
	Password     string                 `json:"password"`     // Hashed, for internal use only
	PasswordSalt string                 `json:"passwordSalt"` // Hashed, for internal use only
}

// PermissionLevel represents the permission level for journal access
type PermissionLevel int

const (
	NoAccess    PermissionLevel = -1
	Observer    PermissionLevel = 0
	Contributor PermissionLevel = 1
	Editor      PermissionLevel = 2
	Owner       PermissionLevel = 3
)

// CheckPermission determines if a user has access to a journal entry
// Returns the effective permission level for the user
func CheckPermission(ownership map[string]int, userID string) PermissionLevel {
	if ownership == nil {
		return Owner // Default to owner if no permissions set
	}

	// Check user-specific permission first
	if level, ok := ownership[userID]; ok {
		return PermissionLevel(level)
	}

	// Fall back to default permission
	if level, ok := ownership["default"]; ok {
		return PermissionLevel(level)
	}

	return Owner // No permissions set, default to owner
}

// HasAccess checks if a user has at least observer access
func HasAccess(ownership map[string]int, userID string) bool {
	return CheckPermission(ownership, userID) >= Observer
}

// ToInt converts PermissionLevel to int for use in maps
func ToInt(level PermissionLevel) int {
	return int(level)
}

// JournalEntryWithPages combines a journal entry with its page data
type JournalEntryWithPages struct {
	Entry JournalEntry
	Pages []JournalPage
}

// PageTypeStats tracks page type statistics
type PageTypeStats struct {
	Text  int
	Image int
	Video int
	PDF   int
	Other int
	Total int
}

// Add increments the statistics for a page
func (s *PageTypeStats) Add(page JournalPage) {
	s.Total++
	switch page.Type {
	case "text":
		s.Text++
	case "image":
		s.Image++
	case "video":
		s.Video++
	case "pdf":
		s.PDF++
	default:
		s.Other++
	}
}

// WorldStats contains statistics about all journals in a world
type WorldStats struct {
	WorldName    string
	TotalEntries int
	TotalPages   int
	PageTypes    PageTypeStats
	Entries      []EntryStats
}

// EntryStats contains statistics for a single entry
type EntryStats struct {
	ID           string
	Name         string
	PageCount    int
	PageTypes    PageTypeStats
	LastModified int64
	Permission   PermissionLevel
}
