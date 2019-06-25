package waiter

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoll(t *testing.T) {
	var n uint64
	op := func() (bool, error) {
		return atomic.AddUint64(&n, 1) == 10, nil
	}

	ctx, _ := context.WithTimeout(context.TODO(), 500*time.Second)
	ctx = WithInterval(ctx, 200*time.Millisecond)
	if err := Poll(ctx, op); err != nil {
		t.Fatal(err)
	}

	if n != 10 {
		t.Errorf("Expected 10 but was %v", n)
	}
}
