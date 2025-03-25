package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AnalysisDuration tracks how long each page analysis takes
	AnalysisDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "webpage_analysis_duration_seconds",
		Help:    "Time taken to analyze a webpage",
		Buckets: prometheus.DefBuckets,
	})

	// AnalysisRequests tracks the total number of analysis requests
	AnalysisRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webpage_analysis_requests_total",
		Help: "Total number of webpage analysis requests",
	})

	// AnalysisErrors tracks the number of failed analyses
	AnalysisErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webpage_analysis_errors_total",
		Help: "Total number of webpage analysis errors",
	})

	// LinkCounts tracks various link metrics
	LinkCounts = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "webpage_link_counts",
		Help: "Counts of different types of links found on pages",
	}, []string{"type"})

	// HeadingCounts tracks the number of headings by level
	HeadingCounts = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "webpage_heading_counts",
		Help: "Counts of headings by level (h1, h2, etc)",
	}, []string{"level"})

	// LoginFormCount tracks the number of pages with login forms
	LoginFormCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "webpage_login_forms_total",
		Help: "Total number of pages with login forms",
	})

	// HTMLVersionCount tracks the distribution of HTML versions
	HTMLVersionCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "webpage_html_versions",
		Help: "Distribution of HTML versions found",
	}, []string{"version"})
)
