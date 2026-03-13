package leveldb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

// Reader provides LevelDB access for Foundry VTT worlds
type Reader struct {
	db   *leveldb.DB
	path string
}

// Open opens a LevelDB database at the given path
func Open(path string) (*Reader, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("LevelDB not found at %s", path)
		}
		return nil, fmt.Errorf("failed to open LevelDB: %w", err)
	}

	return &Reader{
		db:   db,
		path: path,
	}, nil
}

// Close closes the database connection
func (r *Reader) Close() error {
	return r.db.Close()
}

// Iterate iterates over all key-value pairs in the database
func (r *Reader) Iterate(filter func(key, value []byte) (continueIterating bool)) error {
	iter := r.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		if !filter(iter.Key(), iter.Value()) {
			break
		}
	}

	return iter.Error()
}

// Get retrieves a single value by key
func (r *Reader) Get(key []byte) ([]byte, error) {
	return r.db.Get(key, nil)
}

// ParseJournalKey parses a LevelDB key and returns its components
// Format: journal.{compendium_id}.{entry_id} or journal.pages.{compendium_id}.{page_id}
type JournalKey struct {
	Type         string // "journal" or "journal.pages"
	CompendiumID string
	EntityID     string
}

// ParseJournalKey parses a LevelDB key into its components
func ParseJournalKey(key []byte) (*JournalKey, bool) {
	keyStr := string(key)

	if !strings.HasPrefix(keyStr, "journal.") {
		return nil, false
	}

	parts := strings.SplitN(keyStr[8:], ".", 2) // Skip "journal." prefix
	if len(parts) != 2 {
		return nil, false
	}

	// Check if it's journal.pages
	isPage := strings.HasPrefix(parts[0], "pages")

	if isPage {
		// Format: journal.pages.{compendium_id}.{page_id}
		parts2 := strings.SplitN(parts[1], ".", 2)
		if len(parts2) != 2 {
			return nil, false
		}
		return &JournalKey{
			Type:         "journal.pages",
			CompendiumID: parts2[0],
			EntityID:     parts2[1],
		}, true
	}

	// Format: journal.{compendium_id}.{entry_id}
	return &JournalKey{
		Type:         "journal",
		CompendiumID: parts[0],
		EntityID:     parts[1],
	}, true
}

// ParseUserKey parses a LevelDB user key
// Format: !users!{user_id}
type UserKey struct {
	UserID string
}

// ParseUserKey parses a user key
func ParseUserKey(key []byte) (*UserKey, bool) {
	keyStr := string(key)

	if !strings.HasPrefix(keyStr, "!users!") {
		return nil, false
	}

	userID := keyStr[7:] // Skip "!users!" prefix
	return &UserKey{
		UserID: userID,
	}, true
}

// CleanJSON removes control characters from JSON data
func CleanJSON(data []byte) []byte {
	// Remove null bytes and other control characters
	cleaned := make([]byte, 0, len(data))
	for _, b := range data {
		if b == 0 {
			continue
		}
		// Keep printable ASCII and common whitespace
		if b >= 32 && b <= 126 || b == '\t' || b == '\n' || b == '\r' {
			cleaned = append(cleaned, b)
		}
	}
	return cleaned
}

// ParseJSON decodes JSON data into a map
func ParseJSON(data []byte) (map[string]interface{}, error) {
	cleaned := CleanJSON(data)

	var result map[string]interface{}
	if err := json.Unmarshal(cleaned, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// FindWorlds searches for LevelDB journal databases in a worlds directory
func FindWorlds(worldsPath string) ([]string, error) {
	entries, err := os.ReadDir(worldsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read worlds directory: %w", err)
	}

	var worlds []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it has a journal database
		journalPath := fmt.Sprintf("%s/%s/data/journal", worldsPath, entry.Name())
		if _, err := os.Stat(journalPath); err == nil {
			worlds = append(worlds, entry.Name())
		}
	}

	return worlds, nil
}

// OpenWorldJournal opens the journal database for a specific world
func OpenWorldJournal(worldsPath, worldName string) (*Reader, error) {
	journalPath := fmt.Sprintf("%s/%s/data/journal", worldsPath, worldName)
	return Open(journalPath)
}

// OpenWorldUsers opens the users database for a specific world
func OpenWorldUsers(worldsPath, worldName string) (*Reader, error) {
	usersPath := fmt.Sprintf("%s/%s/data/users", worldsPath, worldName)
	return Open(usersPath)
}
