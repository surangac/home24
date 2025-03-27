package analyzer

// LinkInfo represents information about a link found in the page
type LinkInfo struct {
	URL        string
	IsInternal bool
}

// AnalysisResult represents the complete analysis of a webpage
type AnalysisResult struct {
	URL             string
	Title           string
	Headings        map[string]int
	Links           []LinkInfo
	AccessibleLinks int
	HasLoginForm    bool
	HTMLVersion     string
}
