package pkg

import (
	"fmt"
    "testing"
    "time"
)

func TestIsSuccess_WithOKStatus(t *testing.T) {
    r := Result{
        Timestamp:  time.Now(),
        StatusCode: 200,
        Duration:   102 * time.Millisecond,
        Error:      nil,
    }

    if !r.IsSuccess() {
        t.Errorf("expected 200 with no error to be a success")
    }
}

func TestIsSuccess_WithError(t *testing.T) {
    r := Result{
        Timestamp:  time.Now(),
        StatusCode: 0,
        Duration:   0,
        Error:      fmt.Errorf("connection refused"),
    }

    if r.IsSuccess() {
        t.Errorf("expected result with error to not be a success")
    }
}

func TestIsSuccess_With500Status(t *testing.T) {
    r := Result{
        Timestamp:  time.Now(),
        StatusCode: 500,
        Duration:   80 * time.Millisecond,
        Error:      nil,
    }

    if r.IsSuccess() {
        t.Errorf("expected 500 status to not be a success")
    }
}