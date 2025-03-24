package analyzer

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// Test that we can create a new analyzer
func TestNew(t *testing.T) {
	// Create a logger
	var log = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create the analyzer
	var analyzer = New(log)

	// Make sure it's not nil
	if analyzer == nil {
		t.Fatal("expected analyzer to be non-nil")
	}
}

// Test that we can analyze a webpage
func TestAnalyze(t *testing.T) {
	// Create a test server that serves HTML
	var ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is the HTML that will be served
		var html = `
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

		// Set content type and send the HTML
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))

	// Make sure we close the server when the test ends
	defer ts.Close()

	// Create a logger and analyzer
	var log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	var analyzer = New(log)

	// Analyze the test server's URL
	var result, err = analyzer.Analyze(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("Error analyzing page: %v", err)
	}

	// Check that the results are what we expect
	if result.HTMLVersion != "HTML 5" {
		t.Errorf("Expected HTML 5, got %s", result.HTMLVersion)
	}

	if result.Title != "Test Page" {
		t.Errorf("Expected 'Test Page', got %s", result.Title)
	}

	if result.HeadingCount["h1"] != 1 {
		t.Errorf("Expected 1 h1, got %d", result.HeadingCount["h1"])
	}

	if result.HeadingCount["h2"] != 2 {
		t.Errorf("Expected 2 h2, got %d", result.HeadingCount["h2"])
	}

	if result.HeadingCount["h3"] != 1 {
		t.Errorf("Expected 1 h3, got %d", result.HeadingCount["h3"])
	}

	if result.InternalLinks != 2 {
		t.Errorf("Expected 2 internal links, got %d", result.InternalLinks)
	}

	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}

	if result.HasLoginForm != true {
		t.Error("Expected to detect a login form")
	}
}

// Test that we get an error for an invalid URL
func TestAnalyzeError(t *testing.T) {
	// Create a logger and analyzer
	var log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	var analyzer = New(log)

	// Try to analyze an invalid URL
	var _, err = analyzer.Analyze(context.Background(), "http://invalid-url-that-does-not-exist.example")

	// Make sure we got an error
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

// Test that we can detect password inputs in forms
func TestContainsPasswordInput(t *testing.T) {
	// Create a logger and analyzer
	var log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	var analyzer = New(log)

	// Test with a form that has a password input
	var formHTML = `
	<form>
		<input type="text" name="username">
		<input type="password" name="password">
	</form>
	`

	// Parse the HTML
	var doc, err = ParseHTMLString(formHTML)
	if err != nil {
		t.Fatalf("Error parsing HTML: %v", err)
	}

	// Check if it has a password input
	var hasPasswordInput = analyzer.containsPasswordInput(doc)

	// It should have one
	if hasPasswordInput != true {
		t.Error("Expected to detect password input")
	}

	// Test with a form that doesn't have a password input
	formHTML = `
	<form>
		<input type="text" name="username">
		<input type="submit" value="Submit">
	</form>
	`

	// Parse the HTML
	doc, err = ParseHTMLString(formHTML)
	if err != nil {
		t.Fatalf("Error parsing HTML: %v", err)
	}

	// Check if it has a password input
	hasPasswordInput = analyzer.containsPasswordInput(doc)

	// It should not have one
	if hasPasswordInput == true {
		t.Error("Expected not to detect password input")
	}
}

// Helper function to parse HTML strings
func ParseHTMLString(htmlString string) (*html.Node, error) {
	// Create a reader from the string and parse it
	var reader = strings.NewReader(htmlString)
	return html.Parse(reader)
}
