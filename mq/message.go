package server

import (
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrTimeout           = errors.New("timeout")
	ErrAlreadyClosed     = errors.New("already closed.")
	ErrMoreThanMaxRead   = errors.New("more than maximum read.")
	ErrUnexceptedMessage = errors.New("recv a unexcepted message.")
	ErrUnexceptedAck     = errors.New("recv a unexcepted ack message.")
	ErrEmptyString       = errors.New("empty error message.")
	ErrMagicNumber       = errors.New("magic number is error.")
	ErrLengthExceed      = errors.New("message length is exceed.")
	ErrLengthNotDigit    = errors.New("length field of message isn't number.")
	ErrQueueFull         = errors.New("queue is full.")
	ErrPartialSend       = errors.New("send is partial consumer.")
	ErrHandlerType       = errors.New("handler isn't http.Handler.")
)

// Message - 一个消息的数据
type Message []byte

func (msg Message) Bytes() []byte {
	return msg
}
func CreateDataMessage(bs []byte) Message {
	return Message(bs)
}

type Producer interface {
	Chan() chan<- Message
	Send(msg Message) error
	SendTimeout(msg Message, timeout time.Duration) error
}

type Consumer struct {
	success, discard uint64
	id               int
	Topic            *Topic
	send             chan Message
	C                <-chan Message
	closed           int32
	closer           func() error
}

func (consumer *Consumer) Success() uint64 {
	return atomic.LoadUint64(&consumer.success)
}

func (consumer *Consumer) Discard() uint64 {
	return atomic.LoadUint64(&consumer.discard)
}

func (consumer *Consumer) addSuccess() {
	atomic.AddUint64(&consumer.success, 1)
}

func (consumer *Consumer) addDiscard() {
	atomic.AddUint64(&consumer.discard, 1)
}

func (consumer *Consumer) Close() error {
	if atomic.CompareAndSwapInt32(&consumer.closed, 0, 1) {
		if consumer.closer != nil {
			return consumer.closer()
		}
	}
	return nil
}
