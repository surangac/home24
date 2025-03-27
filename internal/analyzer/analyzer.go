package analyzer

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

// DefaultPageAnalyzer implements the PageAnalyzer interface
type DefaultPageAnalyzer struct {
	client  *http.Client
	parser  HTMLParser
	checker LinkChecker
	metrics MetricsCollector
	log     Logger
	config  *AnalyzerConfig
}

// NewDefaultPageAnalyzer creates a new DefaultPageAnalyzer
func NewDefaultPageAnalyzer(config *AnalyzerConfig) *DefaultPageAnalyzer {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	log := NewAnalyzerLogger(slog.Default())
	parser := NewDefaultHTMLParser(log)
	checker := NewDefaultLinkChecker(client, log, config)
	metrics := NewPrometheusMetricsCollector(*config)

	return &DefaultPageAnalyzer{
		client:  client,
		parser:  parser,
		checker: checker,
		metrics: metrics,
		log:     log,
		config:  config,
	}
}

// Analyze performs a complete analysis of a webpage
func (a *DefaultPageAnalyzer) Analyze(ctx context.Context, targetURL string) (*AnalysisResult, error) {
	startTime := time.Now()
	a.log.LogAnalysisStart(targetURL)

	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, NewAnalysisError(ErrInvalidURL, "invalid URL", err)
	}

	// Fetch the page
	resp, err := a.fetchPage(ctx, parsedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse HTML
	doc, err := a.parser.ParseHTML(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract information
	title := a.parser.ExtractTitle(doc)
	headings := a.parser.ExtractHeadings(doc)
	links := a.parser.ExtractLinks(doc, parsedURL)
	forms := a.parser.ExtractForms(doc)
	htmlVersion := a.parser.ExtractHTMLVersion(doc)

	// Check links concurrently
	linkResults := make(map[string]bool)
	for _, link := range links {
		isAccessible := a.checker.CheckAccessibility(ctx, link.URL)
		linkResults[link.URL] = isAccessible
		a.log.LogLinkCheck(link.URL, isAccessible)
	}

	// Count accessible links
	accessibleLinks := 0
	for _, isAccessible := range linkResults {
		if isAccessible {
			accessibleLinks++
		}
	}

	// Check for login form
	hasLoginForm := false
	for _, form := range forms {
		if a.isLoginForm(form) {
			hasLoginForm = true
			break
		}
	}

	// Create result
	result := &AnalysisResult{
		URL:             targetURL,
		Title:           title,
		Headings:        headings,
		Links:           links,
		AccessibleLinks: accessibleLinks,
		HasLoginForm:    hasLoginForm,
		HTMLVersion:     htmlVersion,
	}

	// Record metrics
	duration := time.Since(startTime).Seconds()
	a.metrics.RecordDuration(duration)
	a.metrics.RecordResults(result)

	a.log.LogAnalysisComplete(targetURL, duration)
	return result, nil
}

// fetchPage fetches the webpage with retry logic
func (a *DefaultPageAnalyzer) fetchPage(ctx context.Context, url *url.URL) (*http.Response, error) {
	var resp *http.Response

	// Try with retry logic
	for i := 0; i < a.config.RetryAttempts; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
		if err != nil {
			return nil, NewAnalysisError(ErrFetchFailed, "failed to create request", err)
		}

		req.Header.Set("User-Agent", a.config.UserAgent)
		resp, err = a.client.Do(req)
		if err != nil {
			if i == a.config.RetryAttempts-1 {
				return nil, NewAnalysisError(ErrFetchFailed, "failed to fetch page", err)
			}
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			if i == a.config.RetryAttempts-1 {
				return nil, NewAnalysisError(ErrFetchFailed, "server error", nil)
			}
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		return resp, nil
	}

	return nil, NewAnalysisError(ErrFetchFailed, "max retries exceeded", nil)
}

// isLoginForm checks if a form is likely a login form
func (a *DefaultPageAnalyzer) isLoginForm(form *html.Node) bool {
	hasPassword := false
	hasUsername := false
	hasSubmit := false

	var checkInputs func(*html.Node)
	checkInputs = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "input" {
				var inputType string
				for _, attr := range n.Attr {
					if attr.Key == "type" {
						inputType = attr.Val
						break
					}
				}

				switch inputType {
				case "password":
					hasPassword = true
				case "text", "email":
					hasUsername = true
				case "submit":
					hasSubmit = true
				}
			} else if n.Data == "button" {
				for _, attr := range n.Attr {
					if attr.Key == "type" && attr.Val == "submit" {
						hasSubmit = true
						break
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			checkInputs(c)
		}
	}

	checkInputs(form)
	return hasPassword && hasUsername && hasSubmit
}
