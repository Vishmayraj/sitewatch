package checker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCheck_Success spins up a local server and checks we get a 200 result.
func TestCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New(server.URL, 5*time.Second)
	result := c.Check()

	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected 200, got %d", result.StatusCode)
	}
	if result.Duration <= 0 {
		t.Errorf("expected positive duration, got %v", result.Duration)
	}
}

// TestCheck_Timeout checks that a hanging server triggers a timeout error.
func TestCheck_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // hang
	}))
	defer server.Close()

	c := New(server.URL, 100*time.Millisecond) // very short timeout
	result := c.Check()

	if result.Error == nil {
		t.Errorf("expected timeout error, got nil")
	}
}

// TestCheck_Non200 checks that a 500 response is captured correctly.
func TestCheck_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := New(server.URL, 5*time.Second)
	result := c.Check()

	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}
	if result.StatusCode != 500 {
		t.Errorf("expected 500, got %d", result.StatusCode)
	}
	if result.IsSuccess() {
		t.Errorf("expected IsSuccess() to be false for 500")
	}
}