package weaver

import (
	"io"
	"log"
	"os"
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

	Read() (ValueType, error)
}

// Callback 菜单的读取函数
type Callback func() (ValueType, error)

// Connect 连接到 weaver 服务
func Connect(env *environment.Environment, appID environment.ENV_PROXY_TYPE, cb Callback, mode string) Client {
	//wsrv := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)

	// hubURL := so.UrlFor(env.DaemonUrlPath, "/mq/")
	// builder := hub.Connect(hubURL)

	switch mode {
	case "apart":
		wsrv := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)
		apart := &apartClient{
			logger:    log.New(os.Stderr, "[menus]", log.LstdFlags),
			env:       env,
			wsrv:      wsrv,
			appSrv:    env.GetServiceConfig(appID),
			client:    wsrv.Client(env.DaemonUrlPath, "/menu/"),
			queueName: "menus.changed",
			cb:        cb,
			shutdown:  make(chan struct{}),
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
	shutdown chan struct{}
	cached   atomic.Value
}

func (srv *apartClient) Close() error {
	if atomic.CompareAndSwapInt32(&srv.isClosed, 0, 1) {
		close(srv.shutdown)
		return srv.cw.Close()
	}
	return nil
}

func (srv *apartClient) Read() (ValueType, error) {
	o := srv.cached.Load()
	if o != nil {
		if result, ok := o.(*apartResult); ok {
			return result.value, result.err
		}
	}

	value, err := srv.read()
	srv.cached.Store(&apartResult{
		value: value,
		err:   err,
	})
	return value, err
}

func (srv *apartClient) read() (ValueType, error) {
	var value ValueType
	err := srv.client.GET(&value)
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
			if !isSame(value, result.value) {
				return nil
			}
		}
	}

	return srv.client.
		SetParam("app", srv.appSrv.Name).
		POST(value)
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
			continue
		}
		srv.cw.Set(topic)

		errCount = 0
		err = topic.Run(func(sub *hub.Subscription, msg hub.Message) {
			value, err := srv.read()
			srv.cached.Store(&apartResult{
				value: value,
				err:   err,
			})
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

	for {
		select {
		case <-srv.shutdown:
			return
		case <-timer.C:
			if err := srv.write(); err != nil {
				srv.logger.Println("write value fail", err)
			} else {
				writed = true
			}

			if value, err := srv.read(); err != nil {
				srv.logger.Println("read value fail", err)
			} else {
				writed = true

				srv.cached.Store(&apartResult{
					value: value,
					err:   err,
				})
			}

			if writed {
				timer.Reset(5 * time.Minute)
			} else {
				timer.Reset(10 * time.Second)
			}
		}
	}
}
