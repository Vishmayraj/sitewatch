package pkg

import "time"

// Result holds the outcome of a single HTTP health check.
type Result struct {
    Timestamp  time.Time
    StatusCode int
    Duration   time.Duration
    Error      error
}

// IsSuccess returns true if the check had no error and a 2xx status code.
func (r Result) IsSuccess() bool {
    return r.Error == nil && r.StatusCode >= 200 && r.StatusCode < 300
}