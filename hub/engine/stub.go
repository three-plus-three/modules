package engine

import (
	"io"
	"log"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"github.com/three-plus-three/modules/hub"

	"golang.org/x/net/websocket"
)

type engineStub struct {
	srv        websocket.Server
	disconnect DisconnectFunc

	createdAt  time.Time
	remoteAddr string
	mode       string
	role       string
	client     string
	name       string
	conn       *websocket.Conn
	logger     *log.Logger

	c      chan struct{}
	closed int32
}

func (stub *engineStub) Close() error {
	if atomic.CompareAndSwapInt32(&stub.closed, 0, 1) {
		if stub.disconnect != nil {
			stub.disconnect()
		}
		close(stub.c)

		stub.conn.Close()
		stub.conn = nil
	}
	return nil
}

func (stub *engineStub) Info() map[string]interface{} {
	return map[string]interface{}{
		"remoteAddr": stub.remoteAddr,
		"mode":       stub.mode,
		"role":       stub.role,
		"client":     stub.client,
		"name":       stub.name,
		"createdAt":  stub.createdAt,
	}
}

func (stub *engineStub) subscribe(consumer *Consumer) {
	is_running := true
	for is_running {
		select {
		case msg, ok := <-consumer.C:
			if !ok {
				is_running = false
				stub.logger.Println("[", stub.client, "] connection(write:", stub.remoteAddr, ") is closed - queue is shutdown.")
				break
			}

			if e := websocket.Message.Send(stub.conn, msg.Bytes()); nil != e {
				is_running = false

				if !consumer.Unread(msg) {
					stub.logger.Println("[", stub.client, "] message is missing on the connection(write:", stub.remoteAddr, ") is .")
				}

				if strings.Contains(e.Error(), "use of closed network connection") {
					stub.logger.Println("[", stub.client, "] connection(write:", stub.remoteAddr, ") is closed.")
				} else {
					stub.logger.Println("[", stub.client, "] connection(write:", stub.remoteAddr, ") is closed -", e)
				}
				break
			}
		case <-stub.c:
			is_running = false
			stub.logger.Println("[", stub.client, "] connection(write:", stub.remoteAddr, ") is closed - queue is shutdown.")
		}
	}
}

func (stub *engineStub) publish(producer chan<- hub.Message) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	const trySendCount = 2
	isRunning := true
	for isRunning {
		var data []byte
		if e := websocket.Message.Receive(stub.conn, &data); nil != e {
			if e == io.EOF {
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed - peer is shutdown.")
			} else if strings.Contains(e.Error(), "use of closed network connection") {
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed.")
			} else {
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed -", e)
			}
			isRunning = false
			continue
		}

		msg := hub.CreateDataMessage(data)

		continueTick := 0
		for continueTick < trySendCount {
			select {
			case producer <- msg:
				continueTick = math.MaxInt32
			case <-ticker.C:
				continueTick++
				if continueTick >= trySendCount {
					stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed - queue is overflow.")
				}
			case <-stub.c:
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed - queue is shutdown.")

				continueTick = math.MaxInt32
				isRunning = false
			}
		}
	}
}

func (stub *engineStub) sendToTopic(producer Producer) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		var data []byte
		if e := websocket.Message.Receive(stub.conn, &data); nil != e {
			if e == io.EOF {
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed - peer is shutdown.")
			} else if strings.Contains(e.Error(), "use of closed network connection") {
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed.")
			} else {
				stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is closed -", e)
			}
			break
		}

		msg := hub.CreateDataMessage(data)
		rs, err := producer.SendWithContext(msg, ticker.C)
		if err != hub.ErrPartialSend {
			stub.logger.Println("[", stub.client, "] connection(read:", stub.remoteAddr, ") is fail -", err)
		} else {
			rs.SendWithContext(msg, ticker.C)
		}

		rs.Close()
	}
}
