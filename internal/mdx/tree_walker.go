package mdx

import (
	"strings"

	"golang.org/x/net/html"
)

// ParseAndConvert parses HTML using tree-based approach and converts to Markdown
func ParseAndConvert(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fall back to existing regex converter if parsing fails
		c := NewConverter()
		return c.Convert(htmlContent)
	}

	// Find body node or use document children
	var startNode *html.Node
	if doc != nil {
		for c := doc.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				tag := strings.ToLower(c.Data)
				if tag == "body" {
					startNode = c
					break
				} else if tag == "html" {
					// Look for body inside html
					for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
						if cc.Type == html.ElementNode && strings.ToLower(cc.Data) == "body" {
							startNode = cc
							break
						}
					}
					if startNode == nil {
						startNode = c
					}
					break
				}
			}
		}
	}

	if startNode == nil {
		startNode = doc
	}

	// Simple tree walker for text extraction
	walker := &treeWalker{}
	walker.Walk(startNode)

	result := walker.String()

	// Clean up formatting
	result = strings.TrimSpace(result)
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")

	return result
}

// treeWalker is a simple HTML tree walker for Markdown conversion
type treeWalker struct {
	builder       *strings.Builder
	listDepth     int
	listCounters  []int
	isOrderedList []bool
}

func (w *treeWalker) Walk(n *html.Node) {
	if w.builder == nil {
		w.builder = &strings.Builder{}
	}
	for n != nil {
		w.processNode(n)
		n = n.NextSibling
	}
}

func (w *treeWalker) processNode(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		text := n.Data
		// Preserve at least single spaces to avoid merging inline elements
		if text != "" {
			w.builder.WriteString(text)
		}
	case html.ElementNode:
		tag := strings.ToLower(n.Data)
		w.processElement(tag, n)
	}
}

func (w *treeWalker) processElement(tag string, n *html.Node) {
	switch tag {
	case "p":
		w.Walk(n.FirstChild)
		w.builder.WriteString("\n\n")
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(tag[1] - '0')
		w.builder.WriteString("\n" + strings.Repeat("#", level) + " ")
		w.Walk(n.FirstChild)
		w.builder.WriteString("\n\n")
	case "ul":
		w.listDepth++
		w.listCounters = append(w.listCounters, 0)
		w.isOrderedList = append(w.isOrderedList, false)
		w.Walk(n.FirstChild)
		w.listDepth--
		w.listCounters = w.listCounters[:len(w.listCounters)-1]
		w.isOrderedList = w.isOrderedList[:len(w.isOrderedList)-1]
	case "ol":
		w.listDepth++
		w.listCounters = append(w.listCounters, 1)
		w.isOrderedList = append(w.isOrderedList, true)
		w.Walk(n.FirstChild)
		w.listDepth--
		w.listCounters = w.listCounters[:len(w.listCounters)-1]
		w.isOrderedList = w.isOrderedList[:len(w.isOrderedList)-1]
	case "li":
		if len(w.listCounters) > 0 {
			counter := w.listCounters[len(w.listCounters)-1]
			indent := strings.Repeat("  ", w.listDepth-1)
			prefix := "- "
			if len(w.isOrderedList) > 0 && w.isOrderedList[len(w.isOrderedList)-1] {
				prefix = string(rune('0'+counter)) + ". "
				w.listCounters[len(w.listCounters)-1]++
			}
			w.builder.WriteString(indent + prefix)
			w.Walk(n.FirstChild)
			w.builder.WriteString("\n")
		} else {
			w.Walk(n.FirstChild)
		}
	case "strong", "b":
		w.builder.WriteString("**")
		w.Walk(n.FirstChild)
		w.builder.WriteString("**")
	case "em", "i":
		w.builder.WriteString("*")
		w.Walk(n.FirstChild)
		w.builder.WriteString("*")
	case "a":
		href := w.getAttribute(n, "href")
		// Collect link text first
		linkBuilder := &strings.Builder{}
		tempWalker := &treeWalker{builder: linkBuilder}
		tempWalker.Walk(n.FirstChild)
		linkText := strings.TrimSpace(linkBuilder.String())
		if href != "" && linkText != "" {
			w.builder.WriteString("[" + linkText + "](" + href + ")")
		} else if linkText != "" {
			w.builder.WriteString(linkText)
		}
	case "img":
		src := w.getAttribute(n, "src")
		if src != "" {
			alt := w.getAttribute(n, "alt")
			if alt == "" {
				alt = "Image"
			}
			w.builder.WriteString("![Image](" + src + ")")
		}
	case "br":
		w.builder.WriteString("\n")
	case "hr":
		w.builder.WriteString("\n---\n\n")
	case "blockquote":
		w.builder.WriteString("> ")
		w.Walk(n.FirstChild)
		w.builder.WriteString("\n\n")
	case "code":
		w.builder.WriteString("`")
		w.Walk(n.FirstChild)
		w.builder.WriteString("`")
	case "pre":
		w.builder.WriteString("\n```\n")
		w.Walk(n.FirstChild)
		w.builder.WriteString("\n```\n\n")
	case "table":
		w.buildTable(n)
	case "tr":
		// Handled by table
	case "th", "td":
		// Handled by table
	case "thead", "tbody":
		w.Walk(n.FirstChild)
	case "div", "span":
		w.Walk(n.FirstChild)
	default:
		w.Walk(n.FirstChild)
	}
}

func (w *treeWalker) getAttribute(n *html.Node, attr string) string {
	for _, a := range n.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

func (w *treeWalker) buildTable(n *html.Node) {
	var headers []string
	var rows [][]string
	var currentRow []*html.Node
	var isFirstRow = true

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		for node != nil {
			if node.Type == html.ElementNode {
				tag := strings.ToLower(node.Data)
				switch tag {
				case "tr":
					if len(currentRow) > 0 {
						cells := w.extractCells(currentRow)
						if isFirstRow {
							headers = cells
							isFirstRow = false
						} else {
							rows = append(rows, cells)
						}
						currentRow = nil
					}
					for cell := node.FirstChild; cell != nil; cell = cell.NextSibling {
						if cell.Type == html.ElementNode {
							currentRow = append(currentRow, cell)
						}
					}
				case "thead", "tbody", "table":
					walk(node.FirstChild)
					// Don't continue to NextSibling - we've handled the children
					return
				}
			}
			node = node.NextSibling
		}
		if len(currentRow) > 0 {
			cells := w.extractCells(currentRow)
			if isFirstRow {
				headers = cells
				isFirstRow = false
			} else {
				rows = append(rows, cells)
			}
		}
	}

	walk(n.FirstChild)

	if len(headers) > 0 {
		tableBuilder := &strings.Builder{}
		tableBuilder.WriteString("| " + strings.Join(headers, " | ") + " |\n")
		separator := make([]string, len(headers))
		for i := range separator {
			separator[i] = "---"
		}
		tableBuilder.WriteString("| " + strings.Join(separator, " | ") + " |\n")
		for _, row := range rows {
			tableBuilder.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
		w.builder.WriteString(tableBuilder.String())
	}
}

func (w *treeWalker) extractCells(cells []*html.Node) []string {
	var result []string
	for _, cell := range cells {
		cellBuilder := &strings.Builder{}
		tempWalker := &treeWalker{builder: cellBuilder}
		tempWalker.Walk(cell.FirstChild)
		content := strings.TrimSpace(cellBuilder.String())
		result = append(result, content)
	}
	return result
}

func (w *treeWalker) String() string {
	return w.builder.String()
}

// ParseAndConvertWithUUIDLinks converts HTML and handles @UUID{} links
func ParseAndConvertWithUUIDLinks(htmlContent string) string {
	result := ParseAndConvert(htmlContent)
	result = extractUUIDLinks(result)
	return result
}

// extractUUIDLinks finds and converts @UUID{} patterns to markdown links
func extractUUIDLinks(content string) string {
	start := 0
	var result strings.Builder

	for {
		uuidStart := strings.Index(content[start:], "@UUID[")
		if uuidStart == -1 {
			result.WriteString(content[start:])
			break
		}

		result.WriteString(content[start : start+uuidStart])

		searchStart := start + uuidStart + 6
		if searchStart >= len(content) {
			result.WriteString("@UUID[")
			start = start + uuidStart + 1
			continue
		}

		bracketPos := strings.Index(content[searchStart:], "]{")
		if bracketPos == -1 {
			result.WriteString("@UUID[")
			start = start + uuidStart + 1
			continue
		}

		textStart := searchStart + bracketPos + 2
		if textStart >= len(content) {
			result.WriteString("@UUID[")
			start = start + uuidStart + 1
			continue
		}

		closeBracePos := strings.Index(content[textStart:], "}")
		if closeBracePos == -1 {
			result.WriteString("@UUID[")
			start = start + uuidStart + 1
			continue
		}

		middle := content[searchStart : searchStart+bracketPos]
		text := content[textStart : textStart+closeBracePos]

		parts := strings.SplitN(middle, ".", 2)
		if len(parts) == 2 {
			uuid := "uuid://" + parts[0] + "/" + parts[1]
			result.WriteString("[" + text + "](" + uuid + ")")
		} else {
			result.WriteString("@UUID[" + middle + "]{text}")
		}

		start = textStart + closeBracePos + 1
	}

	return result.String()
}

// ExtractUUIDLinks extracts UUID references for MCP tool resolution
func ExtractUUIDLinks(content string) []UUIDReference {
	var refs []UUIDReference
	start := 0

	for {
		uuidStart := strings.Index(content[start:], "@UUID[")
		if uuidStart == -1 {
			break
		}

		searchStart := start + uuidStart + 6
		if searchStart >= len(content) {
			start = start + uuidStart + 1
			continue
		}

		bracketPos := strings.Index(content[searchStart:], "]{")
		if bracketPos == -1 {
			start = start + uuidStart + 1
			continue
		}

		textStart := searchStart + bracketPos + 2
		if textStart >= len(content) {
			start = start + uuidStart + 1
			continue
		}

		closeBracePos := strings.Index(content[textStart:], "}")
		if closeBracePos == -1 {
			start = start + uuidStart + 1
			continue
		}

		middle := content[searchStart : searchStart+bracketPos]
		text := content[textStart : textStart+closeBracePos]

		parts := strings.SplitN(middle, ".", 2)
		if len(parts) == 2 {
			refs = append(refs, UUIDReference{
				Type:    parts[0],
				ID:      parts[1],
				Display: text,
			})
		}

		start = textStart + closeBracePos + 1
	}

	return refs
}

// UUIDReference represents a @UUID{} reference found in content
type UUIDReference struct {
	Type    string
	ID      string
	Display string
}
