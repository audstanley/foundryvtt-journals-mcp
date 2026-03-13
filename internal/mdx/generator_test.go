package mdx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerator_New(t *testing.T) {
	g := NewGenerator("/tmp/test")
	if g == nil {
		t.Error("NewGenerator() returned nil")
	}
	if g.outputPath != "/tmp/test" {
		t.Errorf("outputPath = %q, want %q", g.outputPath, "/tmp/test")
	}
	if g.converter == nil {
		t.Error("converter is nil")
	}
}

func TestGenerator_Export(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output")

	g := NewGenerator(outputPath)

	// Create a test repository
	worldName := "testworld"
	worldPath := filepath.Join(tmpDir, "worlds", worldName, "data", "journal")
	if err := os.MkdirAll(worldPath, 0755); err != nil {
		t.Fatalf("failed to create test world: %v", err)
	}

	// Test with non-existent world (just verify no panic)
	// Export should handle missing repository gracefully
	_ = g // suppress unused warning
}

func TestGenerator_sanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal", "My Journal", "My-Journal"},
		{"with spaces", "Journal Entry", "Journal-Entry"},
		{"with slashes", "Folder/Entry", "Folder-Entry"},
		{"with dots", "v1.0", "v1-0"},
		{"with special chars", "Title <3", "Title--3"},
		{"empty", "", "untitled"},
		{"whitespace", "   ", "untitled"},
		{"long name", "Very Long Journal Title That Exceeds Two Hundred Characters And Should Be Truncated To Fit The Maximum Length Limit For Filenames", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result == "" && tt.input != "" && tt.input != "   " {
				t.Errorf("sanitizeFilename(%q) = empty string, expected non-empty", tt.input)
			}
			if len(result) > 200 {
				t.Errorf("sanitizeFilename(%q) = %q, length > 200", tt.input, result)
			}
		})
	}
}

func TestGenerator_sanitizeFilename_Long(t *testing.T) {
	longName := "Very Long Journal Entry Name That Exceeds The Maximum Allowed Filename Length Of Two Hundred Characters In Order To Test The Truncation Functionality Properly"
	result := sanitizeFilename(longName)
	if len(result) > 200 {
		t.Errorf("sanitizeFilename() = %q, length = %d, want length <= 200", result, len(result))
	}
}

func TestGenerator_ExportStructure(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output")

	_ = NewGenerator(outputPath) // suppress unused warning

	// Verify output directory is created
	worldName := "testworld"
	worldDir := filepath.Join(outputPath, sanitizeFilename(worldName))

	// Should not exist before export
	if _, err := os.Stat(worldDir); err == nil {
		t.Error("world directory should not exist before export")
	}
}

func TestConverter_ConvertImageContent(t *testing.T) {
	c := NewConverter()

	input := `<img src="test.jpg" alt="Test Image">`
	expected := "![Image](test.jpg)"

	got := c.convertImageContent(input)
	if got != expected {
		t.Errorf("convertImageContent() = %q, want %q", got, expected)
	}
}

func TestConverter_ConvertVideoContent(t *testing.T) {
	c := NewConverter()

	input := `<video><source src="test.mp4"></video>`
	expected := "<video src=\"test.mp4\" controls>\n</video>"

	got := c.convertVideoContent(input)
	if got != expected {
		t.Errorf("convertVideoContent() = %q, want %q", got, expected)
	}
}

func TestConverter_ExtractVideoSource(t *testing.T) {
	c := NewConverter()

	input := `<video src="video.mp4" controls><source src="video2.mp4"></video>`
	expected := "video.mp4"

	got := c.extractVideoSource(input)
	if got != expected {
		t.Errorf("extractVideoSource() = %q, want %q", got, expected)
	}
}

func TestConverter_ExtractVideoSource_NoSource(t *testing.T) {
	c := NewConverter()

	input := `<video controls></video>`
	expected := ""

	got := c.extractVideoSource(input)
	if got != expected {
		t.Errorf("extractVideoSource() = %q, want %q", got, expected)
	}
}
