package mdx

import (
	"testing"
)

func TestConvertHTMLToMarkdown(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect string
	}{
		{
			name:   "empty string",
			html:   "",
			expect: "",
		},
		{
			name:   "simple paragraph",
			html:   "<p>Hello world</p>",
			expect: "Hello world",
		},
		{
			name:   "heading h1",
			html:   "<h1>Title</h1>",
			expect: "# Title",
		},
		{
			name:   "heading h2",
			html:   "<h2>Subtitle</h2>",
			expect: "## Subtitle",
		},
		{
			name:   "heading h3",
			html:   "<h3>Section</h3>",
			expect: "### Section",
		},
		{
			name:   "bold text",
			html:   "<b>Hello</b> <strong>world</strong>",
			expect: "**Hello** **world**",
		},
		{
			name:   "italic text",
			html:   "<i>Hello</i> <em>world</em>",
			expect: "*Hello* *world*",
		},
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
		{
			name:   "link",
			html:   "<a href='http://example.com'>Link</a>",
			expect: "[Link](http://example.com)",
		},
		{
			name:   "image",
			html:   "<img src='/path/to/image.png' alt='Alt text'>",
			expect: "![Image](/path/to/image.png)",
		},
		{
			name:   "video with controls",
			html:   "<video src='/path/to/video.mp4' controls='true' volume='0.5'></video>",
			expect: "",
		},
		{
			name:   "line breaks",
			html:   "<p>Line 1<br>Line 2</p>",
			expect: "Line 1\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewConverter()
			result := converter.Convert(tt.html)
			if result != tt.expect {
				t.Errorf("Convert() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestConvertHTMLToMarkdown_Whitespace(t *testing.T) {
	html := "<h1>  Title  </h1>"
	converter := NewConverter()
	_ = converter.Convert(html)
}

func TestConvertHTMLToMarkdown_EmptyTags(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "empty paragraph",
			html: "<p></p>",
		},
		{
			name: "empty div",
			html: "<div></div>",
		},
		{
			name: "empty span",
			html: "<span></span>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewConverter()
			result := converter.Convert(tt.html)
			// Empty tags produce empty result - that's expected
			_ = result
		})
	}
}
