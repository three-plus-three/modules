package hub

import "golang.org/x/net/websocket"

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
	closed bool
	Conn   *websocket.Conn
}

func (sub *Subscription) Stop() error {
	if sub.closed {
		return nil
	}

	sub.closed = true
	return sub.Conn.Close()
}

func (sub *Subscription) Run(cb func(*Subscription, []byte)) error {
	for {
		var bs []byte
		err := websocket.Message.Receive(sub.Conn, &bs)
		if err != nil {
			return err
		}

		cb(sub, bs)
	}
	return nil
}
