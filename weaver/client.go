package weaver

import (
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cheekybits/genny/generic"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/hub"
)

// ValueType 用于泛型替换的类型
type ValueType generic.Type

// Client 菜单服务
type Client interface {
	io.Closer

	WhenChanged(cb func())

	Read() (ValueType, error)

	Flush() error
}

// Callback 菜单的读取函数
type Callback func() (ValueType, error)

// Connect 连接到 weaver 服务
func Connect(env *environment.Environment, appID environment.ENV_PROXY_TYPE,
	cb Callback, mode, queueName, urlPath string, logger *log.Logger) Client {
	// wsrv := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)
	// hubURL := so.UrlFor(env.DaemonUrlPath, "/mq/")
	// builder := hub.Connect(hubURL)

	switch mode {
	case "apart":
		wsrv := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)
		apart := &apartClient{
			logger:    logger, // log.New(os.Stderr, "[menus]", log.LstdFlags),
			env:       env,
			wsrv:      wsrv,
			appSrv:    env.GetServiceConfig(appID),
			client:    wsrv.Client(urlPath),
			queueName: queueName,
			cb:        cb,
			c:         make(chan struct{}),
		}
		go apart.run()
		go apart.runSub()
		return apart
	default:
		return &standaloneClient{env: env, cb: cb}
	}
}

type standaloneClient struct {
	env *environment.Environment
	cb  Callback
}

func (srv *standaloneClient) Close() error {
	return nil
}

func (srv *standaloneClient) Flush() error {
	return nil
}

func (srv *standaloneClient) WhenChanged(cb func()) {
}

func (srv *standaloneClient) Read() (ValueType, error) {
	return srv.cb()
}

type apartResult struct {
	value ValueType
	err   error
}

type apartClient struct {
	logger    *log.Logger
	env       *environment.Environment
	wsrv      *environment.ServiceConfig
	appSrv    *environment.ServiceConfig
	client    environment.HttpClient
	queueName string
	cb        Callback

	isClosed int32
	cw       concurrency.CloseWrapper
	pad      int32
	c        chan struct{}
	cached   atomic.Value
	mu       sync.Mutex
	cbList   []func()
}

func (srv *apartClient) Close() error {
	if atomic.CompareAndSwapInt32(&srv.isClosed, 0, 1) {
		close(srv.c)
		return srv.cw.Close()
	}
	return nil
}

func (srv *apartClient) save(value ValueType, err error) {
	srv.cached.Store(&apartResult{
		value: value,
		err:   err,
	})
	srv.mu.Lock()
	defer srv.mu.Unlock()

	for _, cb := range srv.cbList {
		go cb()
	}
}

func (srv *apartClient) Read() (ValueType, error) {
	o := srv.cached.Load()
	if o != nil {
		if result, ok := o.(*apartResult); ok {
			return result.value, result.err
		}
	}

	value, err := srv.read()
	srv.save(value, err)
	return value, err
}

func (srv *apartClient) read() (ValueType, error) {
	var value ValueType
	err := srv.client.
		SetParam("app", srv.appSrv.Name).
		GET(&value)
	return value, err
}

func (srv *apartClient) write() error {
	value, err := srv.cb()
	if err != nil {
		return err
	}
	o := srv.cached.Load()
	if o != nil {
		if result, ok := o.(*apartResult); ok {
			if isSubset(result.value, value) {
				return nil
			}
		}
	}

	return srv.client.
		SetParam("app", srv.appSrv.Name).
		SetBody(value).
		POST(nil)
}

func (srv *apartClient) WhenChanged(cb func()) {
	if atomic.LoadInt32(&srv.isClosed) != 0 {
		panic(ErrAlreadyClosed)
	}

	srv.mu.Lock()
	srv.cbList = append(srv.cbList, cb)
	srv.mu.Unlock()
}

func (srv *apartClient) Flush() error {
	if atomic.LoadInt32(&srv.isClosed) != 0 {
		return ErrAlreadyClosed
	}
	select {
	case srv.c <- struct{}{}:
	default:
	}
	return nil
}

func (srv *apartClient) runSub() {
	errCount := 0
	hubURL := srv.wsrv.UrlFor(srv.env.DaemonUrlPath, "/mq/")
	builder := hub.Connect(hubURL)

	for atomic.LoadInt32(&srv.isClosed) == 0 {

		topic, err := builder.SubscribeTopic(srv.queueName)
		if err != nil {
			errCount++
			if errCount%50 < 3 {
				srv.logger.Println("subscribe", srv.queueName, "fail,", err)
			}

			select {
			case v, ok := <-srv.c:
				if ok {
					srv.c <- v
				}
			case <-time.After(1 * time.Second):
			}
			continue
		}
		srv.cw.Set(topic)

		errCount = 0
		err = topic.Run(func(sub *hub.Subscription, msg hub.Message) {
			value, err := srv.read()
			srv.save(value, err)
		})
		if err != nil {
			srv.logger.Println("subscribe", srv.queueName, "fail,", err)
		}
		srv.cw.Set(nil)
	}
}

func (srv *apartClient) run() {
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	writed := false

	flush := func() {
		if err := srv.write(); err != nil {
			srv.logger.Println("write value fail", err)
		} else {
			writed = true
		}

		if value, err := srv.read(); err != nil {
			srv.logger.Println("read value fail", err)
		} else {
			srv.save(value, err)
		}

		if writed {
			timer.Reset(5 * time.Minute)
		} else {
			timer.Reset(10 * time.Second)
		}
	}

	for atomic.LoadInt32(&srv.isClosed) == 0 {
		select {
		case _, ok := <-srv.c:
			if !ok {
				return
			}
			flush()
		case <-timer.C:
			flush()
		}
	}
}
