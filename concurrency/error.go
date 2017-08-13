package concurrency

import "sync/atomic"

type syncError struct {
	err error
}

type ErrorValue struct {
	value atomic.Value
}

func (ev *ErrorValue) Set(e error) {
	ev.value.Store(&syncError{err: e})
}

func (ev *ErrorValue) Get() error {
	o := ev.value.Load()
	if o == nil {
		return nil
	}
	if e, ok := o.(*syncError); ok {
		return e.err
	}
	return nil
}
