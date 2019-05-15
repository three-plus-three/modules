package netutil

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
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

const UNIXSOCKET = "unixsocket"

type HttpDialer struct {
	DialWithContext func(ctx context.Context, network, addr string) (net.Conn, error)
}

// Dial connects to the address on the named network.
func (d *HttpDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func WrapDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return (&HttpDialer{DialWithContext: dialContext}).DialContext
}

type Dialer struct {
	net.Dialer
}
