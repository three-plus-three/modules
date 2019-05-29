package concurrency

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/three-plus-three/modules/util"
)

type Closes struct {
	mu      sync.Mutex
	closers []io.Closer
}

func (self *Closes) OnClosing(closers ...io.Closer) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.closers = append(self.closers, closers...)
}

func (self *Closes) CloseWith(closeHandle func() error) error {
	var err error
	if nil != closeHandle {
		err = closeHandle()
	}

	func() {
		self.mu.Lock()
		defer self.mu.Unlock()
		for _, closer := range self.closers {
			if e := closer.Close(); e != nil {
				if err == nil {
					err = e
				}
			}
		}
	}()
	return err
}

type Base struct {
	closed int32
	S      chan struct{}
	Wait   sync.WaitGroup

	Closes
}

func (self *Base) CloseWith(closeHandle func() error) error {
	if !atomic.CompareAndSwapInt32(&self.closed, 0, 1) {
		return nil
	}
	err := self.Closes.CloseWith(func() error {
		if nil != self.S {
			close(self.S)
		}

		if nil != closeHandle {
			return closeHandle()
		}
		return nil
	})
	self.Wait.Wait()
	return err
}

func (self *Base) IsClosed() bool {
	return 0 != atomic.LoadInt32(&self.closed)
}

func (self *Base) CatchThrow(message string, err *error) {
	if o := recover(); nil != o {
		var buffer bytes.Buffer
		if "" == message {
			buffer.WriteString(fmt.Sprintf("[panic] %v", o))
		} else {
			buffer.WriteString(fmt.Sprintf("[panic] %v - %v", message, o))
		}

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

		errMsg := buffer.String()
		log.Println(errMsg)
		if nil != err {
			*err = errors.New(errMsg)
		}
	}
}

func (self *Base) RunItInGoroutine(cb func()) {
	self.Wait.Add(1)
	go func() {
		cb()
		self.Wait.Done()
	}()
}

type CloseWrapper struct {
	v atomic.Value
}

func (cw *CloseWrapper) Set(closer io.Closer) {
	cw.v.Store(&closeWrapper{v: closer})
}

func (cw *CloseWrapper) Close() error {
	o := cw.v.Load()
	if o == nil {
		return nil
	}
	if closer, ok := o.(io.Closer); ok && closer != nil {
		err := closer.Close()
		cw.Set(nil)
		return err
	}
	return nil
}

type closeWrapper struct {
	v io.Closer
}

func (cw *closeWrapper) Close() error {
	if cw.v == nil {
		return nil
	}
	return cw.v.Close()
}

type CloseW struct {
	C interface {
		Close()
	}
}

func (cw *CloseW) Close() error {
	if cw.C != nil {
		cw.C.Close()
	}
	return nil
}

func ToCloser(c interface{}) io.Closer {
	if cw, ok := c.(interface {
		Close()
	}); ok {
		return CloseFunc(func() error {
			if cw != nil {
				cw.Close()
			}
			return nil
		})
	}

	if cf, ok := c.(func() error); ok {
		return CloseFunc(cf)
	}

	if closer, ok := c.(io.Closer); ok {
		return closer
	}
	panic(errors.New("it isn't a closer"))
}

type CloseFunc = util.CloseFunc
