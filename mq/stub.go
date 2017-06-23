package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

func checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	return err
}

type engineStub struct {
	srv        websocket.Server
	disconnect DisconnectFunc

	createdAt  time.Time
	remoteAddr string
	mode       string
	role       string
	name       string
	conn       *websocket.Conn

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
		"name":       stub.name,
		"createdAt":  stub.createdAt,
	}
}

func (stub *engineStub) subscribe(consumer *Consumer) {
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString("[panic] [broker] connection(write: ")
			buffer.WriteString(stub.remoteAddr)
			buffer.WriteString(") \r\n")
			buffer.Write(debug.Stack())
			log.Println(buffer.String())
		}
		stub.conn.Close()
	}()

	is_running := true
	for is_running {
		select {
		case msg, ok := <-consumer.C:
			if !ok {
				is_running = false
				log.Println("[broker] connection(write:", stub.remoteAddr, ") is closed - queue is shutdown.")
				break
			}

			if e := websocket.Message.Send(stub.conn, msg.Bytes()); nil != e {
				is_running = false

				if !consumer.Unread(msg) {
					log.Println("[broker] message is missing on the connection(write:", stub.remoteAddr, ") is .")
				}

				if strings.Contains(e.Error(), "use of closed network connection") {
					log.Println("[broker] connection(write:", stub.remoteAddr, ") is closed.")
				} else {
					log.Println("[broker] connection(write:", stub.remoteAddr, ") is closed -", e)
				}
				break
			}
		case <-stub.c:
			is_running = false
			log.Println("[broker] connection(write:", stub.remoteAddr, ") is closed - queue is shutdown.")
		}
	}
}

func (stub *engineStub) publish(producer chan<- Message) {
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString("[panic] [broker] connection(write: ")
			buffer.WriteString(stub.remoteAddr)
			buffer.WriteString(") \r\n")
			buffer.Write(debug.Stack())
			log.Println(buffer.String())
		}
		stub.conn.Close()
	}()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	const trySendCount = 2
	is_running := true
	for is_running {
		var data []byte
		if e := websocket.Message.Receive(stub.conn, &data); nil != e {
			if e == io.EOF {
				log.Println("[broker] connection(read:", stub.remoteAddr, ") is closed - peer is shutdown.")
			} else if strings.Contains(e.Error(), "use of closed network connection") {
				log.Println("[broker] connection(read:", stub.remoteAddr, ") is closed.")
			} else {
				log.Println("[broker] connection(read:", stub.remoteAddr, ") is closed -", e)
			}
			is_running = false
			break
		}

		continueTick := 0
		for continueTick < trySendCount {
			select {
			case producer <- CreateDataMessage(data):
				continueTick = math.MaxInt32
			case <-ticker.C:
				continueTick++
				if continueTick >= trySendCount {
					log.Println("[broker] connection(write:", stub.remoteAddr, ") is closed - queue is overflow.")
				}
			case <-stub.c:
				log.Println("[broker] connection(write:", stub.remoteAddr, ") is closed - queue is shutdown.")

				continueTick = math.MaxInt32
				is_running = false
			}
		}
	}
}
