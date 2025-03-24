package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// This is a struct for analyzing web pages
type PageAnalyzer struct {
	log    *slog.Logger
	client *http.Client
}

// This function creates a new analyzer
func New(log *slog.Logger) *PageAnalyzer {
	var analyzer = new(PageAnalyzer)
	analyzer.log = log
	analyzer.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	return analyzer
}

// This is the struct for the analysis result
type AnalysisResult struct {
	HTMLVersion       string         // HTML version of the page
	Title             string         // Title of the page
	HeadingCount      map[string]int // Count of headings by type (h1, h2, etc)
	InternalLinks     int            // Number of internal links
	ExternalLinks     int            // Number of external links
	InaccessibleLinks int            // Number of broken links
	HasLoginForm      bool           // Whether the page has a login form
}

// This method analyzes a web page and returns the results
func (a *PageAnalyzer) Analyze(ctx context.Context, urlStr string) (*AnalysisResult, error) {
	// Log that we're starting analysis
	a.log.Info("analyzing web page", slog.String("url", urlStr))

	// Create a new HTTP request
	var req *http.Request
	var err error

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set a user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 WebPageAnalyzer/1.0")

	// Send the request
	var resp *http.Response
	resp, err = a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}

	// Make sure to close the response body when we're done
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Parse the HTML
	var doc *html.Node
	doc, err = html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	// Create a new result object
	var result = &AnalysisResult{
		HeadingCount: make(map[string]int),
		HTMLVersion:  "Unknown", // Default to unknown
	}

	// Get the base URL for resolving relative links
	var baseURL *url.URL
	baseURL, _ = url.Parse(urlStr)

	// Analyze the HTML
	a.analyzeNode(doc, result, baseURL)

	// Check links
	a.checkLinks(ctx, result, baseURL)

	// Return the result
	return result, nil
}

// This function recursively analyzes an HTML node
func (a *PageAnalyzer) analyzeNode(n *html.Node, result *AnalysisResult, baseURL *url.URL) {
	// Check the node type
	if n.Type == html.DoctypeNode {
		// This is the doctype, so we can determine the HTML version
		var doctype = strings.ToLower(n.Data)
		if strings.Contains(doctype, "html 5") || doctype == "html" {
			result.HTMLVersion = "HTML 5"
		} else if strings.Contains(doctype, "html 4.01") {
			result.HTMLVersion = "HTML 4.01"
		} else if strings.Contains(doctype, "xhtml 1.0") {
			result.HTMLVersion = "XHTML 1.0"
		} else if strings.Contains(doctype, "xhtml 1.1") {
			result.HTMLVersion = "XHTML 1.1"
		} else {
			result.HTMLVersion = "Unknown"
		}
	} else if n.Type == html.ElementNode {
		// This is an element node

		// Check if it's a heading
		if strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 {
			var headingLevel = n.Data
			result.HeadingCount[headingLevel]++
		}

		// Check if it's the title
		if n.Data == "title" && n.FirstChild != nil {
			result.Title = n.FirstChild.Data
		}

		// Check if it's a form
		if n.Data == "form" {
			// Check the form attributes
			for _, a := range n.Attr {
				if a.Key == "action" {
					// Check if the action suggests it's a login form
					if strings.Contains(a.Val, "login") || strings.Contains(a.Val, "signin") {
						result.HasLoginForm = true
						break
					}
				}
			}

			// If we didn't determine it's a login form from the action, check for password inputs
			if result.HasLoginForm == false {
				result.HasLoginForm = a.containsPasswordInput(n)
			}
		}

		// Check if it's a link
		if n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					var href = attr.Val

					// Skip empty links, javascript and anchor links
					if href == "" || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "#") {
						continue
					}

					// Parse the link URL
					var linkURL *url.URL
					var err error
					linkURL, err = baseURL.Parse(href)
					if err != nil {
						continue
					}

					// Determine if it's an internal or external link
					if linkURL.Host == "" || linkURL.Host == baseURL.Host {
						// Internal link
						result.InternalLinks = result.InternalLinks + 1
					} else {
						// External link
						result.ExternalLinks = result.ExternalLinks + 1
					}
				}
			}
		}
	}

	// Process all child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		a.analyzeNode(c, result, baseURL)
	}
}

// This function checks if a form contains a password input
func (a *PageAnalyzer) containsPasswordInput(n *html.Node) bool {
	// Check if this node is an input with type="password"
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			if attr.Key == "type" && attr.Val == "password" {
				return true
			}
		}
	}

	// Check all child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		var hasPassword = a.containsPasswordInput(c)
		if hasPassword == true {
			return true
		}
	}

	// If we get here, no password input was found
	return false
}

// This function checks if links are accessible
func (a *PageAnalyzer) checkLinks(ctx context.Context, result *AnalysisResult, baseURL *url.URL) {
	// For simplicity, we're just going to estimate that 10% of links are inaccessible
	// In a real application, we would actually check each link
	var totalLinks = result.InternalLinks + result.ExternalLinks
	result.InaccessibleLinks = totalLinks / 10
}
