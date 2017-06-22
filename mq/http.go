package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

func getRealIP(conn *websocket.Conn) string {
	if nil != conn.Request() {
		address := conn.Request().Header.Get("X-Real-IP")
		if "" == address {
			address = conn.Request().Header.Get("X-Forwarded-For")
			if "" == address {
				address = conn.Request().RemoteAddr
			}
		}
		port := conn.Request().Header.Get("X-Real-Port")
		if "" != port {
			address += (":" + port)
		}
		return address
	} else {
		return "unknow"
	}
}

type StandardEngine struct {
	Server  *Server
	NoRoute http.Handler
}

func (se *StandardEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/queues", "/queues/":
		se.queuesIndex(w, r)
	case "/topics", "/topics/":
		se.topicsIndex(w, r)
	case "/clients", "/clients/":
		se.clientsIndex(w, r)
	default:
		if strings.HasPrefix(r.URL.Path, "/queues/") {
			urlPath := strings.TrimPrefix(r.URL.Path, "/queues/")
			if "" == urlPath {
				se.queuesIndex(w, r)
				return
			}

			switch r.Method {
			case "GET":
				se.doGet(w, r, urlPath,
					func(name string) *Consumer {
						return se.Server.CreateQueueIfNotExists(name).ListenOn()
					})
			case "POST", "PUT":
				se.doPost(w, r, urlPath,
					func(name string) Producer {
						return se.Server.CreateQueueIfNotExists(name)
					})
			default:
				if nil != r.Body {
					io.Copy(ioutil.Discard, r.Body)
					r.Body.Close()
				}

				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte("Method must is PUT or GET."))
			}

		} else if strings.HasPrefix(r.URL.Path, "/topics/") {
			urlPath := strings.TrimPrefix(r.URL.Path, "/topics/")
			if "" == urlPath {
				se.topicsIndex(w, r)
				return
			}

			switch r.Method {
			case "GET":
				se.doGet(w, r, urlPath,
					func(name string) *Consumer {
						return se.Server.CreateTopicIfNotExists(name).ListenOn()
					})
			case "POST", "PUT":
				se.doPost(w, r, urlPath,
					func(name string) Producer {
						return se.Server.CreateTopicIfNotExists(name)
					})
			default:
				if nil != r.Body {
					io.Copy(ioutil.Discard, r.Body)
					r.Body.Close()
				}

				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte("Method must is PUT or GET."))
			}
		} else {
			se.NoRoute.ServeHTTP(w, r)
		}
	}
}

func readMore(c <-chan Message, msg Message) []Message {
	results := append(make([]Message, 0, 12), msg)
	for i := 0; i < 100; i++ {
		select {
		case m, ok := <-c:
			if !ok {
				return results
			}
			results = append(results, m)
		default:
			return results
		}
	}
	return results
}

func (se *StandardEngine) doGet(w http.ResponseWriter, r *http.Request,
	urlPath string, cb func(name string) *Consumer) {
	query_params := r.URL.Query()

	timeout := GetTimeout(query_params, 1*time.Second)
	timer := time.NewTimer(timeout)
	consumer := cb(urlPath)
	defer consumer.Close()

	select {
	case msg, ok := <-consumer.C:
		timer.Stop()
		if !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("queue is closed."))
			return
		}

		if query_params.Get("batch") != "true" {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			bs := msg.Bytes()
			if len(bs) > 0 {
				w.Write(msg.Bytes())
			}
		} else {
			msgList := readMore(consumer.C, msg)
			w.Header().Add("X-HW-Batch", strconv.FormatInt(int64(len(msgList)), 10))
			w.WriteHeader(http.StatusOK)

			w.Write([]byte("["))
			is_frist := true
			for _, m := range msgList {
				bs := m.Bytes()
				if len(bs) > 0 {
					if is_frist {
						is_frist = false
					} else {
						w.Write([]byte(","))
					}

					w.Write(bs)
				}
			}
			w.Write([]byte("]"))
		}

	case <-timer.C:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (se *StandardEngine) doPost(w http.ResponseWriter, r *http.Request,
	urlPath string, cb func(name string) Producer) {
	query_params := r.URL.Query()

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	r.Body.Close()

	timeout := GetTimeout(query_params, 0)
	msg := CreateDataMessage(bs)
	send := cb(urlPath)
	if timeout == 0 {
		err = send.Send(msg)
	} else {
		err = send.SendTimeout(msg, timeout)
	}

	w.Header().Add("Content-Type", "text/plain")
	if err != nil {
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func GetTimeout(query_params url.Values, value time.Duration) time.Duration {
	s := query_params.Get("timeout")
	if "" == s {
		return value
	}
	t, e := time.ParseDuration(s)
	if nil != e {
		return value
	}
	return t
}

func (se *StandardEngine) queuesIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(se.Server.GetQueues())
}

func (se *StandardEngine) topicsIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(se.Server.GetTopics())
}

func (se *StandardEngine) clientsIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(se.Server.GetClients())
}

func (se *StandardEngine) subscribe(ws *websocket.Conn, consumer *Consumer, c chan struct{}) {
	var remoteAddr = getRealIP(ws)
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString("[panic] [broker] connection(write: ")
			buffer.WriteString(remoteAddr)
			buffer.WriteString(") \r\n")
			buffer.Write(debug.Stack())
			log.Println(buffer.String())
		}
		ws.Close()
	}()

	is_running := true
	for is_running {
		select {
		case msg, ok := <-consumer.C:
			if !ok {
				is_running = false
				log.Println("[broker] connection(write:", remoteAddr, ") is closed - queue is shutdown.")
				break
			}

			if e := websocket.Message.Send(ws, msg.Bytes()); nil != e {
				is_running = false

				if strings.Contains(e.Error(), "use of closed network connection") {
					log.Println("[broker] connection(write:", remoteAddr, ") is closed.")
				} else {
					log.Println("[broker] connection(write:", remoteAddr, ") is closed -", e)
				}
				consumer.Unread(msg)
				break
			}
		case <-c:
			is_running = false
			log.Println("[broker] connection(write:", remoteAddr, ") is closed - queue is shutdown.")
		}
	}
}

func (se *StandardEngine) publish(ws *websocket.Conn, producer Producer, c chan struct{}) {
	var remoteAddr = getRealIP(ws)
	defer func() {
		if o := recover(); nil != o {
			var buffer bytes.Buffer
			buffer.WriteString("[panic] [broker] connection(write: ")
			buffer.WriteString(remoteAddr)
			buffer.WriteString(") \r\n")
			buffer.Write(debug.Stack())
			log.Println(buffer.String())
		}
		ws.Close()
	}()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	const trySendCount = 2
	is_running := true
	for is_running {
		var data []byte
		if e := websocket.Message.Receive(ws, &data); nil != e {
			if e == io.EOF {
				log.Println("[broker] connection(read:", remoteAddr, ") is closed - peer is shutdown.")
			} else if strings.Contains(e.Error(), "use of closed network connection") {
				log.Println("[broker] connection(read:", remoteAddr, ") is closed.")
			} else {
				log.Println("[broker] connection(read:", remoteAddr, ") is closed -", e)
			}
			is_running = false
			break
		}

		continueTick := 0
		for continueTick < trySendCount {
			select {
			case producer.Chan() <- CreateDataMessage(data):
				continueTick = math.MaxInt32
			case <-ticker.C:
				continueTick++
				if continueTick >= trySendCount {
					log.Println("[broker] connection(write:", remoteAddr, ") is closed - queue is overflow.")
				}
			case <-c:
				log.Println("[broker] connection(write:", remoteAddr, ") is closed - queue is shutdown.")

				continueTick = math.MaxInt32
				is_running = false
			}
		}
	}
}
