package hub

import "errors"

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

func (msg Message) Data() []byte {
	return msg
}

func CreateDataMessage(bs []byte) Message {
	return Message(bs)
}
