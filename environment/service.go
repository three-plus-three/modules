package environment

import (
	"container/list"
	"io"
	"net"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/three-plus-three/modules/urlutil"
)

var UnknownServiceConfig = &ServiceConfig{Id: ENV_MAX_PROXY_ID}

// ServiceConfig 服务的配置
type ServiceConfig struct {
	env   *Environment
	Id    ENV_PROXY_TYPE
	Name  string
	IsSSL bool
	Host  string
	Port  string
	Path  string

	surl atomic.Value

	listeners_mu sync.Mutex
	listeners    list.List
}

func (cfg *ServiceConfig) copyFrom(src *ServiceConfig) {
	cfg.env = src.env
	cfg.Id = src.Id
	cfg.Name = src.Name
	cfg.Host = src.Host
	cfg.Port = src.Port
	cfg.Path = src.Path
	if o := src.surl.Load(); o != nil {
		cfg.surl.Store(o)
	}

	cfg.listeners.Init()

	src.listeners_mu.Lock()
	defer src.listeners_mu.Unlock()
	cfg.listeners.PushBackList(&src.listeners)
}

// Notify 发送变动通知
func (cfg *ServiceConfig) Notify() {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	cfg.listeners_mu.Lock()
	defer cfg.listeners_mu.Unlock()
	for current := cfg.listeners.Front(); current != nil; current = current.Next() {
		if nil == current.Value {
			continue
		}

		if cb, ok := current.Value.(func(*ServiceConfig)); ok {
			cb(cfg)
		}
	}
}

type serviceListener struct {
	cfg *ServiceConfig
	el  *list.Element
}

func (s serviceListener) Close() error {
	s.cfg.listeners_mu.Lock()
	defer s.cfg.listeners_mu.Unlock()

	s.cfg.listeners.Remove(s.el)
	return nil
}

// On 注册变动事件
func (cfg *ServiceConfig) On(cb func(*ServiceConfig)) io.Closer {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	cb(cfg)
	cfg.listeners_mu.Lock()
	defer cfg.listeners_mu.Unlock()

	el := cfg.listeners.PushBack(cb)
	return serviceListener{cfg, el}
}

// RemoveAllListener 清空所有监听器
func (cfg *ServiceConfig) RemoveAllListener() {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	cfg.listeners_mu.Lock()
	defer cfg.listeners_mu.Unlock()
	cfg.listeners.Init()
}

// ListenAddr 服务的监听地址
func (cfg *ServiceConfig) ListenAddr(s string) string {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}
	if s != "" {
		return s
	}

	listenAddress := cfg.env.Config.StringWithDefault("listen_address", "")
	return net.JoinHostPort(listenAddress, cfg.Port)
}

// ListenAddr 服务的连接地址
func (cfg *ServiceConfig) RemoteAddr(s string) string {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}
	if s != "" {
		if _, _, e := net.SplitHostPort(s); nil == e {
			return s
		}
		return net.JoinHostPort(s, cfg.Port)
	}
	if "" == cfg.Host {
		return "127.0.0.1:" + cfg.Port
	}
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

// SetHost 指定地址
func (cfg *ServiceConfig) SetHost(s string) {
	cfg.Host = s
	cfg.surl.Store("")
	cfg.Notify()
}

// SetPort 指定端口
func (cfg *ServiceConfig) SetPort(s string) {
	cfg.Port = s
	cfg.surl.Store("")
	cfg.Notify()
}

// SetPath 指定路径
func (cfg *ServiceConfig) SetPath(s string) {
	cfg.Path = s
	cfg.surl.Store("")
	cfg.Notify()
}

// SetUrl 指定 URL
func (cfg *ServiceConfig) SetUrl(s string) {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	cfg.surl.Store(s)
	cfg.Notify()
}

// Url 服务的访问 URL
func (cfg *ServiceConfig) Url() string {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	if sp := cfg.surl.Load(); nil != sp {
		if s, ok := sp.(string); ok && "" != s {
			return s
		}
	}

	isSSL := cfg.IsSSL
	host := cfg.Host
	port := cfg.Port

	if engine := cfg.env.GetEngineConfig(); engine.IsEnabled && !engine.IsMasterHost {
		host = engine.RemoteHost

		if remotePort := engine.RemotePort; "" != remotePort && "0" != remotePort {
			isSSL = engine.IsSSL
			port = remotePort
		}
	}

	var s string
	if isSSL || ENV_LCN_PROXY_ID == cfg.Id {
		if "" == cfg.Path {
			s = "https://" + net.JoinHostPort(host, port)
		} else {
			s = "https://" + net.JoinHostPort(host, port) + "/" + cfg.Path
		}
	} else {
		if "" == cfg.Path {
			s = "http://" + net.JoinHostPort(host, port)
		} else {
			s = "http://" + net.JoinHostPort(host, port) + "/" + cfg.Path
		}
	}
	cfg.surl.Store(s)
	return s
}

// UrlFor 服务的访问 URL
func (cfg *ServiceConfig) UrlFor(s ...string) string {
	baseUrl := cfg.Url()
	return urlutil.JoinWith(baseUrl, s)
}

// UrlFor 服务的访问 URL
func (cfg *ServiceConfig) URI() *url.URL {
	s := cfg.Url()
	if u, e := url.Parse(s); nil != e {
		panic(e)
	} else {
		return u
	}
}

// Client 服务的访问
func (cfg *ServiceConfig) Client(paths ...string) *HttpClient {
	return &HttpClient{
		cfg:      cfg,
		basePath: urlutil.Join(paths...),
	}
}
