package check

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

type Operation wait.ConditionFunc

func Noop() Operation {
	return func() (done bool, err error) {
		return true, nil
	}
}

type Waiter struct {
	Interval time.Duration
	Timeout  time.Duration
}

func (w Waiter) Wait(op Operation) error {
	return wait.Poll(w.Interval, w.Timeout, wait.ConditionFunc(op))
}

func (w Waiter) WaitAll(ops []Operation) error {
	for _, op := range ops {
		if err := w.Wait(op); err != nil {
			return err
		}
	}
	return nil
}
