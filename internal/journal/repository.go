package journal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anomalyco/fvtt-journal-mcp/internal/leveldb"
	"github.com/anomalyco/fvtt-journal-mcp/internal/ndjson"
)

// Repository provides access to journal data
type Repository struct {
	worldName  string
	journalDB  *leveldb.Reader
	usersDB    *leveldb.Reader
	ndjsonDB   *ndjson.Reader
	worldsPath string
}

// NewRepository creates a new journal repository for a world
func NewRepository(worldsPath, worldName string) (*Repository, error) {
	journalDB, err := leveldb.OpenWorldJournal(worldsPath, worldName)
	if err != nil {
		return nil, fmt.Errorf("failed to open journal database: %w", err)
	}

	usersDB, err := leveldb.OpenWorldUsers(worldsPath, worldName)
	if err != nil {
		journalDB.Close()
		return nil, fmt.Errorf("failed to open users database: %w", err)
	}

	ndjsonDB, err := ndjson.Open(worldsPath, worldName)
	if err != nil {
		journalDB.Close()
		usersDB.Close()
		return nil, fmt.Errorf("failed to open ndjson database: %w", err)
	}

	return &Repository{
		worldName:  worldName,
		journalDB:  journalDB,
		usersDB:    usersDB,
		ndjsonDB:   ndjsonDB,
		worldsPath: worldsPath,
	}, nil
}

// Close closes all database connections
func (r *Repository) Close() error {
	var firstErr error
	if r.journalDB != nil {
		if err := r.journalDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if r.usersDB != nil {
		if err := r.usersDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if r.ndjsonDB != nil {
		if err := r.ndjsonDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// ListEntries returns all journal entries in the world
func (r *Repository) ListEntries() ([]JournalEntry, error) {
	var entries []JournalEntry

	err := r.journalDB.Iterate(func(key, value []byte) bool {
		jKey, ok := leveldb.ParseJournalKey(key)
		if !ok || jKey.Type != "journal" {
			return true
		}

		data, err := r.journalDB.Get(key)
		if err != nil {
			return true
		}

		entry, err := parseJournalEntry(data)
		if err != nil {
			return true
		}

		entries = append(entries, entry)
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}

	return entries, nil
}

// GetEntry retrieves a specific journal entry
func (r *Repository) GetEntry(entryID string) (*JournalEntry, error) {
	var entry *JournalEntry

	err := r.journalDB.Iterate(func(key, value []byte) bool {
		jKey, ok := leveldb.ParseJournalKey(key)
		if !ok || jKey.Type != "journal" {
			return true
		}

		if jKey.EntityID == entryID {
			data, err := r.journalDB.Get(key)
			if err != nil {
				return false
			}

			parsed, err := parseJournalEntry(data)
			if err != nil {
				return false
			}

			entry = &parsed
			return false
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if entry == nil {
		return nil, fmt.Errorf("entry not found: %s", entryID)
	}

	return entry, nil
}

// GetEntryWithPages retrieves a journal entry along with all its pages
func (r *Repository) GetEntryWithPages(entryID string) (*JournalEntryWithPages, error) {
	entry, err := r.GetEntry(entryID)
	if err != nil {
		return nil, err
	}

	pages, err := r.getPagesForEntry(entry)
	if err != nil {
		return nil, err
	}

	return &JournalEntryWithPages{
		Entry: *entry,
		Pages: pages,
	}, nil
}

// ListPages returns all pages for a specific entry
func (r *Repository) ListPages(entryID string) ([]JournalPage, error) {
	entry, err := r.GetEntry(entryID)
	if err != nil {
		return nil, err
	}

	return r.getPagesForEntry(entry)
}

// GetPage retrieves a specific page by its ID
func (r *Repository) GetPage(pageID string) (*JournalPage, error) {
	var page *JournalPage

	err := r.journalDB.Iterate(func(key, value []byte) bool {
		jKey, ok := leveldb.ParseJournalKey(key)
		if !ok || jKey.Type != "journal.pages" {
			return true
		}

		if jKey.EntityID == pageID {
			data, err := r.journalDB.Get(key)
			if err != nil {
				return false
			}

			parsed, err := parseJournalPage(data)
			if err != nil {
				return false
			}

			page = &parsed
			return false
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	if page == nil {
		return nil, fmt.Errorf("page not found: %s", pageID)
	}

	return page, nil
}

// SearchEntries searches for journal entries by name
func (r *Repository) SearchEntries(query string) ([]JournalEntry, error) {
	var results []JournalEntry

	err := r.journalDB.Iterate(func(key, value []byte) bool {
		jKey, ok := leveldb.ParseJournalKey(key)
		if !ok || jKey.Type != "journal" {
			return true
		}

		data, err := r.journalDB.Get(key)
		if err != nil {
			return true
		}

		entry, err := parseJournalEntry(data)
		if err != nil {
			return true
		}

		// Case-sensitive partial match
		if strings.Contains(entry.Name, query) {
			results = append(results, entry)
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search entries: %w", err)
	}

	return results, nil
}

// SearchPages searches for pages within content
func (r *Repository) SearchPages(query string, entryID *string) ([]JournalPage, error) {
	var results []JournalPage

	err := r.journalDB.Iterate(func(key, value []byte) bool {
		jKey, ok := leveldb.ParseJournalKey(key)
		if !ok || jKey.Type != "journal.pages" {
			return true
		}

		// If entryID is specified, check if page belongs to that entry
		if entryID != nil {
			entry, err := r.GetEntry(*entryID)
			if err != nil {
				return true
			}

			// Check if page is in this entry's pages list
			found := false
			for _, pageID := range entry.Pages {
				if pageID == jKey.EntityID {
					found = true
					break
				}
			}
			if !found {
				return true
			}
		}

		data, err := r.journalDB.Get(key)
		if err != nil {
			return true
		}

		page, err := parseJournalPage(data)
		if err != nil {
			return true
		}

		// Case-sensitive partial match in text content
		if page.Text != nil && strings.Contains(page.Text.Content, query) {
			results = append(results, page)
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search pages: %w", err)
	}

	return results, nil
}

// GetUser retrieves a user by their name
func (r *Repository) GetUser(username string) (*User, error) {
	var user *User

	err := r.usersDB.Iterate(func(key, value []byte) bool {
		_, ok := leveldb.ParseUserKey(key)
		if !ok {
			return true
		}

		data, err := r.usersDB.Get(key)
		if err != nil {
			return true
		}

		parsed, err := parseUser(data)
		if err != nil {
			return true
		}

		if parsed.Name == username {
			user = &parsed
			return false
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	return user, nil
}

// GetUserID retrieves a user's ID by their name
func (r *Repository) GetUserID(username string) (string, error) {
	user, err := r.GetUser(username)
	if err != nil {
		return "", err
	}

	// Extract user ID from the !users!{id} format
	if len(user.ID) > 7 {
		return user.ID[7:], nil
	}
	return user.ID, nil
}

// ListWorlds returns all available worlds
func ListWorlds(worldsPath string) ([]string, error) {
	return leveldb.FindWorlds(worldsPath)
}

// SearchNDJSON searches the NDJSON back compendium
func (r *Repository) SearchNDJSON(query string) ([]map[string]interface{}, error) {
	if r.ndjsonDB == nil {
		return nil, fmt.Errorf("ndjson database not available")
	}
	return r.ndjsonDB.Search(query)
}

// GetNDJSONByID retrieves an entity from NDJSON by type and ID
func (r *Repository) GetNDJSONByID(entityType, id string) (map[string]interface{}, bool) {
	if r.ndjsonDB == nil {
		return nil, false
	}
	return r.ndjsonDB.GetByID(entityType, id)
}

// SearchAll performs a unified search across LevelDB (journals) and NDJSON (back compendium)
func (r *Repository) SearchAll(query string) (*SearchResults, error) {
	journalEntries, err := r.SearchEntries(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search journal entries: %w", err)
	}

	journalPages, err := r.SearchPages(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search journal pages: %w", err)
	}

	ndjsonEntities, err := r.SearchNDJSON(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search ndjson: %w", err)
	}

	return MergeSearchResults(journalEntries, journalPages, ndjsonEntities, query), nil
}

// Helper functions

func parseJournalEntry(data []byte) (JournalEntry, error) {
	jsonData, err := leveldb.ParseJSON(data)
	if err != nil {
		return JournalEntry{}, err
	}

	var entry JournalEntry
	entryBytes, _ := json.Marshal(jsonData)
	json.Unmarshal(entryBytes, &entry)

	return entry, nil
}

func parseJournalPage(data []byte) (JournalPage, error) {
	jsonData, err := leveldb.ParseJSON(data)
	if err != nil {
		return JournalPage{}, err
	}

	var page JournalPage
	pageBytes, _ := json.Marshal(jsonData)
	json.Unmarshal(pageBytes, &page)

	return page, nil
}

func parseUser(data []byte) (User, error) {
	jsonData, err := leveldb.ParseJSON(data)
	if err != nil {
		return User{}, err
	}

	var user User
	userBytes, _ := json.Marshal(jsonData)
	json.Unmarshal(userBytes, &user)

	return user, nil
}

func (r *Repository) getPagesForEntry(entry *JournalEntry) ([]JournalPage, error) {
	var pages []JournalPage

	for _, pageID := range entry.Pages {
		page, err := r.GetPage(pageID)
		if err != nil {
			continue // Skip pages that can't be loaded
		}
		pages = append(pages, *page)
	}

	return pages, nil
}
