package server

import (
	"sync"
	"time"
)

type Topic struct {
	name          string
	capacity      int
	last_id       int
	channels      []*Consumer
	channels_lock sync.RWMutex
}

func (topic *Topic) Close() error {
	topic.channels_lock.Lock()
	channels := topic.channels
	topic.channels = nil
	topic.channels_lock.Unlock()

	for _, ch := range channels {
		ch.Close()
	}
	return nil
}

func (topic *Topic) Chan() chan<- Message {
	return nil
}

func (topic *Topic) Send(msg Message) error {
	topic.channels_lock.RLock()
	defer topic.channels_lock.RUnlock()

	for _, consumer := range topic.channels {
		select {
		case consumer.send <- msg:
			consumer.addSuccess()
		default:
			consumer.addDiscard()
		}
	}
	return nil
}

func (topic *Topic) SendTimeout(msg Message, timeout time.Duration) error {
	var channels []*Consumer

	var timer *time.Timer
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		defer timer.Stop()
	}

	func() {
		topic.channels_lock.RLock()
		defer topic.channels_lock.RUnlock()

		for _, consumer := range topic.channels {
			select {
			case consumer.send <- msg:
				consumer.addSuccess()
			default:
				channels = append(channels, consumer)
			}
		}
	}()

	if len(channels) == 0 {
		return nil
	}

	if timeout > 0 {
		return ErrPartialSend
	}

	for idx, consumer := range channels {
		select {
		case consumer.send <- msg:
			consumer.addSuccess()
		case <-timer.C:
			consumer.addDiscard()

			channels = channels[idx+1:]
			goto skip_ff
		}
	}
	return nil

skip_ff:
	for _, consumer := range channels {
		select {
		case consumer.send <- msg:
			consumer.addSuccess()
		default:
			consumer.addDiscard()
		}
	}
	return ErrPartialSend
}

func (topic *Topic) ListenOn() *Consumer {
	c := make(chan Message, topic.capacity)

	listener := &Consumer{Topic: topic, send: c, C: c}

	topic.channels_lock.Lock()
	topic.last_id++
	listener.id = topic.last_id
	topic.channels = append(topic.channels, listener)
	topic.channels_lock.Unlock()

	listener.closer = func() error {
		if nil != listener.Topic {
			listener.Topic.remove(listener.id)
			close(listener.send)
		}
		listener.Topic = nil
		return nil
	}
	return listener
}

func (topic *Topic) remove(id int) (ret *Consumer) {
	topic.channels_lock.Lock()
	for idx, consumer := range topic.channels {
		if consumer.id == id {
			ret = consumer

			copy(topic.channels[idx:], topic.channels[idx+1:])
			topic.channels = topic.channels[:len(topic.channels)-1]
			break
		}
	}
	topic.channels_lock.Unlock()
	return ret
}

func creatTopic(srv *Server, name string, capacity int) *Topic {
	return &Topic{name: name, capacity: capacity}
}
