package analyzer

import "fmt"

// AnalysisError represents an error that occurred during page analysis
type AnalysisError struct {
	Code    string
	Message string
	Err     error
}

func (e *AnalysisError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AnalysisError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrInvalidURL      = "INVALID_URL"
	ErrFetchFailed     = "FETCH_FAILED"
	ErrParseFailed     = "PARSE_FAILED"
	ErrTimeout         = "TIMEOUT"
	ErrMaxLinksReached = "MAX_LINKS_REACHED"
	ErrMaxDepthReached = "MAX_DEPTH_REACHED"
)

// NewAnalysisError creates a new AnalysisError
func NewAnalysisError(code, message string, err error) *AnalysisError {
	return &AnalysisError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
