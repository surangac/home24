package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AnalysisDuration tracks how long each analysis takes
	AnalysisDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "webpage_analyzer_analysis_duration_seconds",
		Help:    "How long the webpage analysis took in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
	})

	// AnalysisRequests counts the number of analysis requests
	AnalysisRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webpage_analyzer_requests_total",
		Help: "The total number of webpage analysis requests",
	})

	// AnalysisErrors counts the number of analysis errors
	AnalysisErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webpage_analyzer_errors_total",
		Help: "The total number of webpage analysis errors",
	})

	// LinkCounts tracks different types of links found
	LinkCounts = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "webpage_analyzer_link_counts",
		Help: "Number of different types of links found",
	}, []string{"type"})

	// HeadingCounts tracks the number of different heading levels
	HeadingCounts = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "webpage_analyzer_heading_counts",
		Help: "Number of different heading levels found",
	}, []string{"level"})

	// LoginFormCount tracks the number of login forms found
	LoginFormCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webpage_analyzer_login_forms_total",
		Help: "The total number of login forms found",
	})

	// HTMLVersionCount tracks the number of different HTML versions
	HTMLVersionCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webpage_analyzer_html_versions_total",
		Help: "The total number of different HTML versions encountered",
	}, []string{"version"})
)
