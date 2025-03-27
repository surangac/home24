package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

// Test create a new analyzer
func TestNew(t *testing.T) {
	var config = DefaultConfig()
	var analyzer = NewDefaultPageAnalyzer(&config)

	if analyzer == nil {
		t.Fatal("expected analyzer to be non-nil")
	}
}

// Test can analyze a webpage with parallel processing
func TestAnalyzeParallel(t *testing.T) {
	// Create test servers for different scenarios
	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<h1>Heading 1</h1>
			<h2>Heading 2</h2>
			<h2>Heading 2-2</h2>
			<h3>Heading 3</h3>
			<a href="/">Home</a>
			<a href="/about">About</a>
			<a href="https://example.com">External</a>
			<form action="/login">
				<input type="text" name="username">
				<input type="password" name="password">
				<button type="submit">Login</button>
			</form>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer mainServer.Close()

	// Create a server that simulates slow responses
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// Create a server that returns errors
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer errorServer.Close()

	var config = DefaultConfig()
	var analyzer = NewDefaultPageAnalyzer(&config)

	// Test with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var result, err = analyzer.Analyze(ctx, mainServer.URL)
	if err != nil {
		t.Fatalf("Error analyzing page: %v", err)
	}

	// Test HTML version detection
	if result.HTMLVersion != "HTML 5" {
		t.Errorf("Expected HTML 5, got %s", result.HTMLVersion)
	}

	// Test title extraction
	if result.Title != "Test Page" {
		t.Errorf("Expected 'Test Page', got %s", result.Title)
	}

	// Test heading counts
	if result.Headings["h1"] != 1 {
		t.Errorf("Expected 1 h1, got %d", result.Headings["h1"])
	}
	if result.Headings["h2"] != 2 {
		t.Errorf("Expected 2 h2, got %d", result.Headings["h2"])
	}
	if result.Headings["h3"] != 1 {
		t.Errorf("Expected 1 h3, got %d", result.Headings["h3"])
	}

	// Test link counting
	var internalCount = 0
	var externalCount = 0
	for _, link := range result.Links {
		if link.IsInternal {
			internalCount++
		} else {
			externalCount++
		}
	}
	if internalCount != 2 {
		t.Errorf("Expected 2 internal links, got %d", internalCount)
	}
	if externalCount != 1 {
		t.Errorf("Expected 1 external link, got %d", externalCount)
	}

	// Test login form detection
	if !result.HasLoginForm {
		t.Error("Expected to detect a login form")
	}
}

// Test parallel link processing
func TestParallelLinkProcessing(t *testing.T) {
	// Create a test server that returns different status codes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/slow":
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/error":
			w.WriteHeader(http.StatusNotFound)
		case "/redirect":
			w.Header().Set("Location", "/ok")
			w.WriteHeader(http.StatusMovedPermanently)
		case "/ok2":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create HTML with multiple links
	html := `
	<!DOCTYPE html>
	<html>
	<body>
		<a href="` + server.URL + `/ok">OK Link</a>
		<a href="` + server.URL + `/slow">Slow Link</a>
		<a href="` + server.URL + `/error">Error Link</a>
		<a href="` + server.URL + `/redirect">Redirect Link</a>
		<a href="` + server.URL + `/ok2">Another OK Link</a>
	</body>
	</html>
	`

	// Create a test server serving the HTML
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer testServer.Close()

	var config = DefaultConfig()
	var analyzer = NewDefaultPageAnalyzer(&config)

	// Test with a longer timeout to ensure all links are checked
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result, err = analyzer.Analyze(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Error analyzing page: %v", err)
	}

	// Verify link counts
	var externalCount = 0
	for _, link := range result.Links {
		if !link.IsInternal {
			externalCount++
		}
	}
	if externalCount != 5 {
		t.Errorf("Expected 5 external links, got %d", externalCount)
	}
	if len(result.Links)-result.AccessibleLinks != 1 {
		t.Errorf("Expected 1 inaccessible link, got %d", len(result.Links)-result.AccessibleLinks)
	}
}

// Test concurrent form analysis
func TestConcurrentFormAnalysis(t *testing.T) {
	// Create HTML with multiple forms
	html := `
	<!DOCTYPE html>
	<html>
	<body>
		<form action="/login">
			<input type="text" name="username">
			<input type="password" name="password">
			<button type="submit">Login</button>
		</form>
		<form action="/signup">
			<input type="text" name="email">
			<input type="password" name="password">
			<button type="submit">Sign Up</button>
		</form>
		<form action="/search">
			<input type="text" name="query">
			<button type="submit">Search</button>
		</form>
	</body>
	</html>
	`

	// Create a test server serving the HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer server.Close()

	var config = DefaultConfig()
	var analyzer = NewDefaultPageAnalyzer(&config)

	var result, err = analyzer.Analyze(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Error analyzing page: %v", err)
	}

	// Verify login form detection
	if !result.HasLoginForm {
		t.Error("Expected to detect a login form")
	}
}

// Test error handling
func TestAnalyzeError(t *testing.T) {
	var config = DefaultConfig()
	var analyzer = NewDefaultPageAnalyzer(&config)

	// Test with invalid URL
	var _, err = analyzer.Analyze(context.Background(), "http://invalid-url-that-does-not-exist.example")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var _, err2 = analyzer.Analyze(ctx, "http://example.com")
	if err2 == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}

// Helper function to parse HTML string
func ParseHTMLString(htmlString string) (*html.Node, error) {
	var reader = strings.NewReader(htmlString)
	return html.Parse(reader)
}
