package analyzer

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// DefaultLinkChecker implements the LinkChecker interface
type DefaultLinkChecker struct {
	client *http.Client
	log    Logger
	config *AnalyzerConfig
}

// NewDefaultLinkChecker creates a new DefaultLinkChecker
func NewDefaultLinkChecker(client *http.Client, log Logger, config *AnalyzerConfig) *DefaultLinkChecker {
	return &DefaultLinkChecker{
		client: client,
		log:    log,
		config: config,
	}
}

// CheckAccessibility checks if a link is accessible
func (c *DefaultLinkChecker) CheckAccessibility(ctx context.Context, urlStr string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
	if err != nil {
		c.log.LogDebug("Failed to create request for link", "link", urlStr, "error", err)
		return false
	}

	req.Header.Set("User-Agent", c.config.UserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.LogDebug("Failed to check link", "link", urlStr, "error", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// CheckWithRetry checks a link with retry logic
func (c *DefaultLinkChecker) CheckWithRetry(ctx context.Context, urlStr string) bool {
	var isAccessible bool
	for i := 0; i < c.config.RetryAttempts; i++ {
		isAccessible = c.CheckAccessibility(ctx, urlStr)
		if isAccessible {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(time.Duration(1<<uint(i)) * time.Second):
			continue
		}
	}
	return isAccessible
}

// CheckLinksConcurrently checks multiple links concurrently
func (c *DefaultLinkChecker) CheckLinksConcurrently(ctx context.Context, links []LinkInfo) map[string]bool {
	results := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, c.config.MaxConcurrentLinks)

	for _, link := range links {
		wg.Add(1)
		go func(l LinkInfo) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			isAccessible := c.CheckAccessibility(ctx, l.URL)
			mu.Lock()
			results[l.URL] = isAccessible
			mu.Unlock()
			c.log.LogLinkCheck(l.URL, isAccessible)
		}(link)
	}

	wg.Wait()
	return results
}

// RetryWithBackoff retries a function with exponential backoff
func (c *DefaultLinkChecker) RetryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := 0; i < c.config.RetryAttempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			backoff := time.Duration(1<<uint(i)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
		return nil
	}
	return lastErr
}
