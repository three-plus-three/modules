package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

func getRealIP(req *http.Request) string {
	if nil != req {
		address := req.Header.Get("X-Real-IP")
		if "" == address {
			address = req.Header.Get("X-Forwarded-For")
			if "" == address {
				address = req.RemoteAddr
			}
		}
		port := req.Header.Get("X-Real-Port")
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
	case "/sendQueue", "/sendQueue/":
		se.sendQueue(w, r)
	case "/sendTopic", "/sendTopic/":
		se.sendTopic(w, r)
	case "/subscribeQueue", "/subscribeQueue/":
		se.subscribeQueue(w, r)
	case "/subscribeTopic", "/subscribeTopic/":
		se.subscribeTopic(w, r)
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

func (se *StandardEngine) subscribeQueue(w http.ResponseWriter, r *http.Request) {
	se.subscribe(w, r, "queue", func(name string) *Consumer {
		queue := se.Server.CreateQueueIfNotExists(name)
		if queue == nil {
			return nil
		}
		return queue.ListenOn()
	})
}

func (se *StandardEngine) subscribeTopic(w http.ResponseWriter, r *http.Request) {
	se.subscribe(w, r, "topic", func(name string) *Consumer {
		queue := se.Server.CreateTopicIfNotExists(name)
		if queue == nil {
			return nil
		}
		return queue.ListenOn()
	})
}

func (se *StandardEngine) subscribe(w http.ResponseWriter, r *http.Request, mode string, cb func(name string) *Consumer) {
	params := r.URL.Query()

	stub := &engineStub{
		createdAt:  time.Now(),
		remoteAddr: getRealIP(r),
		mode:       mode,
		role:       "subscriber",
		name:       params.Get("name"),
		c:          make(chan struct{})}
	var consumer *Consumer

	stub.srv.Handshake = func(config *websocket.Config, req *http.Request) (err error) {
		config.Origin, err = websocket.Origin(config, req)
		if err == nil && config.Origin == nil {
			return fmt.Errorf("null origin")
		}
		if stub.name == "" {
			return errors.New("queue name is missing")
		}
		consumer = cb(stub.name)
		if consumer == nil {
			return errors.New("create queue fail")
		}
		return nil
	}

	stub.srv.Handler = websocket.Handler(func(conn *websocket.Conn) {
		stub.conn = conn
		defer stub.Close()

		stub.disconnect = se.Server.Connect(stub)

		go func() {
			defer stub.Close()
			for {
				var data []byte
				if e := websocket.Message.Receive(conn, &data); nil != e {
					break
				}
			}
		}()

		stub.subscribe(consumer)
	})

	stub.srv.ServeHTTP(w, r)
}

func (se *StandardEngine) sendQueue(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	stub := &engineStub{
		createdAt:  time.Now(),
		remoteAddr: getRealIP(r),
		mode:       "queue",
		role:       "pushlisher",
		name:       params.Get("name"),
		c:          make(chan struct{})}
	var queue *Queue

	stub.srv.Handshake = func(config *websocket.Config, req *http.Request) (err error) {
		config.Origin, err = websocket.Origin(config, req)
		if err == nil && config.Origin == nil {
			return fmt.Errorf("null origin")
		}
		if stub.name == "" {
			return errors.New("queue name is missing")
		}
		queue = se.Server.CreateQueueIfNotExists(stub.name)
		if queue == nil {
			return errors.New("create queue fail")
		}
		return nil
	}

	stub.srv.Handler = websocket.Handler(func(conn *websocket.Conn) {
		stub.conn = conn
		defer stub.Close()

		stub.disconnect = se.Server.Connect(stub)

		go func() {
			defer stub.Close()
			for {
				var data []byte
				if e := websocket.Message.Receive(conn, &data); nil != e {
					break
				}
			}
		}()

		stub.publish(queue.C)
	})

	stub.srv.ServeHTTP(w, r)
}

func (se *StandardEngine) sendTopic(w http.ResponseWriter, r *http.Request) {
}
