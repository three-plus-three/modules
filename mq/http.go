package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	Core    *Core
	NoRoute http.Handler
}

func (se *StandardEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/sendQueue", "/sendQueue/":
		se.send(w, r,
			func(name string) (interface{}, error) {
				queue := se.Core.CreateQueueIfNotExists(name)
				if queue == nil {
					return nil, errors.New("create queue fail")
				}
				return queue, nil
			},
			func(stub *engineStub, o interface{}) {
				queue := o.(*Queue)
				stub.publish(queue.C)
			})
	case "/sendTopic", "/sendTopic/":
		se.send(w, r,
			func(name string) (interface{}, error) {
				topic := se.Core.CreateTopicIfNotExists(name)
				if topic == nil {
					return nil, errors.New("create topic fail")
				}
				return topic, nil
			},
			func(stub *engineStub, o interface{}) {
				topic := o.(*Topic)
				stub.sendToTopic(topic)
			})
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
						return se.Core.CreateQueueIfNotExists(name).ListenOn()
					})
			case "POST", "PUT":
				se.doPost(w, r, urlPath,
					func(name string) Producer {
						return se.Core.CreateQueueIfNotExists(name)
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
						return se.Core.CreateTopicIfNotExists(name).ListenOn()
					})
			case "POST", "PUT":
				se.doPost(w, r, urlPath,
					func(name string) Producer {
						return se.Core.CreateTopicIfNotExists(name)
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
		timer := time.NewTimer(timeout)
		_, err = send.SendWithContext(msg, timer.C)
		if err == nil {
			timer.Stop()
		}
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
	json.NewEncoder(w).Encode(se.Core.GetQueues())
}

func (se *StandardEngine) topicsIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(se.Core.GetTopics())
}

func (se *StandardEngine) clientsIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(se.Core.GetClients())
}

func (se *StandardEngine) subscribeQueue(w http.ResponseWriter, r *http.Request) {
	se.subscribe(w, r, "queue", func(name string) *Consumer {
		queue := se.Core.CreateQueueIfNotExists(name)
		if queue == nil {
			return nil
		}
		return queue.ListenOn()
	})
}

func (se *StandardEngine) subscribeTopic(w http.ResponseWriter, r *http.Request) {
	se.subscribe(w, r, "topic", func(name string) *Consumer {
		queue := se.Core.CreateTopicIfNotExists(name)
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

		stub.disconnect = se.Core.Connect(stub)

		go func() {
			defer stub.Close()
			for {
				var data []byte
				if e := websocket.Message.Receive(conn, &data); nil != e {
					if e == io.EOF {
						log.Println("[broker] connection(read:", stub.remoteAddr, ") is closed - peer is shutdown.")
					} else if strings.Contains(e.Error(), "use of closed network connection") {
						log.Println("[broker] connection(read:", stub.remoteAddr, ") is closed.")
					} else {
						log.Println("[broker] connection(read:", stub.remoteAddr, ") is closed -", e)
					}
					break
				}
			}
		}()

		stub.subscribe(consumer)
	})

	stub.srv.ServeHTTP(w, r)
}

func (se *StandardEngine) send(w http.ResponseWriter, r *http.Request,
	create func(name string) (interface{}, error),
	run func(stub *engineStub, o interface{})) {

	params := r.URL.Query()

	stub := &engineStub{
		createdAt:  time.Now(),
		remoteAddr: getRealIP(r),
		mode:       "queue",
		role:       "pushlisher",
		name:       params.Get("name"),
		c:          make(chan struct{})}
	var o interface{}

	stub.srv.Handshake = func(config *websocket.Config, req *http.Request) (err error) {
		config.Origin, err = websocket.Origin(config, req)
		if err == nil && config.Origin == nil {
			return fmt.Errorf("null origin")
		}
		if stub.name == "" {
			return errors.New("name is missing")
		}
		o, err = create(stub.name)
		if err != nil {
			return err
		}
		return nil
	}

	stub.srv.Handler = websocket.Handler(func(conn *websocket.Conn) {
		stub.conn = conn
		defer stub.Close()

		stub.disconnect = se.Core.Connect(stub)

		run(stub, o)
	})

	stub.srv.ServeHTTP(w, r)
}

func NewEngine(opts *Options, noRoute http.Handler) (*StandardEngine, error) {
	core, err := NewCore(opts)
	if err != nil {
		return nil, err
	}
	return &StandardEngine{
		Core:    core,
		NoRoute: noRoute,
	}, nil
}
