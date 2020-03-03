package waiter

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoll(t *testing.T) {
	var n uint64
	op := func() (bool, error) {
		return atomic.AddUint64(&n, 1) == 10, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()
	if err := Poll(ctx, 200*time.Millisecond, op); err != nil {
		t.Fatal(err)
	}

	if n != 10 {
		t.Errorf("Expected 10 but was %v", n)
	}
}

func TestPollError(t *testing.T) {
	var n uint64
	op := func() (bool, error) {
		i := atomic.AddUint64(&n, 1)
		if i == 5 {
			return false, fmt.Errorf("weird number: %d", i)
		}
		return i == 10, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	err := Poll(ctx, 200*time.Millisecond, op)

	if err == nil {
		t.Fatal("Expected an error but none returned")
	}

	if err.Error() != "weird number: 5" {
		t.Errorf("Unexpected error: %s", err)
	}

	if n != 5 {
		t.Errorf("Expected 10 but was %v", n)
	}
}

func TestPollTimeout(t *testing.T) {
	var n uint64
	op := func() (bool, error) {
		return atomic.AddUint64(&n, 1) == 100, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := Poll(ctx, 200*time.Millisecond, op)

	if err == nil {
		t.Fatal("Expected an error but none returned")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestPollCancelled(t *testing.T) {
	var n uint64
	op := func() (bool, error) {
		return atomic.AddUint64(&n, 1) == 100, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()
	err := Poll(ctx, 1*time.Second, op)

	if err == nil {
		t.Fatal("Expected an error but none returned")
	}

	if err != context.Canceled {
		t.Errorf("Unexpected error: %s", err)
	}
}
