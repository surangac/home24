package analyzer

import "time"

// AnalyzerConfig holds all configuration options for the PageAnalyzer
type AnalyzerConfig struct {
	// HTTP client configuration
	Timeout            time.Duration
	MaxConcurrentLinks int
	UserAgent          string
	RetryAttempts      int

	// Analysis configuration
	MaxLinksPerPage int
	MaxDepth        int

	// Metrics configuration
	EnableMetrics bool
	MetricsPrefix string
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() AnalyzerConfig {
	return AnalyzerConfig{
		Timeout:            10 * time.Second,
		MaxConcurrentLinks: 10,
		UserAgent:          "Mozilla/5.0 WebPageAnalyzer/1.0",
		RetryAttempts:      3,
		MaxLinksPerPage:    100,
		MaxDepth:           2,
		EnableMetrics:      true,
		MetricsPrefix:      "webpage_analyzer",
	}
}
