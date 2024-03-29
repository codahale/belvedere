package waiter

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codahale/gubbins/assert"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestPoll(t *testing.T) {
	t.Parallel()

	n := uint64(0)
	op := func() (bool, error) {
		return atomic.AddUint64(&n, 1) == 10, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	if err := Poll(ctx, 200*time.Millisecond, op); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Poll count", uint64(10), n)
}

func TestPollError(t *testing.T) {
	t.Parallel()

	want := io.EOF
	n := uint64(0)
	op := func() (bool, error) {
		i := atomic.AddUint64(&n, 1)
		if i == 5 {
			return false, want
		}

		return i == 10, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	err := Poll(ctx, 200*time.Millisecond, op)

	assert.Equal(t, "Poll error", want, err, cmpopts.EquateErrors())
	assert.Equal(t, "Poll count", uint64(5), n)
}

func TestPollTimeout(t *testing.T) {
	t.Parallel()

	n := uint64(0)
	op := func() (bool, error) {
		return atomic.AddUint64(&n, 1) == 100, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := Poll(ctx, 200*time.Millisecond, op)

	assert.Equal(t, "Poll timeout error", context.DeadlineExceeded, err, cmpopts.EquateErrors())
}

func TestPollCancelled(t *testing.T) {
	t.Parallel()

	n := uint64(0)
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

	assert.Equal(t, "Poll timeout error", context.Canceled, err, cmpopts.EquateErrors())
}
