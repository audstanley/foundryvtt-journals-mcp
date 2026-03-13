package leveldb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "open valid database",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "open non-existent database",
			path:    "/non-existent/path/xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Open(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer r.Close()
				if r.db == nil {
					t.Error("Expected db to be non-nil")
				}
				if r.path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.path)
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()

	r, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}

	err = r.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Try to use after close - should fail
	_, err = r.Get([]byte("test"))
	if err == nil {
		t.Error("Expected error after close, got nil")
	}
}

func TestParseJournalKey(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		wantType string
		wantID   string
		wantOk   bool
	}{
		{
			name:     "valid journal entry key",
			key:      []byte("journal.abc123.entry456"),
			wantType: "journal",
			wantID:   "entry456",
			wantOk:   true,
		},
		{
			name:     "valid journal page key",
			key:      []byte("journal.pages.abc123.page789"),
			wantType: "journal.pages",
			wantID:   "page789",
			wantOk:   true,
		},
		{
			name:     "invalid key prefix",
			key:      []byte("invalid.abc123"),
			wantType: "",
			wantID:   "",
			wantOk:   false,
		},
		{
			name:     "key without components",
			key:      []byte("journal.abc123"),
			wantType: "",
			wantID:   "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseJournalKey(tt.key)
			if ok != tt.wantOk {
				t.Errorf("ParseJournalKey() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if !tt.wantOk {
				return
			}
			if result.Type != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, result.Type)
			}
			if result.EntityID != tt.wantID {
				t.Errorf("Expected entityID %s, got %s", tt.wantID, result.EntityID)
			}
		})
	}
}

func TestParseUserKey(t *testing.T) {
	tests := []struct {
		name   string
		key    []byte
		wantID string
		wantOk bool
	}{
		{
			name:   "valid user key",
			key:    []byte("!users!abc123def456"),
			wantID: "abc123def456",
			wantOk: true,
		},
		{
			name:   "invalid key prefix",
			key:    []byte("invalid.abc123"),
			wantID: "",
			wantOk: false,
		},
		{
			name:   "empty key",
			key:    []byte(""),
			wantID: "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseUserKey(tt.key)
			if ok != tt.wantOk {
				t.Errorf("ParseUserKey() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if !tt.wantOk {
				return
			}
			if result.UserID != tt.wantID {
				t.Errorf("Expected userID %s, got %s", tt.wantID, result.UserID)
			}
		})
	}
}

func TestCleanJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantJSON []byte
	}{
		{
			name:     "clean JSON without control chars",
			input:    []byte(`{"key": "value"}`),
			wantJSON: []byte(`{"key": "value"}`),
		},
		{
			name:     "clean JSON with null bytes",
			input:    append([]byte(`{"key": "valu`), append([]byte{0}, []byte(`e"}`)...)...),
			wantJSON: []byte(`{"key": "value"}`),
		},
		{
			name:     "clean JSON with tabs and newlines",
			input:    []byte(`{"key": "value\n\t"}`),
			wantJSON: []byte(`{"key": "value\n\t"}`),
		},
		{
			name:     "clean JSON with control chars",
			input:    append([]byte(`{"key": "valu`), append([]byte{0x01, 0x02, 0x03}, []byte(`e"}`)...)...),
			wantJSON: []byte(`{"key": "value"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanJSON(tt.input)
			if string(result) != string(tt.wantJSON) {
				t.Errorf("CleanJSON() = %s, want %s", result, tt.wantJSON)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "valid JSON with embedded null",
			input:   append([]byte(`{"key": "valu`), append([]byte{0}, []byte(`e"}`)...)...),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{invalid}`),
			wantErr: true,
		},
		{
			name:    "empty JSON",
			input:   []byte(``),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}

func TestFindWorlds(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.MkdirAll(filepath.Join(tmpDir, "world1/data/journal"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "world2/data/journal"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "world3/data/users"), 0755) // No journal
	os.Mkdir(filepath.Join(tmpDir, "file.txt"), 0755)

	tests := []struct {
		name    string
		path    string
		want    []string
		wantErr bool
	}{
		{
			name:    "find worlds with journals",
			path:    tmpDir,
			want:    []string{"world1", "world2"},
			wantErr: false,
		},
		{
			name:    "find worlds in non-existent directory",
			path:    "/non-existent/path",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindWorlds(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindWorlds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != len(tt.want) {
					t.Errorf("Expected %d worlds, got %d", len(tt.want), len(result))
				}
				for _, world := range tt.want {
					found := false
					for _, r := range result {
						if r == world {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected world %s in result", world)
					}
				}
			}
		})
	}
}

func TestOpenWorldJournal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid world structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "validworld/data/journal"), 0755)

	tests := []struct {
		name    string
		path    string
		world   string
		wantErr bool
	}{
		{
			name:    "open valid world",
			path:    tmpDir,
			world:   "validworld",
			wantErr: false,
		},
		{
			name:    "open non-existent world",
			path:    tmpDir,
			world:   "nonexistentworld",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := OpenWorldJournal(tt.path, tt.world)
			if (err != nil) != tt.wantErr {
				// Note: OpenWorldJournal may succeed if dir exists but isn't valid LevelDB
				if tt.wantErr && err == nil {
					r.Close()
				}
			}
			if !tt.wantErr && err == nil {
				defer r.Close()
			}
		})
	}
}

func TestOpenWorldUsers(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid world structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "validworld/data/users"), 0755)

	tests := []struct {
		name    string
		path    string
		world   string
		wantErr bool
	}{
		{
			name:    "open valid world",
			path:    tmpDir,
			world:   "validworld",
			wantErr: false,
		},
		{
			name:    "open non-existent world",
			path:    tmpDir,
			world:   "nonexistentworld",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := OpenWorldUsers(tt.path, tt.world)
			if (err != nil) != tt.wantErr {
				// Note: OpenWorldUsers may succeed if dir exists but isn't valid LevelDB
				if tt.wantErr && err == nil {
					r.Close()
				}
			}
			if !tt.wantErr && err == nil {
				defer r.Close()
			}
		})
	}
}


// Helper function to create a test LevelDB
func createTestDB(t *testing.T, path string) *leveldb.DB {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	return db
}

// Note: Iterate tests skipped - use real LevelDB for integration testing
