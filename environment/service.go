package environment

import (
	"net"
	"net/url"
	"os"

	"github.com/three-plus-three/modules/netutil"
	"github.com/three-plus-three/modules/urlutil"
)

var UnknownServiceConfig = &ServiceConfig{ID: ENV_MAX_PROXY_ID}

// ServiceConfig 服务的配置
type ServiceConfig struct {
	env     *Environment
	ID      ENV_PROXY_TYPE
	Name    string
	IsSSL   bool
	Type    string
	Host    string
	Port    string
	UrlPath string

	// surl atomic.Value
}

func (cfg *ServiceConfig) copyFrom(src *ServiceConfig) {
	cfg.env = src.env
	cfg.ID = src.ID
	cfg.IsSSL = src.IsSSL
	cfg.Type = src.Type
	cfg.Name = src.Name
	cfg.Host = src.Host
	cfg.Port = src.Port
	cfg.UrlPath = src.UrlPath
}

func (sc *ServiceConfig) loadConfig(cfg map[string]string, so ServiceOption) {
	sc.ID = so.ID
	sc.Name = so.Name

	sc.Type = stringWith(cfg, so.Name+".type", so.Type)
	sc.IsSSL = boolWith(cfg, so.Name+".is_ssl", so.IsSSL)

	switch so.ID {
	case ENV_HOME_PROXY_ID:
		sc.Host = hostWith(cfg, so.Name+".host", stringWith(cfg, "daemon.host", so.Host))
		sc.Port = portWith(cfg, so.Name+".port", portWith(cfg, "daemon.port", so.Port))
		//	case ENV_GATEWAY_PROXY_ID:
		//		sc.Host = hostWith(cfg, so.Name+".host", stringWith(cfg, "daemon.host", so.Host))
		//		sc.Port = portWith(cfg, so.Name+".port", portWith(cfg, "daemon.port", so.Port))
	case ENV_MC_DEV_PROXY_ID:
		if mcDevPort := os.Getenv("mc_dev_port"); "" != mcDevPort {
			sc.Port = mcDevPort
		}
	default:
		sc.Host = hostWith(cfg, so.Name+".host", so.Host)
		sc.Port = portWith(cfg, so.Name+".port", so.Port)
	}
}

// ListenAddr 服务的监听地址
func (cfg *ServiceConfig) ListenAddr(typ, pa string) (string, string) {
	if cfg.ID >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}
	if typ == "" {
		typ = cfg.Type
	}

	if typ == "" || typ == "auto" {
		if cfg.env.EnabledPipe() {
			typ = "unix"
		} else {
			typ = "tcp"
		}
	}

	if pa != "" {
		if netutil.IsUnixsocket(typ) {
			_, port, err := net.SplitHostPort(pa)
			if err == nil {
				return typ, netutil.MakePipename(port)
			}
		}

		return typ, pa
	}

	if netutil.IsUnixsocket(typ) {
		return typ, netutil.MakePipename(cfg.Port)
	}

	listenAddress := cfg.env.Config.StringWithDefault("listen_address", "")
	return typ, net.JoinHostPort(listenAddress, cfg.Port)
}

// ListenAddr 服务的连接地址
func (cfg *ServiceConfig) RemoteAddr(typ, pa string) (string, string) {
	if cfg.ID >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}
	if pa != "" {
		if _, _, e := net.SplitHostPort(pa); e == nil {
			return typ, pa
		}

		return typ, net.JoinHostPort(pa, cfg.Port)
	}

	typ = cfg.Type

	if typ == "" || typ == "auto" {
		if cfg.env.EnabledPipe() {
			typ = "unix"
		} else {
			typ = "tcp"
		}
	}

	//	if engine := cfg.env.GetEngineConfig(); engine.IsEnabled && !engine.IsMasterHost {
	//		host := engine.RemoteHost
	//		port := cfg.Port
	//		typ = "tcp"

	//		if remotePort := engine.RemotePort; "" != remotePort && "0" != remotePort {
	//			port = remotePort
	//		}
	//		return typ, net.JoinHostPort(host, port)
	//	}

	if netutil.IsUnixsocket(typ) {
		return typ, netutil.MakePipename(cfg.Port)
	}
	if "" == cfg.Host {
		return typ, net.JoinHostPort("127.0.0.1", cfg.Port)
	}
	return typ, net.JoinHostPort(cfg.Host, cfg.Port)
}

// SetHost 指定地址
func (cfg *ServiceConfig) SetHost(s string) {
	cfg.Host = s
	// 	cfg.surl.Store("")
	// 	cfg.Notify()
}

// SetPort 指定端口
func (cfg *ServiceConfig) SetPort(s string) {
	cfg.Port = s
	//cfg.surl.Store("")
	//cfg.Notify()
}

// UrlFor 服务的访问 URL
func (cfg *ServiceConfig) URLFor(s ...string) string {
	if cfg.ID >= ENV_MAX_PROXY_ID {
		panic("unknow service")
	}

	// if sp := cfg.surl.Load(); nil != sp {
	// 	if s, ok := sp.(string); ok && "" != s {
	// 		return s
	// 	}
	// }

	isSSL := cfg.IsSSL
	host := cfg.Host
	port := cfg.Port

	if cfg.Type == "" || cfg.Type == "auto" {
		if cfg.env.EnabledPipe() {
			host = netutil.UNIXSOCKET
		}
	} else if netutil.IsUnixsocket(cfg.Type) {
		host = netutil.UNIXSOCKET
	}

	if engine := cfg.env.GetEngineConfig(); engine.IsEnabled && !engine.IsMasterHost {
		host = engine.RemoteHost

		if remotePort := engine.RemotePort; "" != remotePort && "0" != remotePort {
			isSSL = engine.IsSSL
			port = remotePort
		}
	}

	var baseUrl string
	var scheme = "http://"
	if isSSL || ENV_LCN_PROXY_ID == cfg.ID {
		scheme = "https://"
	}
	if "" == cfg.UrlPath {
		baseUrl = scheme + net.JoinHostPort(host, port)
	} else {
		baseUrl = scheme + net.JoinHostPort(host, port) + "/" + cfg.UrlPath
	}

	return urlutil.JoinWith(baseUrl, s)
}

// UrlFor 服务的访问 URL
func (cfg *ServiceConfig) URIFor(s ...string) *url.URL {
	if u, e := url.Parse(cfg.URLFor(s...)); nil != e {
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
