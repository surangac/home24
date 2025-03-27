package analyzer

import (
	"context"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

// PageAnalyzer defines the main interface for webpage analysis
type PageAnalyzer interface {
	Analyze(ctx context.Context, urlStr string) (*AnalysisResult, error)
}

// LinkChecker defines the interface for checking link accessibility
type LinkChecker interface {
	CheckAccessibility(ctx context.Context, urlStr string) bool
	CheckWithRetry(ctx context.Context, urlStr string) bool
}

// HTMLParser defines the interface for HTML parsing operations
type HTMLParser interface {
	ParseHTML(reader io.Reader) (*html.Node, error)
	ExtractTitle(doc *html.Node) string
	ExtractHeadings(doc *html.Node) map[string]int
	ExtractLinks(doc *html.Node, baseURL *url.URL) []LinkInfo
	ExtractForms(doc *html.Node) []*html.Node
	ExtractHTMLVersion(doc *html.Node) string
}

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	RecordDuration(duration float64)
	RecordResults(result *AnalysisResult)
	RecordError(err error)
	RecordRequest()
}

// Logger defines the interface for logging operations
type Logger interface {
	LogAnalysisStart(url string)
	LogAnalysisError(err error, url string)
	LogAnalysisComplete(url string, duration float64)
	LogLinkCheck(url string, isAccessible bool)
	LogDebug(msg string, args ...interface{})
}
