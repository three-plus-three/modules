package httputil

import (
	"net"
	"net/http"
	"time"

	"github.com/three-plus-three/modules/netutil"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type TcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln TcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func RunTLS(network, addr, certFile, keyFile string, engine http.Handler) (err error) {
	if network == "" {
		network = "tcp"
	}

	if !netutil.IsUnixsocket(network) {
		return http.ListenAndServeTLS(addr, certFile, keyFile, engine)
	}

	listener, err := netutil.NewUnixListener(network, addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	return http.ServeTLS(listener, engine, certFile, keyFile)
}

func RunHTTP(network, addr string, engine http.Handler) (err error) {
	if network == "" {
		network = "tcp"
	}

	if !netutil.IsUnixsocket(network) {
		return http.ListenAndServe(addr, engine)
	}

	listener, err := netutil.NewUnixListener(network, addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	return http.Serve(listener, engine)
}

func RunServer(srv *http.Server, network, addr string) error {
	if network == "" {
		network = "tcp"
	}

	if !netutil.IsUnixsocket(network) {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		return srv.Serve(TcpKeepAliveListener{ln.(*net.TCPListener)})
	}
	ln, err := netutil.NewUnixListener(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	return srv.Serve(ln)
}

func RunServerTLS(srv *http.Server, network, addr, certFile, keyFile string) error {
	if network == "" {
		network = "tcp"
	}

	if !netutil.IsUnixsocket(network) {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		return srv.ServeTLS(TcpKeepAliveListener{ln.(*net.TCPListener)}, certFile, keyFile)
	}

	ln, err := netutil.NewUnixListener(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	return srv.ServeTLS(ln, certFile, keyFile)
}
