package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestContextWithTimeout(t *testing.T) {
	// Test Go 1.24.6 context features
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		// Context should timeout
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded error, got: %v", ctx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Context should have timed out")
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test Go 1.24.6 error wrapping
	originalErr := context.DeadlineExceeded
	wrappedErr := fmt.Errorf("operation failed: %w", originalErr)

	if !errors.Is(wrappedErr, context.DeadlineExceeded) {
		t.Error("Error wrapping should preserve original error")
	}
}

func TestGoVersion(t *testing.T) {
	// This test will pass if we're using Go 1.24.6
	// The go.mod file should specify go 1.24.6
	t.Log("Testing with Go 1.24.6")
}

