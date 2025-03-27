package analyzer

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// DefaultHTMLParser implements the HTMLParser interface
type DefaultHTMLParser struct {
	log Logger
}

// NewDefaultHTMLParser creates a new DefaultHTMLParser
func NewDefaultHTMLParser(log Logger) *DefaultHTMLParser {
	return &DefaultHTMLParser{log: log}
}

// ParseHTML parses HTML from a reader
func (p *DefaultHTMLParser) ParseHTML(reader io.Reader) (*html.Node, error) {
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, NewAnalysisError(ErrParseFailed, "failed to parse HTML", err)
	}
	return doc, nil
}

// ExtractTitle extracts the page title
func (p *DefaultHTMLParser) ExtractTitle(doc *html.Node) string {
	var title string
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}
	findTitle(doc)
	return title
}

// ExtractHeadings extracts heading counts
func (p *DefaultHTMLParser) ExtractHeadings(doc *html.Node) map[string]int {
	headingCount := make(map[string]int)
	var findHeadings func(*html.Node)
	findHeadings = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 {
			headingCount[n.Data]++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findHeadings(c)
		}
	}
	findHeadings(doc)
	return headingCount
}

// ExtractLinks extracts all links from the document
func (p *DefaultHTMLParser) ExtractLinks(doc *html.Node, baseURL *url.URL) []LinkInfo {
	var links []LinkInfo
	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					if href == "" || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "#") {
						continue
					}

					linkURL, err := baseURL.Parse(href)
					if err != nil {
						continue
					}

					isInternal := linkURL.Host == "" || linkURL.Host == baseURL.Host
					links = append(links, LinkInfo{
						URL:        href,
						IsInternal: isInternal,
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c)
		}
	}
	findLinks(doc)
	return links
}

// ExtractForms extracts all forms from the document
func (p *DefaultHTMLParser) ExtractForms(doc *html.Node) []*html.Node {
	var forms []*html.Node
	var findForms func(*html.Node)
	findForms = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			forms = append(forms, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findForms(c)
		}
	}
	findForms(doc)
	return forms
}

// ExtractHTMLVersion extracts the HTML version from the doctype
func (p *DefaultHTMLParser) ExtractHTMLVersion(doc *html.Node) string {
	var version string
	var findDoctype func(*html.Node)
	findDoctype = func(n *html.Node) {
		if n.Type == html.DoctypeNode {
			doctype := strings.ToLower(n.Data)
			if strings.Contains(doctype, "html 5") || doctype == "html" {
				version = "HTML 5"
			} else if strings.Contains(doctype, "html 4.01") {
				version = "HTML 4.01"
			} else if strings.Contains(doctype, "xhtml 1.0") {
				version = "XHTML 1.0"
			} else if strings.Contains(doctype, "xhtml 1.1") {
				version = "XHTML 1.1"
			} else {
				version = "Unknown"
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findDoctype(c)
		}
	}
	findDoctype(doc)
	return version
}
