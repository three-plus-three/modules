package environment

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

type Base struct {
	Closed  int32
	S       chan struct{}
	Wait    sync.WaitGroup
	closers []io.Closer
}

func (self *Base) CloseWith(closeHandle func() error) error {
	if !atomic.CompareAndSwapInt32(&self.Closed, 0, 1) {
		return nil
	}
	if nil != self.S {
		close(self.S)
	}
	var err error
	if nil != closeHandle {
		err = closeHandle()
	}

	for _, closer := range self.closers {
		if e := closer.Close(); e != nil {
			if err == nil {
				err = e
			}
		}
	}
	self.Wait.Wait()
	return err
}

func (self *Base) IsClosed() bool {
	return 0 != atomic.LoadInt32(&self.Closed)
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
