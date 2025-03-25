package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"home24/internal/metrics"

	"golang.org/x/net/html"
)

// PageAnalyzer helps us dig into web pages and figure out what makes them tick
type PageAnalyzer struct {
	log    *slog.Logger
	client *http.Client
}

// New creates a fresh analyzer with some sensible defaults
func New(log *slog.Logger) *PageAnalyzer {
	var analyzer = new(PageAnalyzer)
	analyzer.log = log
	analyzer.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	return analyzer
}

// AnalysisResult holds all the interesting stuff we find on a webpage
type AnalysisResult struct {
	HTMLVersion       string         // What version of HTML this page is using
	Title             string         // The page title that shows up in browser tabs
	HeadingCount      map[string]int // How many of each heading type we found (h1, h2, etc)
	InternalLinks     int            // Links that stay within the same site
	ExternalLinks     int            // Links that point to other sites
	InaccessibleLinks int            // Links that seem to be broken or unreachable
	HasLoginForm      bool           // Whether we found a login form on the page
}

// LinkInfo keeps track of what we know about each link we find
type LinkInfo struct {
	URL          string
	IsInternal   bool
	IsAccessible bool
}

// Analyze digs into a webpage and tells us everything we found
func (a *PageAnalyzer) Analyze(ctx context.Context, urlStr string) (*AnalysisResult, error) {
	// Track how long this analysis takes
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.AnalysisDuration.Observe(duration)
	}()

	// Count this analysis request
	metrics.AnalysisRequests.Inc()

	// Let's start digging into this page
	a.log.Info("analyzing web page", slog.String("url", urlStr))

	// Set up our request to fetch the page
	var req *http.Request
	var err error

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		metrics.AnalysisErrors.Inc()
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// We'll pretend to be a regular browser to avoid getting blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 WebPageAnalyzer/1.0")

	// Go fetch that page for us
	var resp *http.Response
	resp, err = a.client.Do(req)
	if err != nil {
		metrics.AnalysisErrors.Inc()
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}

	// Don't forget to clean up after ourselves
	defer resp.Body.Close()

	// Make sure we got a good response
	if resp.StatusCode != http.StatusOK {
		metrics.AnalysisErrors.Inc()
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Parse the HTML so we can dig through it
	var doc *html.Node
	doc, err = html.Parse(resp.Body)
	if err != nil {
		metrics.AnalysisErrors.Inc()
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	// Set up our results container
	var result = &AnalysisResult{
		HeadingCount: make(map[string]int),
		HTMLVersion:  "Unknown", // We'll try to figure this out later
	}

	// Figure out the base URL so we can make sense of relative links
	var baseURL *url.URL
	baseURL, _ = url.Parse(urlStr)

	// Set up our channels for collecting link info
	linkChan := make(chan LinkInfo, 100)

	// We'll need this to make sure all our goroutines finish up
	var wg sync.WaitGroup

	// Start collecting links in the background
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.collectLinks(doc, baseURL, linkChan)
		close(linkChan)
	}()

	// Process all those links we found
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.processLinks(ctx, linkChan, result)
	}()

	// Look through the HTML structure
	a.analyzeNodeParallel(doc, result)

	// Wait for everything to wrap up
	wg.Wait()

	// Update our metrics with what we found
	metrics.LinkCounts.WithLabelValues("internal").Set(float64(result.InternalLinks))
	metrics.LinkCounts.WithLabelValues("external").Set(float64(result.ExternalLinks))
	metrics.LinkCounts.WithLabelValues("inaccessible").Set(float64(result.InaccessibleLinks))

	for level, count := range result.HeadingCount {
		metrics.HeadingCounts.WithLabelValues(level).Set(float64(count))
	}

	if result.HasLoginForm {
		metrics.LoginFormCount.Inc()
	}

	metrics.HTMLVersionCount.WithLabelValues(result.HTMLVersion).Inc()

	return result, nil
}

// collectLinks digs through the HTML and finds all the links
func (a *PageAnalyzer) collectLinks(n *html.Node, baseURL *url.URL, linkChan chan<- LinkInfo) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				var href = attr.Val

				// Skip the boring links that don't go anywhere
				if href == "" || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "#") {
					continue
				}

				// Try to make sense of this link URL
				var linkURL *url.URL
				var err error
				linkURL, err = baseURL.Parse(href)
				if err != nil {
					continue
				}

				// Figure out if this link stays on the same site
				isInternal := linkURL.Host == "" || linkURL.Host == baseURL.Host
				linkChan <- LinkInfo{
					URL:        href,
					IsInternal: isInternal,
				}
			}
		}
	}

	// Keep digging through all the nested elements
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		a.collectLinks(c, baseURL, linkChan)
	}
}

// processLinks checks all our links to see if they work
func (a *PageAnalyzer) processLinks(ctx context.Context, linkChan <-chan LinkInfo, result *AnalysisResult) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	// Don't overwhelm the server with too many requests at once
	semaphore := make(chan struct{}, 10)

	for link := range linkChan {
		wg.Add(1)
		semaphore <- struct{}{} // Grab a slot

		go func(l LinkInfo) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release our slot when done

			// Check if this link actually works
			isAccessible := a.checkLinkAccessibility(ctx, l.URL)
			a.log.Debug("link check result",
				slog.String("url", l.URL),
				slog.Bool("is_internal", l.IsInternal),
				slog.Bool("is_accessible", isAccessible))

			// Update our counts safely
			mutex.Lock()
			if l.IsInternal {
				result.InternalLinks++
			} else {
				result.ExternalLinks++
			}
			if !isAccessible {
				result.InaccessibleLinks++
			}
			mutex.Unlock()
		}(link)
	}

	wg.Wait()
}

// checkLinkAccessibility tries to figure out if a link actually works
func (a *PageAnalyzer) checkLinkAccessibility(ctx context.Context, urlStr string) bool {
	// First try a quick HEAD request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, urlStr, nil)
	if err != nil {
		a.log.Debug("error creating HEAD request",
			slog.String("url", urlStr),
			slog.String("error", err.Error()))
		return false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 WebPageAnalyzer/1.0")
	resp, err := a.client.Do(req)
	if err != nil {
		a.log.Debug("error in HEAD request",
			slog.String("url", urlStr),
			slog.String("error", err.Error()))
		return false
	}
	defer resp.Body.Close()

	// If HEAD works, we're good to go
	if resp.StatusCode == http.StatusOK {
		return true
	}

	// Some servers don't like HEAD requests, so let's try GET
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		a.log.Debug("error creating GET request",
			slog.String("url", urlStr),
			slog.String("error", err.Error()))
		return false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 WebPageAnalyzer/1.0")
	resp, err = a.client.Do(req)
	if err != nil {
		a.log.Debug("error in GET request",
			slog.String("url", urlStr),
			slog.String("error", err.Error()))
		return false
	}
	defer resp.Body.Close()

	// Any 2xx status means the link works
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// analyzeNodeParallel digs through the HTML structure and finds interesting stuff
func (a *PageAnalyzer) analyzeNodeParallel(n *html.Node, result *AnalysisResult) {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Look at what we found in this node
	if n.Type == html.DoctypeNode {
		mutex.Lock()
		// This tells us what version of HTML we're dealing with
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
		mutex.Unlock()
	} else if n.Type == html.ElementNode {
		mutex.Lock()
		// Count up all the headings we find
		if strings.HasPrefix(n.Data, "h") && len(n.Data) == 2 {
			var headingLevel = n.Data
			result.HeadingCount[headingLevel]++
		}

		// Grab the page title if we find it
		if n.Data == "title" && n.FirstChild != nil {
			result.Title = n.FirstChild.Data
		}
		mutex.Unlock()

		// If we found a form, let's check it out
		if n.Data == "form" {
			wg.Add(1)
			go func(node *html.Node) {
				defer wg.Done()
				a.analyzeForm(node, result, &mutex)
			}(n)
		}
	}

	// Keep looking through all the nested elements
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		wg.Add(1)
		go func(child *html.Node) {
			defer wg.Done()
			a.analyzeNodeParallel(child, result)
		}(c)
	}

	wg.Wait()
}

// analyzeForm figures out if this form is for logging in
func (a *PageAnalyzer) analyzeForm(n *html.Node, result *AnalysisResult, mutex *sync.Mutex) {
	// Check what this form is supposed to do
	for _, a := range n.Attr {
		if a.Key == "action" {
			// Look for hints in the form's action URL
			if strings.Contains(a.Val, "login") || strings.Contains(a.Val, "signin") {
				mutex.Lock()
				result.HasLoginForm = true
				mutex.Unlock()
				return
			}
		}
	}

	// If we didn't figure it out from the action, look for password fields
	if !result.HasLoginForm {
		mutex.Lock()
		result.HasLoginForm = a.containsPasswordInput(n)
		mutex.Unlock()
	}
}

// containsPasswordInput checks if this form has a password field
func (a *PageAnalyzer) containsPasswordInput(n *html.Node) bool {
	// Is this a password input field?
	if n.Type == html.ElementNode && n.Data == "input" {
		for _, attr := range n.Attr {
			if attr.Key == "type" && attr.Val == "password" {
				return true
			}
		}
	}

	// Look through all the nested elements
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		var hasPassword = a.containsPasswordInput(c)
		if hasPassword == true {
			return true
		}
	}

	// No password field found here
	return false
}
