package analyzer

import (
	"log/slog"
	"time"
)

// AnalyzerLogger implements the Logger interface
type AnalyzerLogger struct {
	log *slog.Logger
}

// NewAnalyzerLogger creates a new AnalyzerLogger
func NewAnalyzerLogger(log *slog.Logger) *AnalyzerLogger {
	return &AnalyzerLogger{log: log}
}

// LogAnalysisStart logs the start of a page analysis
func (l *AnalyzerLogger) LogAnalysisStart(url string) {
	l.log.Info("starting analysis",
		slog.String("url", url),
		slog.Time("start_time", time.Now()),
	)
}

// LogAnalysisError logs an error during analysis
func (l *AnalyzerLogger) LogAnalysisError(err error, url string) {
	l.log.Error("analysis failed",
		slog.String("url", url),
		slog.String("error", err.Error()),
	)
}

// LogAnalysisComplete logs the completion of analysis
func (l *AnalyzerLogger) LogAnalysisComplete(url string, duration float64) {
	l.log.Info("analysis complete",
		slog.String("url", url),
		slog.Float64("duration_seconds", duration),
	)
}

// LogLinkCheck logs the result of a link check
func (l *AnalyzerLogger) LogLinkCheck(url string, isAccessible bool) {
	l.log.Debug("link check result",
		slog.String("url", url),
		slog.Bool("is_accessible", isAccessible),
	)
}

// LogDebug logs a debug message
func (l *AnalyzerLogger) LogDebug(msg string, args ...interface{}) {
	l.log.Debug(msg, args...)
}
