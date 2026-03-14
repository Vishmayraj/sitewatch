package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitor_RunsAndStops(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	m := New(server.URL, 100*time.Millisecond, 5*time.Second, "")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		m.Run(ctx)
		close(done)
	}()

	// Let it run for a few checks
	time.Sleep(450 * time.Millisecond)
	cancel()

	// Wait for Run to return
	select {
	case <-done:
		// good
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not stop after context cancellation")
	}

	snap := m.stats.Snapshot()
	if snap.Total < 3 {
		t.Errorf("expected at least 3 checks, got %d", snap.Total)
	}
	if snap.Successes < 3 {
		t.Errorf("expected at least 3 successes, got %d", snap.Successes)
	}
}

func TestMonitor_RecordsFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	m := New(server.URL, 100*time.Millisecond, 5*time.Second, "")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		m.Run(ctx)
		close(done)
	}()

	time.Sleep(250 * time.Millisecond)
	cancel()
	<-done

	snap := m.stats.Snapshot()
	if snap.Failures < 2 {
		t.Errorf("expected at least 2 failures, got %d", snap.Failures)
	}
}