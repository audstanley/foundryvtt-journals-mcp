package mdx

import (
	"testing"
)

func TestTreeWalker_Paragraphs(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name:   "simple paragraph",
			html:   "<p>Hello world</p>",
			expect: "Hello world",
		},
		{
			name:   "multiple paragraphs",
			html:   "<p>First</p><p>Second</p>",
			expect: "First\n\nSecond",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAndConvert(tt.html)
			if result != tt.expect {
				t.Errorf("ParseAndConvert() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestTreeWalker_Headings(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name:   "h1 heading",
			html:   "<h1>Title</h1>",
			expect: "# Title",
		},
		{
			name:   "h2 heading",
			html:   "<h2>Subtitle</h2>",
			expect: "## Subtitle",
		},
		{
			name:   "h3 heading",
			html:   "<h3>Section</h3>",
			expect: "### Section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAndConvert(tt.html)
			if result != tt.expect {
				t.Errorf("ParseAndConvert() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestTreeWalker_InlineFormatting(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name:   "bold with strong",
			html:   "<p><strong>bold text</strong></p>",
			expect: "**bold text**",
		},
		{
			name:   "bold with b",
			html:   "<p><b>bold text</b></p>",
			expect: "**bold text**",
		},
		{
			name:   "italic with em",
			html:   "<p><em>italic text</em></p>",
			expect: "*italic text*",
		},
		{
			name:   "italic with i",
			html:   "<p><i>italic text</i></p>",
			expect: "*italic text*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAndConvert(tt.html)
			if result != tt.expect {
				t.Errorf("ParseAndConvert() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestTreeWalker_Lists(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name:   "unordered list",
			html:   "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expect: "- Item 1\n- Item 2",
		},
		{
			name:   "ordered list",
			html:   "<ol><li>First</li><li>Second</li></ol>",
			expect: "1. First\n2. Second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAndConvert(tt.html)
			if result != tt.expect {
				t.Errorf("ParseAndConvert() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestTreeWalker_Tables(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name: "simple table",
			html: `<table>
				<tr><th>Name</th><th>Value</th></tr>
				<tr><td>Item 1</td><td>100</td></tr>
			</table>`,
			expect: `| Name | Value |
| --- | --- |
| Item 1 | 100 |`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAndConvert(tt.html)
			if result != tt.expect {
				t.Errorf("ParseAndConvert() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestTreeWalker_UUIDLinks(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name:   "single uuid link",
			html:   "@UUID[Item.PVfwaiCVTWLqBDYK]{Apple}",
			expect: "[Apple](uuid://Item/PVfwaiCVTWLqBDYK)",
		},
		{
			name:   "actor uuid link",
			html:   "@UUID[Actor.gfpVDbgCS5cWqrnr.Item.elObA1p704m92iIo]{The Unblinking}",
			expect: "[The Unblinking](uuid://Actor/gfpVDbgCS5cWqrnr.Item.elObA1p704m92iIo)",
		},
		{
			name:   "multiple uuid links",
			html:   "@UUID[Item.A]{One} and @UUID[Item.B]{Two}",
			expect: "[One](uuid://Item/A) and [Two](uuid://Item/B)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAndConvertWithUUIDLinks(tt.html)
			if result != tt.expect {
				t.Errorf("ParseAndConvertWithUUIDLinks() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestTreeWalker_ComplexContent(t *testing.T) {
	html := `<h1>Journal Entry</h1>
		<p>This is the <strong>main content</strong> with <em>formatting</em>.</p>
		<h2>Subsection</h2>
		<ul>
			<li>First item</li>
			<li>Second item</li>
		</ul>
		<p>Final with @UUID[Item.ABC123]{item link}.</p>`

	result := ParseAndConvertWithUUIDLinks(html)

	// Verify key elements
	checks := []string{
		"# Journal Entry",
		"## Subsection",
		"**main content**",
		"*formatting*",
		"- First item",
		"- Second item",
		"[item link](uuid://Item/ABC123)",
	}

	for _, check := range checks {
		if !contains(result, check) {
			t.Errorf("Result missing expected element: %s", check)
		}
	}
}
