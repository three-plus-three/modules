package mq

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

func (topic *Topic) SendWithContext(msg Message, ctx <-chan time.Time) (*RetrySender, error) {
	var channels []*Consumer
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
		return nil, nil
	}

	var rs = &RetrySender{consumers: channels}
	if ctx == nil {
		return rs, ErrPartialSend
	}
	err := rs.send(msg, ctx)
	return rs, err
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

func creatTopic(srv *Core, name string, capacity int) *Topic {
	return &Topic{name: name, capacity: capacity}
}
