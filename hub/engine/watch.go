package engine

import (
	"sync/atomic"

	"sync"
)

type Watcher interface {
	OnNewQueue(name string)
	OnRemoveQueue(name string)
	OnNewTopic(name string)
	OnRemoveTopic(name string)
}

type WatcherImpl struct {
	mu           sync.Mutex
	newQueues    atomic.Value
	deleteQueues atomic.Value
	newTopics    atomic.Value
	deleteTopics atomic.Value
}

func (impl *WatcherImpl) add(value *atomic.Value, cb func(name string)) {
	var funcs []func(name string)

	if o := value.Load(); o != nil {
		if fList, ok := o.([]func(name string)); ok {
			funcs = make([]func(name string), len(fList), len(fList)+1)
			copy(funcs, fList)
		}
	}

	funcs = append(funcs, cb)
	value.Store(funcs)
}

func (impl *WatcherImpl) call(value *atomic.Value, name string) {
	if o := value.Load(); o != nil {
		if fList, ok := o.([]func(name string)); ok {
			for _, cb := range fList {
				cb(name)
			}
		}
	}
}

func (impl *WatcherImpl) ListenNewQueue(cb func(name string)) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	impl.add(&impl.newQueues, cb)
}

func (impl *WatcherImpl) ListenRemoveQueue(cb func(name string)) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	impl.add(&impl.deleteQueues, cb)
}

func (impl *WatcherImpl) ListenNewTopic(cb func(name string)) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	impl.add(&impl.newTopics, cb)
}

func (impl *WatcherImpl) ListenRemoveTopic(cb func(name string)) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	impl.add(&impl.deleteTopics, cb)
}

func (impl *WatcherImpl) OnNewQueue(name string) {
	impl.call(&impl.newQueues, name)
}

func (impl *WatcherImpl) OnRemoveQueue(name string) {
	impl.call(&impl.deleteQueues, name)
}

func (impl *WatcherImpl) OnNewTopic(name string) {
	impl.call(&impl.newTopics, name)
}

func (impl *WatcherImpl) OnRemoveTopic(name string) {
	impl.call(&impl.deleteTopics, name)
}

/*
type watcher struct {
	topic Producer
	watch Watcher
}

func (w *watcher) onNewQueue(name string) {
	msg := mq_client.NewMessageWriter(mq_client.MSG_DATA, len(name)+1).
		Append([]byte("new queue ")).
		Append([]byte(name)).
		Append([]byte("\n")).
		Build()
	w.topic.Send(msg)
	w.watch.OnNewQueue(name)
}

func (w *watcher) onRemoveQueue(name string) {
	msg := mq_client.NewMessageWriter(mq_client.MSG_DATA, len(name)+1).
		Append([]byte("del queue ")).
		Append([]byte(name)).
		Append([]byte("\n")).
		Build()
	w.topic.Send(msg)
	w.watch.OnRemoveQueue(name)
}

func (w *watcher) onNewTopic(name string) {
	msg := mq_client.NewMessageWriter(mq_client.MSG_DATA, len(name)+1).
		Append([]byte("new topic ")).
		Append([]byte(name)).
		Append([]byte("\n")).
		Build()
	w.topic.Send(msg)
	w.watch.OnNewTopic(name)
}

func (w *watcher) onRemoveTopic(name string) {
	msg := mq_client.NewMessageWriter(mq_client.MSG_DATA, len(name)+1).
		Append([]byte("del topic ")).
		Append([]byte(name)).
		Append([]byte("\n")).
		Build()
	w.topic.Send(msg)
	w.watch.OnRemoveTopic(name)
}
*/

type dummyWatcher struct{}

func (self *dummyWatcher) OnNewQueue(name string) {
}

func (self *dummyWatcher) OnRemoveQueue(name string) {
}

func (self *dummyWatcher) OnNewTopic(name string) {
}

func (self *dummyWatcher) OnRemoveTopic(name string) {
}

var DummyWatcher Watcher = &dummyWatcher{}
