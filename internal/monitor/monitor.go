package monitor

import (
	"context"
	"fmt"
	"log"
	"os"
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
	logger   *log.Logger
}

// New creates a Monitor. If logPath is empty, file logging is disabled.
func New(url string, interval time.Duration, timeout time.Duration, logPath string) *Monitor {
	m := &Monitor{
		checker:  checker.New(url, timeout),
		stats:    stats.New(),
		interval: interval,
	}

	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not open log file: %v\n", err)
		} else {
			m.logger = log.New(f, "", 0)
		}
	}

	return m
}

// Run starts the monitoring loop. It blocks until the context is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	results := make(chan pkg.Result)

	for {
		select {
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
			m.logResult(result)
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

// logResult writes a single check result to the log file, if logging is enabled.
func (m *Monitor) logResult(r pkg.Result) {
	if m.logger == nil {
		return
	}

	timestamp := r.Timestamp.Format(time.RFC3339)

	if r.Error != nil {
		m.logger.Printf("%s ERROR %v", timestamp, r.Error)
		return
	}

	m.logger.Printf("%s %d %dms", timestamp, r.StatusCode, r.Duration.Milliseconds())
}

// output writes a line to stdout and to the log file if enabled.
func (m *Monitor) output(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
	if m.logger != nil {
		m.logger.Printf(format, args...)
	}
}

// printFinalStats prints the summary when monitoring stops.
func (m *Monitor) printFinalStats() {
	snap := m.stats.Snapshot()

	m.output("\n--- Monitoring stopped ---")
	m.output("Total checks : %d", snap.Total)
	m.output("Successes    : %d", snap.Successes)
	m.output("Failures     : %d", snap.Failures)
	m.output("Uptime       : %.2f%%", snap.UptimePct)

	if snap.Successes > 0 {
		m.output("Latency min  : %dms", snap.MinLatency.Milliseconds())
		m.output("Latency avg  : %dms", snap.AvgLatency.Milliseconds())
		m.output("Latency max  : %dms", snap.MaxLatency.Milliseconds())
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