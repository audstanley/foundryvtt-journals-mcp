package journal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/fvtt-journal-mcp/internal/leveldb"
)

func TestIntegration_Repository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Use testworld from the worlds directory
	worldsPath := "../../worlds"
	worldName := "testworld"

	// Check if world exists
	_, err := os.Stat(filepath.Join(worldsPath, worldName, "data", "journal"))
	if os.IsNotExist(err) {
		t.Skipf("World %s not found, skipping integration test", worldName)
	}

	repo, err := NewRepository(worldsPath, worldName)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	t.Run("list entries", func(t *testing.T) {
		entries, err := repo.ListEntries()
		if err != nil {
			t.Errorf("ListEntries() error = %v", err)
			return
		}
		if len(entries) == 0 {
			t.Log("No entries found in world")
		} else {
			t.Logf("Found %d entries", len(entries))
		}
	})

	t.Run("list pages", func(t *testing.T) {
		entries, err := repo.ListEntries()
		if err != nil {
			t.Skip("No entries to list pages from")
			return
		}

		if len(entries) == 0 {
			t.Skip("No entries to list pages from")
			return
		}

		firstEntry := entries[0]
		pages, err := repo.ListPages(firstEntry.ID)
		if err != nil {
			t.Errorf("ListPages() error = %v", err)
			return
		}
		t.Logf("Entry %s has %d pages", firstEntry.Name, len(pages))
	})

	t.Run("search entries", func(t *testing.T) {
		// Search for common terms
		results, err := repo.SearchEntries("Test")
		if err != nil {
			t.Errorf("SearchEntries() error = %v", err)
			return
		}
		t.Logf("Search 'Test' found %d entries", len(results))
	})
}

func TestIntegration_UserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	worldsPath := "../../worlds"
	worldName := "testworld"

	// Check if world exists
	usersPath := filepath.Join(worldsPath, worldName, "data", "users")
	if _, err := os.Stat(usersPath); os.IsNotExist(err) {
		t.Skipf("Users database not found in %s, skipping integration test", worldName)
	}

	repo, err := NewRepository(worldsPath, worldName)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	t.Run("get users", func(t *testing.T) {
		// Try to get users - we don't know the usernames, so just iterate
		var userCount int
		err := repo.usersDB.Iterate(func(key, value []byte) bool {
			_, ok := leveldb.ParseUserKey(key)
			if ok {
				userCount++
			}
			return true
		})
		if err != nil {
			t.Errorf("Failed to iterate users: %v", err)
			return
		}
		t.Logf("Found %d users", userCount)
	})
}
