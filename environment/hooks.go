package environment

import (
	"bytes"
	"errors"
	"sync"
)

var (
	hookLock sync.Mutex
	hooks    []func(env *Environment) error
)

func On(cb func(env *Environment) error) {
	hookLock.Lock()
	defer hookLock.Unlock()
	hooks = append(hooks, cb)
}

func callHooks(env *Environment) error {
	var errList []error
	hookLock.Lock()
	defer hookLock.Unlock()
	for _, cb := range hooks {
		if e := cb(env); e != nil {
			errList = append(errList, e)
		}
	}
	if len(errList) == 0 {
		return nil
	}

	var buf bytes.Buffer
	buf.WriteString("call hooks:")
	for _, e := range errList {
		buf.WriteString("\r\n\t\t")
		buf.WriteString(e.Error())
	}
	return errors.New(buf.String())
}
