package stats

import (
	"math"
	"sync"
	"time"
)

// Stats tracks running statistics across all checks.
type Stats struct {
	mu           sync.Mutex
	total        int
	successes    int
	failures     int
	totalLatency time.Duration
	minLatency   time.Duration
	maxLatency   time.Duration
}

// New creates a fresh Stats instance.
func New() *Stats {
	return &Stats{
		minLatency: time.Duration(math.MaxInt64),
	}
}

// Record adds a single result to the stats.
func (s *Stats) Record(success bool, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.total++

	if success {
		s.successes++
		s.totalLatency += duration

		if duration < s.minLatency {
			s.minLatency = duration
		}
		if duration > s.maxLatency {
			s.maxLatency = duration
		}
	} else {
		s.failures++
	}
}

// Summary holds a snapshot of current statistics.
type Summary struct {
	Total      int
	Successes  int
	Failures   int
	UptimePct  float64
	AvgLatency time.Duration
	MinLatency time.Duration
	MaxLatency time.Duration
}

// Snapshot returns a consistent point-in-time summary.
func (s *Stats) Snapshot() Summary {
	s.mu.Lock()
	defer s.mu.Unlock()

	var uptime float64
	var avg time.Duration

	if s.total > 0 {
		uptime = float64(s.successes) / float64(s.total) * 100
	}
	if s.successes > 0 {
		avg = s.totalLatency / time.Duration(s.successes)
	}

	min := s.minLatency
	if s.successes == 0 {
		min = 0
	}

	return Summary{
		Total:      s.total,
		Successes:  s.successes,
		Failures:   s.failures,
		UptimePct:  uptime,
		AvgLatency: avg,
		MinLatency: min,
		MaxLatency: s.maxLatency,
	}
}