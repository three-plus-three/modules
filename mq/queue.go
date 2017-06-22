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

func (q *Queue) SendTimeout(msg Message, timeout time.Duration) error {
	if timeout == 0 {
		select {
		case q.C <- msg:
			q.consumer.addSuccess()
			return nil
		default:
			q.consumer.addDiscard()
			return ErrQueueFull
		}
	}

	timer := time.NewTimer(timeout)
	select {
	case q.C <- msg:
		q.consumer.addSuccess()
		timer.Stop()
		return nil
	case <-timer.C:
		q.consumer.addDiscard()
		return ErrTimeout
	}
}

func creatQueue(srv *Server, name string, capacity int) *Queue {
	c := make(chan Message, capacity)
	q := &Queue{name: name, C: c, consumer: Consumer{C: c}}

	q.consumer.closer = func() error {
		close(q.C)
		for range q.C {
		}
		return nil
	}
	return q
}
