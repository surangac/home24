package analyzer

import (
	"home24/internal/metrics"
)

// PrometheusMetricsCollector implements the MetricsCollector interface
type PrometheusMetricsCollector struct {
	config AnalyzerConfig
}

// NewPrometheusMetricsCollector creates a new PrometheusMetricsCollector
func NewPrometheusMetricsCollector(config AnalyzerConfig) *PrometheusMetricsCollector {
	return &PrometheusMetricsCollector{
		config: config,
	}
}

// RecordDuration records the duration of an analysis
func (m *PrometheusMetricsCollector) RecordDuration(duration float64) {
	metrics.AnalysisDuration.Observe(duration)
}

// RecordResults records the results of an analysis
func (m *PrometheusMetricsCollector) RecordResults(result *AnalysisResult) {
	// Record link counts
	for _, link := range result.Links {
		if link.IsInternal {
			metrics.LinkCounts.WithLabelValues("internal").Inc()
		} else {
			metrics.LinkCounts.WithLabelValues("external").Inc()
		}
	}

	// Record heading counts
	for level, count := range result.Headings {
		metrics.HeadingCounts.WithLabelValues(level).Set(float64(count))
	}

	// Record login form count
	if result.HasLoginForm {
		metrics.LoginFormCount.Inc()
	}

	// Record HTML version
	metrics.HTMLVersionCount.WithLabelValues(result.HTMLVersion).Inc()
}

// RecordError records an error during analysis
func (m *PrometheusMetricsCollector) RecordError(err error) {
	metrics.AnalysisErrors.Inc()
}

// RecordRequest records a new analysis request
func (m *PrometheusMetricsCollector) RecordRequest() {
	metrics.AnalysisRequests.Inc()
}
