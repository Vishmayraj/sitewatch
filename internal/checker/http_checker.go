package checker

import (
	"net/http"
	"time"
	"sitewatch/pkg"
)

// Checker performs HTTP health checks against a target URL.
type Checker struct {
	client *http.Client
	url    string
}

// New creates a Checker with a fixed timeout.
func New(url string, timeout time.Duration) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: timeout,
		},
		url: url,
	}
}

// Check performs a single HTTP GET and returns a Result.
func (c *Checker) Check() pkg.Result {
	start := time.Now()

	resp, err := c.client.Get(c.url)

	duration := time.Since(start)

	if err != nil {
		return pkg.Result{
			Timestamp: start,
			Duration:  duration,
			Error:     err,
		}
	}

	defer resp.Body.Close()

	return pkg.Result{
		Timestamp:  start,
		StatusCode: resp.StatusCode,
		Duration:   duration,
		Error:      nil,
	}
}