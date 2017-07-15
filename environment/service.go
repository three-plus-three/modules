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

type ServiceConfig struct {
	env  *Environment
	Id   ENV_PROXY_TYPE
	Name string
	Host string
	Port string
	Path string

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

func (cfg *ServiceConfig) RemoveAllListener() {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	cfg.listeners_mu.Lock()
	defer cfg.listeners_mu.Unlock()
	cfg.listeners.Init()
}

func (cfg *ServiceConfig) ListenAddr(s string) string {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}
	if s != "" {
		return s
	}
	return ":" + cfg.Port
}

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

func (cfg *ServiceConfig) SetHost(s string) {
	cfg.Host = s
	cfg.surl.Store("")
	cfg.Notify()
}

func (cfg *ServiceConfig) SetPort(s string) {
	cfg.Port = s
	cfg.surl.Store("")
	cfg.Notify()
}

func (cfg *ServiceConfig) SetPath(s string) {
	cfg.Path = s
	cfg.surl.Store("")
	cfg.Notify()
}

func (cfg *ServiceConfig) SetUrl(s string) {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	cfg.surl.Store(s)
	cfg.Notify()
}

func (cfg *ServiceConfig) Url() string {
	if cfg.Id >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	if sp := cfg.surl.Load(); nil != sp {
		if s, ok := sp.(string); ok && "" != s {
			return s
		}
	}

	host := cfg.Host
	port := cfg.Port

	if engine := cfg.env.GetEngineConfig(); engine.IsEnabled && !engine.IsMasterHost {
		host = engine.RemoteHost

		if remotePort := engine.RemotePort; "" != remotePort && "0" != remotePort {
			port = remotePort
		}
	}

	var s string
	if ENV_LCN_PROXY_ID == cfg.Id {
		if "" == cfg.Path {
			s = "https://" + host + ":" + port
		} else {
			s = "https://" + host + ":" + port + "/" + cfg.Path
		}
	} else {
		if "" == cfg.Path {
			s = "http://" + host + ":" + port
		} else {
			s = "http://" + host + ":" + port + "/" + cfg.Path
		}
	}
	cfg.surl.Store(s)
	return s
}

func (cfg *ServiceConfig) UrlFor(s ...string) string {
	baseUrl := cfg.Url()
	return urlutil.JoinWith(baseUrl, s)
}

func (cfg *ServiceConfig) URI() *url.URL {
	s := cfg.Url()
	if u, e := url.Parse(s); nil != e {
		panic(e)
	} else {
		return u
	}
}

func (cfg *ServiceConfig) Client(paths ...string) HttpClient {
	return HttpClient{
		cfg:      cfg,
		basePath: urlutil.Join(paths...),
	}
}
