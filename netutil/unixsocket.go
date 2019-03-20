package netutil

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	winio "github.com/Microsoft/go-winio"
)

var isWindows = runtime.GOOS == "windows"

func IsUnixsocket(network string) bool {
	return network == "unix"
}

var PipeDir string = os.Getenv("tpt_unix_socket_dir")

func init() {
	if PipeDir == "" {
		if isWindows {
			PipeDir = `\\.\pipe\`
		} else {
			PipeDir = os.TempDir()
		}
	}
}

func MakePipenameWindows(port string) string {
	return PipeDir + `hengwei_` + port
}

func MakePipenameUnix(port string) string {
	return filepath.Join(PipeDir, `hengwei_`+port+".sock")
}

func MakePipename(port string) string {
	if isWindows {
		return MakePipenameWindows(port)
	}
	return MakePipenameUnix(port)
}

func NewUnixListener(network, file string) (net.Listener, error) {
	if isWindows {
		if !strings.HasPrefix(file, `\\`) {
			file = MakePipenameWindows(file)
		}

		listener, err := winio.ListenPipe(file, nil)
		if err != nil {
			return nil, err
		}
		return listener, nil
	}

	if !filepath.IsAbs(file) {
		file = MakePipenameUnix(file)
	}

	if err := os.Remove(file); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	listener, err := net.Listen(network, file)
	if err != nil {
		return nil, err
	}
	os.Chmod(file, 0777)
	return listener, nil
}

const UNIXSOCKET = "unixsocket"

func Dial(network, addr string) (net.Conn, error) {
	if IsUnixsocket(network) && isWindows {
		return winio.DialPipe(addr, nil)
	}

	return net.Dial(network, addr)
}

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if IsUnixsocket(network) && isWindows {
		return winio.DialPipeContext(ctx, addr)
	}

	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

func DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	if IsUnixsocket(network) && isWindows {
		return winio.DialPipe(addr, &timeout)
	}

	return net.DialTimeout(network, addr, timeout)
}

type HttpDialer struct {
	DialWithContext func(ctx context.Context, network, addr string) (net.Conn, error)
}

// Dial connects to the address on the named network.
func (d *HttpDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (dialer *HttpDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = "80"
	}

	if host == UNIXSOCKET {
		if isWindows {
			pipeName := MakePipenameWindows(port)
			return winio.DialPipeContext(ctx, pipeName)
		}

		network = "unix"
		addr = MakePipenameUnix(port)
	}

	if dialer.DialWithContext == nil {
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, addr)
	}
	return dialer.DialWithContext(ctx, network, addr)
}

func WrapDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return (&HttpDialer{DialWithContext: dialContext}).DialContext
}

type Dialer struct {
	net.Dialer
}

// Dial connects to the address on the named network.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	if IsUnixsocket(network) && isWindows {
		return winio.DialPipe(address, nil)
	}

	return d.Dialer.DialContext(context.Background(), network, address)
}

// DialContext connects to the address on the named network using
// the provided context.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if IsUnixsocket(network) && isWindows {
		return winio.DialPipeContext(ctx, address)
	}
	return d.Dialer.DialContext(ctx, network, address)
}
