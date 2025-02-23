package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// FetchDynamicCoursePage navigates to the given DTU course URL,
// waits for the dynamic content to load, and returns a goquery Document.
func FetchDynamicCoursePage(url string) (*goquery.Document, error) {
	// Create a base Chromedp context.
	baseCtx := context.Background()
	// Create a new Chromedp context.
	ctx, cancel := chromedp.NewContext(baseCtx)
	defer cancel()

	// Set a timeout; if the page doesn't load within this duration, cancel the operation.
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 5*time.Second)
	defer cancelTimeout()

	var htmlContent string

	// Run the browser automation steps in the timeout context.
	err := chromedp.Run(timeoutCtx,
		// Navigate to the page.
		chromedp.Navigate(url),
		// Disable automatic reload by overriding setTimeout.
		chromedp.Evaluate(`window.setTimeout = function() {}`, nil),
		// Wait until the course information element is visible.
		chromedp.WaitVisible("div.box.information", chromedp.ByQuery),
		// Capture the full HTML.
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}

	// Create a goquery document from the fully rendered HTML string.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// ExtractText extracts the plain text from the specified selector.
// If an adjacentSelector is provided (as the first element in adjacentSelector),
// it will search that element for anchor tags and append a comma‐separated Markdown
// formatted link string to the plain text.
func ExtractText(doc *goquery.Document, selector string, fieldName string, adjacentSelector ...string) string {
	sel := doc.Find(selector).FilterFunction(func(i int, s *goquery.Selection) bool {
		// Try getting the text from a nested <label><a>, but if absent, fallback to <label>
		prev := s.Prev()

		labelText := strings.TrimSpace(prev.Find("label a").Text())
		if labelText == "" {
			labelText = strings.TrimSpace(prev.Find("label").Text())
		}
		return labelText == fieldName
	})

	var sb strings.Builder
	sel.Contents().Each(func(i int, node *goquery.Selection) {
		switch goquery.NodeName(node) {
		case "#text":
			sb.WriteString(node.Text())
		case "a":
			href, exists := node.Attr("href")
			text := strings.TrimSpace(node.Text())
			if exists && href != "" && text != "" {
				if strings.HasPrefix(href, "mailto:") {
					sb.WriteString(fmt.Sprintf("(%s)", href))
				} else {
					if fieldName == "Home page" {
						text = "Home page"
						sb.WriteString(fmt.Sprintf("[%s](%s)", text, href))
						break
					}
					sb.WriteString(fmt.Sprintf("[%s](%s)", text, href))
				}
			} else {
				sb.WriteString(text)
			}
		default:
			sb.WriteString(node.Text())
		}
	})

	plainText := strings.TrimSpace(sb.String())

	var links []string
	if len(adjacentSelector) > 0 {
		adjSel := doc.Find(adjacentSelector[0])
		adjSel.Find("a").Each(func(i int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			aText := strings.TrimSpace(a.Text())
			if exists && href != "" && aText != "" {
				links = append(links, fmt.Sprintf("([Link to %s](%s))", aText, href))
			}
		})
	}

	if len(links) > 0 {
		joined := strings.Join(links, ", ")
		if plainText == "" {
			return joined
		}
		return fmt.Sprintf("%s %s", plainText, joined)
	}
	return plainText
}

// ExtractFieldByName finds the table row whose label exactly equals fieldName,
// then extracts Markdown-formatted text from the second cell.
func ExtractFieldByName(doc *goquery.Document, fieldName string) string {
	// First, try to find the td with a label that has an <a>
	labelSelector := fmt.Sprintf("td:has(label a:contains('%s'))", fieldName)
	if doc.Find(labelSelector).Length() == 0 {
		// Fallback: use the label without the anchor.
		labelSelector = fmt.Sprintf("td:has(label:contains('%s'))", fieldName)
	}
	valueSelector := labelSelector + " + td"
	return ExtractText(doc, valueSelector, fieldName, labelSelector)
}

// CourseTypeBlock holds one “top-level” item plus its hidden expansions.
type CourseTypeBlock struct {
	Title      string
	Expansions []string
}

// GetSectionName returns a header for the block.
func (ctb CourseTypeBlock) GetSectionName() string {
	return "Course Type: " + ctb.Title
}

// GetSectionValue returns the formatted expansions wrapped in spoiler syntax.
func (ctb CourseTypeBlock) GetSectionValue() string {
	if len(ctb.Expansions) == 0 {
		return ""
	}
	// Join all expansions into a single line.
	exp := strings.Join(ctb.Expansions, ", ")
	return exp
}

// GetSectionInline returns whether this section should be inline.
func (ctb CourseTypeBlock) GetSectionInline() bool {
	return false
}

// SetInLine panics since inline formatting is not supported on individual blocks.
func (ctb CourseTypeBlock) SetInLine(isInline bool) {
	panic("CourseTypeBlock does not support inline formatting")
}

func ExtractCourseTypeAdvanced(doc *goquery.Document) []CourseTypeBlock {
	var results []CourseTypeBlock

	// Find the <td> for "Course type"
	cell := doc.Find("td:has(label:contains('Course type')) + td")

	// The top-level <div> that is not #studiebox, e.g. <div>BSc</div>
	cell.Find("div").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		if id == "studiebox" {
			// Handle the #studiebox div
			results = append(results, parseStudiebox(s)...)
		} else {
			// This is the top-level text, e.g. "BSc"
			txt := strings.TrimSpace(s.Text())
			if txt != "" {
				// Put it in results as a single block with no expansions
				results = append(results, CourseTypeBlock{Title: txt, Expansions: nil})
			}
		}
	})

	return results
}

// parseStudiebox is a helper that looks inside <div id="studiebox"> and builds CourseTypeBlocks.
func parseStudiebox(studiebox *goquery.Selection) []CourseTypeBlock {
	var blocks []CourseTypeBlock

	var currentBlock *CourseTypeBlock

	studiebox.Children().Each(func(i int, span *goquery.Selection) {
		text := strings.TrimSpace(span.Text())

		// If it's an expander <span class="expander">
		if span.HasClass("expander") {
			// Start a new block
			if text != "" {
				// The “see more” text is typically part of the top-level item
				blocks = append(blocks, CourseTypeBlock{
					Title:      text,
					Expansions: []string{},
				})
				// Make currentBlock point to the last block we appended
				currentBlock = &blocks[len(blocks)-1]
			}
		} else {
			// It's presumably <span style="display:none">
			// => expansions for the *currentBlock*
			if currentBlock != nil && text != "" {
				// Skip any expansions that contain "see more" again
				if !strings.Contains(strings.ToLower(text), "see more") {
					currentBlock.Expansions = append(currentBlock.Expansions, text)
				}
			}
		}
	})

	return blocks
}
