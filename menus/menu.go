package menus

import (
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/hub"
	"github.com/three-plus-three/modules/toolbox"
)

// var menuStyle = flag.String("menuStyle", "standalone", "菜单的模式， 可选值：standalone，apart")

// Client 菜单服务
type Client interface {
	io.Closer

	Read() ([]toolbox.Menu, error)
}

// Callback 菜单的读取函数
type Callback func() ([]toolbox.Menu, error)

// Connect 的注册函数
func Connect(env *environment.Environment, appID int, cb Callback, mode string) Client {
	//wsrv := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)

	// hubURL := so.UrlFor(env.DaemonUrlPath, "/mq/")
	// builder := hub.Connect(hubURL)

	switch mode {
	case "apart":
		apart := &apartClient{
			logger:   log.New(os.Stderr, "[menus]", log.LstdFlags),
			env:      env,
			wsrv:     env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID),
			cb:       cb,
			shutdown: make(chan struct{}),
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

func (srv *standaloneClient) Read() ([]toolbox.Menu, error) {
	return srv.cb()
}

type apartResult struct {
	menuList []toolbox.Menu
	err      error
}

type apartClient struct {
	logger *log.Logger
	env    *environment.Environment
	wsrv   *environment.ServiceConfig
	cb     Callback

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

func (srv *apartClient) Read() ([]toolbox.Menu, error) {
	o := srv.cached.Load()
	if o != nil {
		if result, ok := o.(*apartResult); ok {
			return result.menuList, result.err
		}
	}

	menuList, err := srv.read()
	srv.cached.Store(&apartResult{
		menuList: menuList,
		err:      err,
	})
	return menuList, err
}

func (srv *apartClient) read() ([]toolbox.Menu, error) {
	var menuList []toolbox.Menu
	err := srv.wsrv.Client(srv.env.DaemonUrlPath, "/menu/").GET(&menuList)
	return menuList, err
}

func (srv *apartClient) write() error {
	menuList, err := srv.cb()
	if err != nil {
		return err
	}
	o := srv.cached.Load()
	if o != nil {
		if result, ok := o.(*apartResult); ok {
			if !toolbox.IsSameMenuArray(menuList, result.menuList) && len(menuList) != 0 {
				return nil
			}
		}
	}

	return srv.wsrv.Client(srv.env.DaemonUrlPath, "/menu/").POST(&menuList)
}

func (srv *apartClient) runSub() {
	errCount := 0
	hubURL := srv.wsrv.UrlFor(srv.env.DaemonUrlPath, "/mq/")
	builder := hub.Connect(hubURL)

	for atomic.LoadInt32(&srv.isClosed) == 0 {

		topic, err := builder.SubscribeTopic("menus.changed")
		if err != nil {
			errCount++
			if errCount%50 < 3 {
				srv.logger.Println("subscribe menus.changed fail,", err)
			}
			continue
		}
		srv.cw.Set(topic)

		errCount = 0
		err = topic.Run(func(sub *hub.Subscription, msg hub.Message) {
			menuList, err := srv.read()
			srv.cached.Store(&apartResult{
				menuList: menuList,
				err:      err,
			})
		})
		if err != nil {
			srv.logger.Println("subscribe menus.changed fail,", err)
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
				srv.logger.Println("write menu list fail", err)
			} else {
				writed = true
			}

			if menuList, err := srv.read(); err != nil {
				srv.logger.Println("write menu list fail", err)
			} else {
				writed = true

				srv.cached.Store(&apartResult{
					menuList: menuList,
					err:      err,
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
