package concurrency

import (
	"sync/atomic"
)

type errorWrapper struct {
	err error
}

type ErrorValue struct {
	value atomic.Value
}

func (ev *ErrorValue) Set(e error) {
	ev.value.Store(&errorWrapper{err: e})
}

func (ev *ErrorValue) Get() error {
	o := ev.value.Load()
	if o == nil {
		return nil
	}
	if e, ok := o.(*errorWrapper); ok {
		return e.err
	}
	return nil
}
