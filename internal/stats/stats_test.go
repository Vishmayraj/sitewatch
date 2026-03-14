package stats

import (
	"testing"
	"time"
)

func TestRecord_TracksSuccessesAndFailures(t *testing.T) {
	s := New()

	s.Record(true, 100*time.Millisecond)
	s.Record(true, 200*time.Millisecond)
	s.Record(false, 0)

	snap := s.Snapshot()

	if snap.Total != 3 {
		t.Errorf("expected total 3, got %d", snap.Total)
	}
	if snap.Successes != 2 {
		t.Errorf("expected 2 successes, got %d", snap.Successes)
	}
	if snap.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", snap.Failures)
	}
}

func TestSnapshot_UptimePercent(t *testing.T) {
	s := New()

	s.Record(true, 100*time.Millisecond)
	s.Record(true, 100*time.Millisecond)
	s.Record(true, 100*time.Millisecond)
	s.Record(false, 0)

	snap := s.Snapshot()

	expected := 75.0
	if snap.UptimePct != expected {
		t.Errorf("expected uptime %.1f%%, got %.1f%%", expected, snap.UptimePct)
	}
}

func TestSnapshot_LatencyMinMaxAvg(t *testing.T) {
	s := New()

	s.Record(true, 80*time.Millisecond)
	s.Record(true, 120*time.Millisecond)
	s.Record(true, 100*time.Millisecond)

	snap := s.Snapshot()

	if snap.MinLatency != 80*time.Millisecond {
		t.Errorf("expected min 80ms, got %v", snap.MinLatency)
	}
	if snap.MaxLatency != 120*time.Millisecond {
		t.Errorf("expected max 120ms, got %v", snap.MaxLatency)
	}
	if snap.AvgLatency != 100*time.Millisecond {
		t.Errorf("expected avg 100ms, got %v", snap.AvgLatency)
	}
}

func TestSnapshot_NoSuccesses_ZeroLatency(t *testing.T) {
	s := New()

	s.Record(false, 0)
	s.Record(false, 0)

	snap := s.Snapshot()

	if snap.MinLatency != 0 {
		t.Errorf("expected min latency 0 with no successes, got %v", snap.MinLatency)
	}
	if snap.AvgLatency != 0 {
		t.Errorf("expected avg latency 0 with no successes, got %v", snap.AvgLatency)
	}
}