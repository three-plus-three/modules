package engine

import (
	"sync/atomic"
	"time"

	"github.com/three-plus-three/modules/hub"
)

type Queue struct {
	name     string
	C        chan hub.Message
	consumer Consumer

	closed int32
}

func (q *Queue) ListenOn() *Consumer {
	return &q.consumer
}

func (q *Queue) Close() error {
	if atomic.CompareAndSwapInt32(&q.closed, 0, 1) {
		close(q.C)
		for range q.C {
		}
	}

	return q.consumer.Close()
}

func (q *Queue) Send(msg hub.Message) error {
	q.C <- msg
	q.consumer.addSuccess()
	return nil
}

func (q *Queue) SendWithContext(msg hub.Message, ctx <-chan time.Time) (*RetrySender, error) {
	select {
	case q.C <- msg:
		q.consumer.addSuccess()
		return nil, nil
	case <-ctx:
		q.consumer.addDiscard()
		return nil, hub.ErrTimeout
	}
}

func creatQueue(srv *Core, name string, capacity int) *Queue {
	c := make(chan hub.Message, capacity)
	q := &Queue{name: name, C: c, consumer: Consumer{C: c, send: c}}

	q.consumer.closer = func() error {
		return nil
	}
	return q
}
