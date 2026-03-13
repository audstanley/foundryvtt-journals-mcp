package mdx

import (
	"fmt"
	"regexp"
	"strings"
)

// Converter handles HTML to Markdown conversion
type Converter struct {
	// Precompiled regex patterns
	patterns map[string]*regexp.Regexp
}

// NewConverter creates a new HTML to Markdown converter
func NewConverter() *Converter {
	c := &Converter{
		patterns: make(map[string]*regexp.Regexp),
	}
	c.compilePatterns()
	return c
}

// compilePatterns initializes all regex patterns
func (c *Converter) compilePatterns() {
	// Block elements
	c.patterns["p"] = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	c.patterns["br"] = regexp.MustCompile(`(?is)<br\s*/?>`)
	c.patterns["hr"] = regexp.MustCompile(`(?is)<hr\s*/?>`)
	c.patterns["blockquote"] = regexp.MustCompile(`(?is)<blockquote[^>]*>(.*?)</blockquote>`)

	// Headings
	c.patterns["h1"] = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	c.patterns["h2"] = regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`)
	c.patterns["h3"] = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	c.patterns["h4"] = regexp.MustCompile(`(?is)<h4[^>]*>(.*?)</h4>`)
	c.patterns["h5"] = regexp.MustCompile(`(?is)<h5[^>]*>(.*?)</h5>`)
	c.patterns["h6"] = regexp.MustCompile(`(?is)<h6[^>]*>(.*?)</h6>`)

	// Lists
	c.patterns["ul"] = regexp.MustCompile(`(?is)<ul[^>]*>(.*?)</ul>`)
	c.patterns["ol"] = regexp.MustCompile(`(?is)<ol[^>]*>(.*?)</ol>`)
	c.patterns["li"] = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)

	// Tables
	c.patterns["table"] = regexp.MustCompile(`(?is)<table[^>]*>(.*?)</table>`)
	c.patterns["thead"] = regexp.MustCompile(`(?is)<thead[^>]*>(.*?)</thead>`)
	c.patterns["tbody"] = regexp.MustCompile(`(?is)<tbody[^>]*>(.*?)</tbody>`)
	c.patterns["tr"] = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	c.patterns["th"] = regexp.MustCompile(`(?is)<th[^>]*>(.*?)</th>`)
	c.patterns["td"] = regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)

	// Inline elements
	c.patterns["strong"] = regexp.MustCompile(`(?is)<strong[^>]*>(.*?)</strong>|(?is)<b[^>]*>(.*?)</b>`)
	c.patterns["em"] = regexp.MustCompile(`(?is)<em[^>]*>(.*?)</em>|(?is)<i[^>]*>(.*?)</i>`)
	c.patterns["code"] = regexp.MustCompile(`(?is)<code[^>]*>(.*?)</code>`)
	c.patterns["pre"] = regexp.MustCompile(`(?is)<pre[^>]*>(.*?)</pre>`)
	c.patterns["a"] = regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`)
	c.patterns["img"] = regexp.MustCompile(`(?is)<img[^>]*src=["']([^"']*)["'][^>]*alt=["']?([^"']*)["']?\s*/?>`)
	c.patterns["video"] = regexp.MustCompile(`(?is)<video[^>]*>(.*?)</video>`)

	// Clean up tags
	c.patterns["empty"] = regexp.MustCompile(`(?is)</?(?:p|div|h[1-6]|ul|ol|li|table|thead|tbody|tr|th|td|blockquote|br|hr)\s*/?>`)
	c.patterns["whitespace"] = regexp.MustCompile(`(?s)\n\s*\n`)
}

// Convert transforms HTML content to Markdown
func (c *Converter) Convert(html string) string {
	// Handle empty input
	if strings.TrimSpace(html) == "" {
		return ""
	}

	// Try tree-based parser first (more robust)
	result := ParseAndConvert(html)

	// If tree parser produced empty result, fall back to regex
	if result == "" {
		result = c.convertWithRegex(html)
	}

	return result
}

// convertWithRegex is the legacy regex-based converter (kept for fallback)
func (c *Converter) convertWithRegex(html string) string {
	result := html

	// Process tables first (they have complex structure)
	result = c.convertTables(result)

	// Process blockquotes
	result = c.patterns["blockquote"].ReplaceAllStringFunc(result, func(match string) string {
		content := c.patterns["blockquote"].FindStringSubmatch(match)[1]
		content = c.stripTags(content)
		// Convert newlines to proper markdown blockquote format
		lines := strings.Split(strings.TrimSpace(content), "\n")
		var quotedLines []string
		for _, line := range lines {
			quotedLines = append(quotedLines, "> "+strings.TrimSpace(line))
		}
		return "\n" + strings.Join(quotedLines, "\n") + "\n"
	})

	// Process headings (h1-h6)
	for i := 1; i <= 6; i++ {
		pattern := c.patterns["h"+string(rune('0'+i))]
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			content := pattern.FindStringSubmatch(match)[1]
			content = c.stripInlineTags(content)
			content = strings.TrimSpace(content)
			return "\n" + strings.Repeat("#", i) + " " + content + "\n\n"
		})
	}

	// Process inline formatting first (before paragraphs strip them)
	// Process inline code
	result = c.patterns["code"].ReplaceAllStringFunc(result, func(match string) string {
		content := c.patterns["code"].FindStringSubmatch(match)[1]
		content = c.stripTags(content)
		return "`" + content + "`"
	})

	// Process bold
	result = c.patterns["strong"].ReplaceAllStringFunc(result, func(match string) string {
		var content string
		if c.patterns["strong"].FindStringSubmatch(match)[1] != "" {
			content = c.patterns["strong"].FindStringSubmatch(match)[1]
		} else {
			content = c.patterns["strong"].FindStringSubmatch(match)[2]
		}
		content = c.stripInlineTags(content)
		return "**" + content + "**"
	})

	// Process italic
	result = c.patterns["em"].ReplaceAllStringFunc(result, func(match string) string {
		var content string
		if c.patterns["em"].FindStringSubmatch(match)[1] != "" {
			content = c.patterns["em"].FindStringSubmatch(match)[1]
		} else {
			content = c.patterns["em"].FindStringSubmatch(match)[2]
		}
		content = c.stripInlineTags(content)
		return "*" + content + "*"
	})

	// Process links
	result = c.patterns["a"].ReplaceAllStringFunc(result, func(match string) string {
		parts := c.patterns["a"].FindStringSubmatch(match)
		href := parts[1]
		text := parts[2]
		text = c.stripInlineTags(text)
		return "[" + text + "](" + href + ")"
	})

	// Process paragraphs
	result = c.patterns["p"].ReplaceAllStringFunc(result, func(match string) string {
		content := c.patterns["p"].FindStringSubmatch(match)[1]
		content = strings.TrimSpace(content)
		if content != "" {
			return "\n" + content + "\n\n"
		}
		return ""
	})

	// Process lists
	result = c.convertLists(result)

	// Process code blocks
	result = c.patterns["pre"].ReplaceAllStringFunc(result, func(match string) string {
		content := c.patterns["pre"].FindStringSubmatch(match)[1]
		content = c.stripTags(content)
		content = strings.TrimSpace(content)
		return "\n```\n" + content + "\n```\n\n"
	})

	// Process images
	result = c.patterns["img"].ReplaceAllStringFunc(result, func(match string) string {
		parts := c.patterns["img"].FindStringSubmatch(match)
		src := parts[1]
		return "![Image](" + src + ")"
	})

	// Process video tags (simplified - just extract src)
	result = c.patterns["video"].ReplaceAllStringFunc(result, func(match string) string {
		content := c.patterns["video"].FindStringSubmatch(match)[1]
		// Try to extract video source
		videoSrc := c.extractVideoSource(content)
		if videoSrc != "" {
			return "\n<video src=\"" + videoSrc + "\" controls>\n</video>\n\n"
		}
		return ""
	})

	// Process line breaks
	result = c.patterns["br"].ReplaceAllString(result, "  \n")

	// Process horizontal rules
	result = c.patterns["hr"].ReplaceAllString(result, "\n---\n\n")

	// Clean up empty tags (div, p, etc)
	result = c.patterns["empty"].ReplaceAllString(result, "")

	// Clean up extra whitespace
	result = c.patterns["whitespace"].ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}

// convertTables transforms HTML tables to Markdown table format
func (c *Converter) convertTables(html string) string {
	var result strings.Builder

	tableMatch := c.patterns["table"].FindStringSubmatch(html)
	if tableMatch == nil {
		return html
	}

	tableContent := tableMatch[1]

	// Extract rows
	trPattern := regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	rows := trPattern.FindAllString(tableContent, -1)

	if len(rows) == 0 {
		return html
	}

	// Process each row
	var headers []string
	var data [][]string

	for rowIndex, row := range rows {
		// Check if this is a header row (has th elements)
		thPattern := regexp.MustCompile(`(?is)<th[^>]*>(.*?)</th>`)
		tdPattern := regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)

		thMatches := thPattern.FindAllString(row, -1)
		tdMatches := tdPattern.FindAllString(row, -1)

		var cells []string
		if len(thMatches) > 0 {
			// Header row
			for _, th := range thMatches {
				content := thPattern.FindStringSubmatch(th)[1]
				content = c.stripInlineTags(content)
				cells = append(cells, strings.TrimSpace(content))
			}
			if rowIndex == 0 {
				headers = cells
			}
		} else if len(tdMatches) > 0 {
			// Data row
			for _, td := range tdMatches {
				content := tdPattern.FindStringSubmatch(td)[1]
				content = c.stripInlineTags(content)
				cells = append(cells, strings.TrimSpace(content))
			}
			data = append(data, cells)
		}
	}

	// Build Markdown table
	if len(headers) > 0 {
		// Header row
		result.WriteString("| ")
		result.WriteString(strings.Join(headers, " | "))
		result.WriteString(" |\n")

		// Separator
		result.WriteString("| ")
		for i := 0; i < len(headers); i++ {
			if i > 0 {
				result.WriteString(" | ")
			}
			result.WriteString("---")
		}
		result.WriteString(" |\n")

		// Data rows
		for _, row := range data {
			result.WriteString("| ")
			result.WriteString(strings.Join(row, " | "))
			result.WriteString(" |\n")
		}
	}

	return result.String()
}

// convertLists transforms HTML lists to Markdown list format
func (c *Converter) convertLists(html string) string {
	result := html

	// Process unordered lists
	ulPattern := regexp.MustCompile(`(?is)<ul[^>]*>(.*?)</ul>`)
	result = ulPattern.ReplaceAllStringFunc(result, func(ulMatch string) string {
		content := ulPattern.FindStringSubmatch(ulMatch)[1]

		// Extract list items
		liPattern := regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
		items := liPattern.FindAllString(content, -1)

		var itemsMarkdown []string
		for _, item := range items {
			itemContent := liPattern.FindStringSubmatch(item)[1]
			itemContent = c.stripInlineTags(itemContent)
			itemContent = strings.TrimSpace(itemContent)
			if itemContent != "" {
				itemsMarkdown = append(itemsMarkdown, "- "+itemContent)
			}
		}

		if len(itemsMarkdown) > 0 {
			return "\n" + strings.Join(itemsMarkdown, "\n") + "\n\n"
		}
		return ""
	})

	// Process ordered lists
	olPattern := regexp.MustCompile(`(?is)<ol[^>]*>(.*?)</ol>`)
	listNum := 1
	result = olPattern.ReplaceAllStringFunc(result, func(olMatch string) string {
		content := olPattern.FindStringSubmatch(olMatch)[1]

		// Extract list items
		liPattern := regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
		items := liPattern.FindAllString(content, -1)

		var itemsMarkdown []string
		for _, item := range items {
			itemContent := liPattern.FindStringSubmatch(item)[1]
			itemContent = c.stripInlineTags(itemContent)
			itemContent = strings.TrimSpace(itemContent)
			if itemContent != "" {
				itemsMarkdown = append(itemsMarkdown, fmt.Sprintf("%d. %s", listNum, itemContent))
				listNum++
			}
		}

		if len(itemsMarkdown) > 0 {
			return "\n" + strings.Join(itemsMarkdown, "\n") + "\n\n"
		}
		return ""
	})

	return result
}

// stripTags removes all HTML tags from content
func (c *Converter) stripTags(html string) string {
	tagPattern := regexp.MustCompile(`(?is)<[^>]*>`)
	return tagPattern.ReplaceAllString(html, "")
}

// stripInlineTags removes inline HTML tags but preserves content
func (c *Converter) stripInlineTags(html string) string {
	// Remove only formatting tags
	inlineTags := []string{"strong", "em", "code", "b", "i"}
	result := html

	for _, tag := range inlineTags {
		startPattern := regexp.MustCompile(fmt.Sprintf(`(?is)<%s[^>]*>`, tag))
		endPattern := regexp.MustCompile(fmt.Sprintf(`(?is)</%s>`, tag))
		result = startPattern.ReplaceAllString(result, "")
		result = endPattern.ReplaceAllString(result, "")
	}

	return result
}

// extractVideoSource extracts video source from video HTML
func (c *Converter) extractVideoSource(html string) string {
	srcPattern := regexp.MustCompile(`(?is)src=["']([^"']*)["']`)
	match := srcPattern.FindStringSubmatch(html)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// ConvertPage transforms a journal page's HTML content to Markdown
func (c *Converter) ConvertPage(pageType string, htmlContent string) string {
	switch pageType {
	case "image":
		return c.convertImageContent(htmlContent)
	case "video":
		return c.convertVideoContent(htmlContent)
	default:
		return c.Convert(htmlContent)
	}
}

// convertImageContent handles image pages
func (c *Converter) convertImageContent(html string) string {
	imgPattern := regexp.MustCompile(`(?is)<img[^>]*src=["']([^"']*)["'][^>]*>`)
	match := imgPattern.FindStringSubmatch(html)
	if len(match) > 1 {
		return "![Image](" + match[1] + ")"
	}
	return html
}

// convertVideoContent handles video pages
func (c *Converter) convertVideoContent(html string) string {
	srcPattern := regexp.MustCompile(`(?is)src=["']([^"']*)["']`)
	match := srcPattern.FindStringSubmatch(html)
	if len(match) > 1 {
		return "<video src=\"" + match[1] + "\" controls>\n</video>"
	}
	return html
}
