package monitor

import (
	"context"
	"fmt"
	"time"

	"sitewatch/internal/checker"
	"sitewatch/internal/stats"
	"sitewatch/pkg"
)

// Monitor orchestrates periodic health checks against a URL.
type Monitor struct {
	checker  *checker.Checker
	stats    *stats.Stats
	interval time.Duration
}

// New creates a Monitor.
func New(url string, interval time.Duration, timeout time.Duration) *Monitor {
	return &Monitor{
		checker:  checker.New(url, timeout),
		stats:    stats.New(),
		interval: interval,
	}
}

// Run starts the monitoring loop. It blocks until the context is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	results := make(chan pkg.Result)

	for {
		select {
		// Graceful closing
		case <-ctx.Done():
			m.printFinalStats()
			return

		case <-ticker.C:
			go func() {
				results <- m.checker.Check()
			}()

		case result := <-results:
			m.stats.Record(result.IsSuccess(), result.Duration)
			m.printResult(result)
		}
	}
}

// printResult prints a single check result to stdout.
func (m *Monitor) printResult(r pkg.Result) {
	timestamp := r.Timestamp.Format("15:04:05")

	if r.Error != nil {
		fmt.Printf("[%s] ERROR      %v\n", timestamp, r.Error)
		return
	}

	fmt.Printf("[%s] %d %s  %dms\n",
		timestamp,
		r.StatusCode,
		statusText(r.StatusCode),
		r.Duration.Milliseconds(),
	)
}

// printFinalStats prints the summary when monitoring stops.
func (m *Monitor) printFinalStats() {
	snap := m.stats.Snapshot()

	fmt.Println("\n--- Monitoring stopped ---")
	fmt.Printf("Total checks : %d\n", snap.Total)
	fmt.Printf("Successes    : %d\n", snap.Successes)
	fmt.Printf("Failures     : %d\n", snap.Failures)
	fmt.Printf("Uptime       : %.2f%%\n", snap.UptimePct)

	if snap.Successes > 0 {
		fmt.Printf("Latency min  : %dms\n", snap.MinLatency.Milliseconds())
		fmt.Printf("Latency avg  : %dms\n", snap.AvgLatency.Milliseconds())
		fmt.Printf("Latency max  : %dms\n", snap.MaxLatency.Milliseconds())
	}
}

// statusText returns a short string for common HTTP status codes.
func statusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 301:
		return "Moved"
	case 302:
		return "Found"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return "Unknown"
	}
}