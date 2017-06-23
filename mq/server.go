package server

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
)

type Client interface {
	Info() map[string]interface{}
}

type Server struct {
	options Options

	closer    []io.Closer
	waitGroup sync.WaitGroup

	watcher      Watcher
	clients_lock sync.Mutex
	clients      *list.List

	queues_lock sync.RWMutex
	queues      map[string]*Queue

	topics_lock sync.RWMutex
	topics      map[string]*Topic

	closed int32
}

func (self *Server) IsClosed() bool {
	return atomic.LoadInt32(&self.closed) != 0
}

func (self *Server) Close() error {
	if !atomic.CompareAndSwapInt32(&self.closed, 0, 1) {
		return ErrAlreadyClosed
	}

	func() {
		self.clients_lock.Lock()
		defer self.clients_lock.Unlock()
		for el := self.clients.Front(); el != nil; el = el.Next() {
			if conn, ok := el.Value.(io.Closer); ok {
				conn.Close()
			}
		}
	}()

	func() {
		self.queues_lock.Lock()
		defer self.queues_lock.Unlock()
		for _, v := range self.queues {
			v.Close()
		}
	}()

	func() {
		self.topics_lock.Lock()
		defer self.topics_lock.Unlock()
		for _, v := range self.topics {
			v.Close()
		}
	}()

	self.waitGroup.Wait()
	return nil
}

func (self *Server) Wait() {
	self.waitGroup.Wait()
}

func (self *Server) GetOptions() *Options {
	return &self.options
}

func (self *Server) GetQueues() []string {
	self.queues_lock.RLock()
	defer self.queues_lock.RUnlock()
	var results []string
	for k, _ := range self.queues {
		results = append(results, k)
	}
	return results
}

func (self *Server) GetTopics() []string {
	self.topics_lock.RLock()
	defer self.topics_lock.RUnlock()
	var results []string
	for k, _ := range self.topics {
		results = append(results, k)
	}
	return results
}

type DisconnectFunc func()

func (self *Server) Connect(client interface{}) DisconnectFunc {
	var el *list.Element

	self.clients_lock.Lock()
	if self.clients == nil {
		self.clients = list.New()
	}
	el = self.clients.PushBack(client)
	self.clients_lock.Unlock()

	return DisconnectFunc(func() {
		self.clients_lock.Lock()
		if self.clients != nil {
			self.clients.Remove(el)
		}
		self.clients_lock.Unlock()
	})
}

func (self *Server) GetClients() []map[string]interface{} {
	self.clients_lock.Lock()
	defer self.clients_lock.Unlock()
	var results []map[string]interface{}

	for el := self.clients.Front(); el != nil; el = el.Next() {
		if cli, ok := el.Value.(Client); ok {
			results = append(results, cli.Info())
		}
	}

	return results
}

func (self *Server) log(args ...interface{}) {
	self.options.Logger.Println(args...)
}

func (self *Server) logf(format string, args ...interface{}) {
	self.options.Logger.Printf(format, args...)
}

func (self *Server) catchThrow(ctx string, cb func()) {
	if e := recover(); nil != e {
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("[panic] %s %v", ctx, e))
		for i := 1; ; i += 1 {
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			funcinfo := runtime.FuncForPC(pc)
			if nil != funcinfo {
				buffer.WriteString(fmt.Sprintf("    %s:%d %s\r\n", file, line, funcinfo.Name()))
			} else {
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
		}

		if cb != nil {
			cb()
		}
		self.logf(buffer.String())
	}
}

func (self *Server) RunItInGoroutine(cb func()) {
	self.waitGroup.Add(1)
	go func() {
		cb()
		self.waitGroup.Done()
	}()
}

func (self *Server) KillQueueIfExists(name string) {
	self.queues_lock.Lock()
	queue, ok := self.queues[name]
	if ok {
		delete(self.queues, name)
	}
	self.queues_lock.Unlock()

	if ok {
		queue.Close()
		self.watcher.OnRemoveQueue(name)
	}
}

func (self *Server) KillTopicIfExists(name string) {
	self.topics_lock.RLock()
	topic, ok := self.topics[name]
	if ok {
		delete(self.topics, name)
	}
	self.topics_lock.RUnlock()

	if ok {
		topic.Close()
		self.watcher.OnRemoveTopic(name)
	}
}

func (self *Server) GetQueueIfExists(name string) *Queue {
	self.queues_lock.RLock()
	queue, _ := self.queues[name]
	self.queues_lock.RUnlock()
	return queue
}

func (self *Server) GetTopicIfExists(name string) *Topic {
	self.topics_lock.RLock()
	topic, _ := self.topics[name]
	self.topics_lock.RUnlock()
	return topic
}

func (self *Server) CreateQueueIfNotExists(name string) *Queue {
	self.queues_lock.RLock()
	queue, ok := self.queues[name]
	self.queues_lock.RUnlock()

	if ok {
		return queue
	}

	self.queues_lock.Lock()
	queue, ok = self.queues[name]
	if ok {
		self.queues_lock.Unlock()
		return queue
	}

	queue = creatQueue(self, name, self.options.MsgQueueCapacity)
	self.queues[name] = queue
	self.queues_lock.Unlock()

	self.watcher.OnNewQueue(name)
	return queue
}

func (self *Server) CreateTopicIfNotExists(name string) *Topic {
	self.topics_lock.RLock()
	topic, ok := self.topics[name]
	self.topics_lock.RUnlock()

	if ok {
		return topic
	}

	self.topics_lock.Lock()
	topic, ok = self.topics[name]
	if ok {
		self.topics_lock.Unlock()
		return topic
	}
	topic = creatTopic(self, name, self.options.MsgQueueCapacity)
	self.topics[name] = topic
	self.topics_lock.Unlock()

	self.watcher.OnNewTopic(name)
	return topic
}

func NewServer(opts *Options) (*Server, error) {
	opts.ensureDefault()

	srv := &Server{
		options: *opts,
		clients: list.New(),
		queues:  map[string]*Queue{},
		topics:  map[string]*Topic{},
	}

	srv.watcher = DummyWatcher
	if opts.Watch != nil {
		srv.watcher = opts.Watch
	}
	return srv, nil
}
