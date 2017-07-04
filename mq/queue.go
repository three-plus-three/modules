package server

import "time"

type Queue struct {
	name     string
	C        chan Message
	consumer Consumer
}

func (self *Queue) ListenOn() *Consumer {
	return &self.consumer
}

func (self *Queue) Close() error {
	return self.consumer.Close()
}

func (q *Queue) Send(msg Message) error {
	q.C <- msg
	q.consumer.addSuccess()
	return nil
}

func (q *Queue) SendWithContext(msg Message, ctx <-chan time.Time) (*RetrySender, error) {
	select {
	case q.C <- msg:
		q.consumer.addSuccess()
		return nil, nil
	case <-ctx:
		q.consumer.addDiscard()
		return nil, ErrTimeout
	}
}

func creatQueue(srv *Core, name string, capacity int) *Queue {
	c := make(chan Message, capacity)
	q := &Queue{name: name, C: c, consumer: Consumer{C: c, send: c}}

	q.consumer.closer = func() error {
		close(q.C)
		for range q.C {
		}
		return nil
	}
	return q
}
