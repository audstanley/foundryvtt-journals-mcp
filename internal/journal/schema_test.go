package journal

import (
	"testing"
)

func TestCheckPermission(t *testing.T) {
	userID := "testuser123"

	tests := []struct {
		name      string
		ownership map[string]int
		userID    string
		wantLevel PermissionLevel
	}{
		{
			name:      "user-specific permission",
			ownership: map[string]int{userID: ToInt(Editor)},
			userID:    userID,
			wantLevel: Editor,
		},
		{
			name:      "default permission fallback",
			ownership: map[string]int{"default": ToInt(Observer)},
			userID:    userID,
			wantLevel: Observer,
		},
		{
			name:      "nil ownership defaults to owner",
			ownership: nil,
			userID:    userID,
			wantLevel: Owner,
		},
		{
			name:      "empty ownership defaults to owner",
			ownership: map[string]int{},
			userID:    userID,
			wantLevel: Owner,
		},
		{
			name:      "user has higher permission than default",
			ownership: map[string]int{userID: ToInt(Editor), "default": ToInt(Observer)},
			userID:    userID,
			wantLevel: Editor,
		},
		{
			name:      "user has no access but default has editor",
			ownership: map[string]int{userID: ToInt(NoAccess), "default": ToInt(Editor)},
			userID:    userID,
			wantLevel: NoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPermission(tt.ownership, tt.userID)
			if result != tt.wantLevel {
				t.Errorf("CheckPermission() = %v, want %v", result, tt.wantLevel)
			}
		})
	}
}

func TestHasAccess(t *testing.T) {
	userID := "testuser123"

	tests := []struct {
		name      string
		ownership map[string]int
		userID    string
		want      bool
	}{
		{
			name:      "has observer access",
			ownership: map[string]int{"default": ToInt(Observer)},
			userID:    userID,
			want:      true,
		},
		{
			name:      "has editor access",
			ownership: map[string]int{"default": ToInt(Editor)},
			userID:    userID,
			want:      true,
		},
		{
			name:      "no access",
			ownership: map[string]int{"default": ToInt(NoAccess)},
			userID:    userID,
			want:      false,
		},
		{
			name:      "nil ownership has access",
			ownership: nil,
			userID:    userID,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasAccess(tt.ownership, tt.userID)
			if result != tt.want {
				t.Errorf("HasAccess() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestPageTypeStats_Add(t *testing.T) {
	stats := &PageTypeStats{}

	tests := []struct {
		name string
		page JournalPage
	}{
		{
			name: "add text page",
			page: JournalPage{Type: "text"},
		},
		{
			name: "add image page",
			page: JournalPage{Type: "image"},
		},
		{
			name: "add video page",
			page: JournalPage{Type: "video"},
		},
		{
			name: "add pdf page",
			page: JournalPage{Type: "pdf"},
		},
		{
			name: "add other page",
			page: JournalPage{Type: "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats.Add(tt.page)
		})
	}

	// Verify counts
	if stats.Total != 5 {
		t.Errorf("Total = %d, want 5", stats.Total)
	}
	if stats.Text != 1 {
		t.Errorf("Text = %d, want 1", stats.Text)
	}
	if stats.Image != 1 {
		t.Errorf("Image = %d, want 1", stats.Image)
	}
	if stats.Video != 1 {
		t.Errorf("Video = %d, want 1", stats.Video)
	}
	if stats.PDF != 1 {
		t.Errorf("PDF = %d, want 1", stats.PDF)
	}
	if stats.Other != 1 {
		t.Errorf("Other = %d, want 1", stats.Other)
	}
}

func TestPageTypeStats_AddMultiple(t *testing.T) {
	stats := &PageTypeStats{}

	// Add multiple pages of same type
	for i := 0; i < 3; i++ {
		stats.Add(JournalPage{Type: "text"})
	}
	stats.Add(JournalPage{Type: "image"})

	if stats.Text != 3 {
		t.Errorf("Text = %d, want 3", stats.Text)
	}
	if stats.Total != 4 {
		t.Errorf("Total = %d, want 4", stats.Total)
	}
}

func TestPageTypeStats_Zero(t *testing.T) {
	stats := &PageTypeStats{}

	if stats.Total != 0 {
		t.Errorf("Zero Total = %d, want 0", stats.Total)
	}
	if stats.Text != 0 {
		t.Errorf("Zero Text = %d, want 0", stats.Text)
	}
}
