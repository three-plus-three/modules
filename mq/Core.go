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

type Core struct {
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

func (core *Core) IsClosed() bool {
	return atomic.LoadInt32(&core.closed) != 0
}

func (core *Core) Close() error {
	if !atomic.CompareAndSwapInt32(&core.closed, 0, 1) {
		return ErrAlreadyClosed
	}

	func() {
		core.clients_lock.Lock()
		defer core.clients_lock.Unlock()
		for el := core.clients.Front(); el != nil; el = el.Next() {
			if conn, ok := el.Value.(io.Closer); ok {
				conn.Close()
			}
		}
	}()

	func() {
		core.queues_lock.Lock()
		defer core.queues_lock.Unlock()
		for _, v := range core.queues {
			v.Close()
		}
	}()

	func() {
		core.topics_lock.Lock()
		defer core.topics_lock.Unlock()
		for _, v := range core.topics {
			v.Close()
		}
	}()

	core.waitGroup.Wait()
	return nil
}

func (core *Core) Wait() {
	core.waitGroup.Wait()
}

func (core *Core) GetOptions() *Options {
	return &core.options
}

func (core *Core) GetQueues() []string {
	core.queues_lock.RLock()
	defer core.queues_lock.RUnlock()
	var results []string
	for k, _ := range core.queues {
		results = append(results, k)
	}
	return results
}

func (core *Core) GetTopics() []string {
	core.topics_lock.RLock()
	defer core.topics_lock.RUnlock()
	var results []string
	for k, _ := range core.topics {
		results = append(results, k)
	}
	return results
}

type DisconnectFunc func()

func (core *Core) Connect(client interface{}) DisconnectFunc {
	var el *list.Element

	core.clients_lock.Lock()
	if core.clients == nil {
		core.clients = list.New()
	}
	el = core.clients.PushBack(client)
	core.clients_lock.Unlock()

	return DisconnectFunc(func() {
		core.clients_lock.Lock()
		if core.clients != nil {
			core.clients.Remove(el)
		}
		core.clients_lock.Unlock()
	})
}

func (core *Core) GetClients() []map[string]interface{} {
	core.clients_lock.Lock()
	defer core.clients_lock.Unlock()
	var results []map[string]interface{}

	for el := core.clients.Front(); el != nil; el = el.Next() {
		if cli, ok := el.Value.(Client); ok {
			results = append(results, cli.Info())
		}
	}

	return results
}

func (core *Core) log(args ...interface{}) {
	core.options.Logger.Println(args...)
}

func (core *Core) logf(format string, args ...interface{}) {
	core.options.Logger.Printf(format, args...)
}

func (core *Core) catchThrow(ctx string, cb func()) {
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
		core.logf(buffer.String())
	}
}

func (core *Core) RunItInGoroutine(cb func()) {
	core.waitGroup.Add(1)
	go func() {
		cb()
		core.waitGroup.Done()
	}()
}

func (core *Core) KillQueueIfExists(name string) {
	core.queues_lock.Lock()
	queue, ok := core.queues[name]
	if ok {
		delete(core.queues, name)
	}
	core.queues_lock.Unlock()

	if ok {
		queue.Close()
		core.watcher.OnRemoveQueue(name)
	}
}

func (core *Core) KillTopicIfExists(name string) {
	core.topics_lock.RLock()
	topic, ok := core.topics[name]
	if ok {
		delete(core.topics, name)
	}
	core.topics_lock.RUnlock()

	if ok {
		topic.Close()
		core.watcher.OnRemoveTopic(name)
	}
}

func (core *Core) GetQueueIfExists(name string) *Queue {
	core.queues_lock.RLock()
	queue, _ := core.queues[name]
	core.queues_lock.RUnlock()
	return queue
}

func (core *Core) GetTopicIfExists(name string) *Topic {
	core.topics_lock.RLock()
	topic, _ := core.topics[name]
	core.topics_lock.RUnlock()
	return topic
}

func (core *Core) CreateQueueIfNotExists(name string) *Queue {
	core.queues_lock.RLock()
	queue, ok := core.queues[name]
	core.queues_lock.RUnlock()

	if ok {
		return queue
	}

	core.queues_lock.Lock()
	queue, ok = core.queues[name]
	if ok {
		core.queues_lock.Unlock()
		return queue
	}

	queue = creatQueue(core, name, core.options.MsgQueueCapacity)
	core.queues[name] = queue
	core.queues_lock.Unlock()

	core.watcher.OnNewQueue(name)
	return queue
}

func (core *Core) CreateTopicIfNotExists(name string) *Topic {
	core.topics_lock.RLock()
	topic, ok := core.topics[name]
	core.topics_lock.RUnlock()

	if ok {
		return topic
	}

	core.topics_lock.Lock()
	topic, ok = core.topics[name]
	if ok {
		core.topics_lock.Unlock()
		return topic
	}
	topic = creatTopic(core, name, core.options.MsgQueueCapacity)
	core.topics[name] = topic
	core.topics_lock.Unlock()

	core.watcher.OnNewTopic(name)
	return topic
}

func NewCore(opts *Options) (*Core, error) {
	opts.ensureDefault()

	core := &Core{
		options: *opts,
		clients: list.New(),
		queues:  map[string]*Queue{},
		topics:  map[string]*Topic{},
	}

	core.watcher = DummyWatcher
	if opts.Watch != nil {
		core.watcher = opts.Watch
	}
	return core, nil
}
