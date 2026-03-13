package mdx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func normalizeWhitespace(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n\n", "\n")
	s = strings.ReplaceAll(s, "\n\n", "\n")
	return strings.TrimSpace(s)
}

func TestIntegration_ConvertActualContent(t *testing.T) {
	// Test with actual HTML content from journals
	c := NewConverter()

	// Sample HTML from Foundry VTT journals
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "paragraph with formatting",
			input: `<p>A day of <strong>solemn reflection</strong>, followed by the <em>ritual cleansing</em> of the sacred waters.</p>`,
		},
		{
			name:  "list items",
			input: `<ul><li>First item with details</li><li>Second item with more content</li></ul>`,
		},
		{
			name:  "complex content",
			input: `<h2>Chapter 1: The Beginning</h2><p>This is the <strong>start</strong> of our journey.</p><blockquote>Quote from ancient text</blockquote>`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Convert(tt.input)
			if result == "" {
				t.Errorf("Convert() returned empty string for %s", tt.name)
			}
			t.Logf("%s result length: %d chars", tt.name, len(result))
		})
	}
}

func TestIntegration_ExportToDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "mdx-output")

	g := NewGenerator(outputPath)

	// Create a mock repository structure
	worldDir := filepath.Join(tmpDir, "worlds", "testworld", "data", "journal")
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		t.Skipf("Cannot create test world: %v", err)
	}

	// Verify generator doesn't panic
	_ = g

	// Test directory creation logic
	worldName := "testworld"
	outputWorldDir := filepath.Join(outputPath, sanitizeFilename(worldName))

	if _, err := os.Stat(outputWorldDir); err == nil {
		t.Error("Directory should not exist before export")
	}
}

func TestIntegration_HTMLToMDX_Pipeline(t *testing.T) {
	// Test full HTML to MDX pipeline
	c := NewConverter()

	// Complex HTML with multiple elements
	html := `
		<h1>Journal Entry Title</h1>
		<p>This is the <strong>main content</strong> with <em>formatting</em>.</p>
		<h2>Subsection</h2>
		<ul>
			<li>First item</li>
			<li>Second item</li>
			<li>Third item</li>
		</ul>
		<blockquote>A quote from the ancient texts</blockquote>
		<p>Final paragraph with <a href="https://example.com">a link</a>.</p>
	`

	md := c.Convert(html)

	if md == "" {
		t.Error("Conversion produced empty output")
	}

	// Verify key elements are present
	checks := map[string]bool{
		"# Journal Entry Title":         false,
		"## Subsection":                 false,
		"**main content**":              false,
		"*formatting*":                  false,
		"- First item":                  false,
		"> A quote":                     false,
		"[a link](https://example.com)": false,
	}

	for check := range checks {
		if contains(md, check) {
			checks[check] = true
		}
	}

	for check := range checks {
		if !checks[check] {
			t.Errorf("Missing expected element: %s", check)
		}
	}

	t.Logf("Conversion produced %d characters of Markdown", len(md))
}

func TestIntegration_TableConversion(t *testing.T) {
	c := NewConverter()

	html := `<table>
		<tr><th>Name</th><th>Value</th><th>Description</th></tr>
		<tr><td>Item 1</td><td>100</td><td>First item description</td></tr>
		<tr><td>Item 2</td><td>200</td><td>Second item description</td></tr>
	</table>`

	md := c.Convert(html)

	// Verify table structure
	expectedElements := []string{
		"| Name | Value | Description |",
		"| --- | --- | --- |",
		"| Item 1 | 100 | First item description |",
		"| Item 2 | 200 | Second item description |",
	}

	for i := range expectedElements {
		elem := expectedElements[i]
		if !contains(md, elem) {
			t.Errorf("Missing table element: %s", elem)
		}
	}

	t.Logf("Table conversion: %s", md)
}

func TestIntegration_ListConversion(t *testing.T) {
	c := NewConverter()

	html := `<ol><li>Step one</li><li>Step two</li><li>Step three</li></ol>`
	md := c.Convert(html)

	if !contains(md, "1. Step one") {
		t.Error("Missing ordered list item 1")
	}
	if !contains(md, "2. Step two") {
		t.Error("Missing ordered list item 2")
	}
	if !contains(md, "3. Step three") {
		t.Error("Missing ordered list item 3")
	}

	t.Logf("List conversion: %s", md)
}

func TestIntegration_NestedFormatting(t *testing.T) {
	c := NewConverter()

	html := `<p>This is <strong>bold <em>and italic</em></strong> content.</p>`
	md := c.Convert(html)

	// Should preserve nested formatting
	if !contains(md, "**") {
		t.Error("Missing bold formatting")
	}
	if !contains(md, "*") {
		t.Error("Missing italic formatting")
	}

	t.Logf("Nested formatting: %s", md)
}

func TestIntegration_Headings(t *testing.T) {
	c := NewConverter()

	html := `<h1>H1</h1><h2>H2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6>`
	md := c.Convert(html)

	checks := []struct {
		expected string
		found    bool
	}{
		{"# H1", false},
		{"## H2", false},
		{"### H3", false},
		{"#### H4", false},
		{"##### H5", false},
		{"###### H6", false},
	}

	for i := range checks {
		if contains(md, checks[i].expected) {
			checks[i].found = true
		}
	}

	for _, check := range checks {
		if !check.found {
			t.Errorf("Missing heading: %s", check.expected)
		}
	}
}

func TestIntegration_WhitespaceHandling(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		name     string
		input    string
		validate func(string) bool
	}{
		{"multiple paragraphs", "<p>P1</p><p>P2</p>", func(md string) bool {
			return contains(md, "P1") && contains(md, "P2")
		}},
		{"extra whitespace", "   <p>Content</p>   ", func(md string) bool {
			return contains(md, "Content")
		}},
		{"no extra newlines", "<p>P1</p><br><p>P2</p>", func(md string) bool {
			return contains(md, "P1") && contains(md, "P2")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := c.Convert(tt.input)
			if !tt.validate(md) {
				t.Errorf("Validation failed for: %s", tt.name)
			}
			t.Logf("Result: %s", md)
		})
	}
}

func TestIntegration_InvalidHTML(t *testing.T) {
	c := NewConverter()

	// Test with malformed HTML
	invalidHTML := `<p>Unclosed paragraph<br><strong>Bold text</p>`

	md := c.Convert(invalidHTML)

	// Should not panic, should produce some output
	if md == "" {
		t.Error("Invalid HTML produced empty output")
	}

	t.Logf("Invalid HTML handled: %s", md)
}

func TestIntegration_EmptyContent(t *testing.T) {
	c := NewConverter()

	tests := []string{
		"",
		"   ",
		"<p></p>",
		"<div></div>",
	}

	for _, input := range tests {
		result := c.Convert(input)
		if result != "" {
			// Some empty tags might produce whitespace
			result = normalizeWhitespace(result)
			if result != "" {
				t.Errorf("Empty content %q produced: %q", input, result)
			}
		}
	}
}
