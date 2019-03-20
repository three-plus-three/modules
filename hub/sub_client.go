package hub

import (
	"sync/atomic"

	"github.com/three-plus-three/modules/websocket2"
)

type ErrDisconnect struct {
	err error
}

func (e *ErrDisconnect) Error() string {
	return e.err.Error()
}

func IsConnected(e error) bool {
	_, ok := e.(*ErrDisconnect)
	return ok
}

type Subscription struct {
	closed int32
	Conn   *websocket2.Conn
}

func (sub *Subscription) Close() error {
	if atomic.CompareAndSwapInt32(&sub.closed, 0, 1) {
		return sub.Conn.Close()
	}
	return nil
}

func (sub *Subscription) Run(cb func(*Subscription, Message)) error {
	for {
		var bs []byte
		err := websocket2.Message.Receive(sub.Conn, &bs)
		if err != nil {
			return err
		}

		cb(sub, CreateDataMessage(bs))
	}
}
