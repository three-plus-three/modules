// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket2

import (
	"crypto/tls"
	"net"
	"strings"
	"time"
)

func dialWithDialer(dial func(network, address string) (net.Conn, error), config *Config) (conn net.Conn, err error) {
	switch config.Location.Scheme {
	case "ws":
		conn, err = dial("tcp", parseAuthority(config.Location))

	case "wss":
		var rawConn net.Conn
		rawConn, err = dial("tcp", parseAuthority(config.Location))
		if err != nil {
			return
		}

		conn, err = DialWithConn(rawConn, 5*time.Second, parseAuthority(config.Location), config.TlsConfig)
	default:
		err = ErrBadScheme
	}
	return
}

var emptyConfig tls.Config

func defaultConfig() *tls.Config {
	return &emptyConfig
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "tls: DialWithDialer timed out" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// DialWithDialer connects to the given network address using dialer.Dial and
// then initiates a TLS handshake, returning the resulting TLS connection. Any
// timeout or deadline given in the dialer apply to connection and TLS
// handshake as a whole.
//
// DialWithDialer interprets a nil configuration as equivalent to the zero
// configuration; see the documentation of Config for the defaults.
func DialWithConn(rawConn net.Conn, timeout time.Duration, addr string, config *tls.Config) (*tls.Conn, error) {
	// We want the Timeout and Deadline values from dialer to cover the
	// whole process: TCP connection and TLS handshake. This means that we
	// also need to start our own timers now.
	//	timeout := dialer.Timeout
	//
	//	if !dialer.Deadline.IsZero() {
	//		deadlineTimeout := time.Until(dialer.Deadline)
	//		if timeout == 0 || deadlineTimeout < timeout {
	//			timeout = deadlineTimeout
	//		}
	//	}

	//	rawConn, err := dialer.Dial(network, addr)
	//	if err != nil {
	//		return nil, err
	//	}

	colonPos := strings.LastIndex(addr, ":")
	if colonPos == -1 {
		colonPos = len(addr)
	}
	hostname := addr[:colonPos]

	if config == nil {
		config = defaultConfig()
	}
	// If no ServerName is set, infer the ServerName
	// from the hostname we're connecting to.
	if config.ServerName == "" {
		// Make a copy to avoid polluting argument or default.
		c := config.Clone()
		c.ServerName = hostname
		config = c
	}

	conn := tls.Client(rawConn, config)

	var err error
	if timeout == 0 {
		err = conn.Handshake()
	} else {

		errChannel := make(chan error, 2)
		time.AfterFunc(timeout, func() {
			errChannel <- timeoutError{}
		})

		go func() {
			errChannel <- conn.Handshake()
		}()

		err = <-errChannel
	}

	if err != nil {
		rawConn.Close()
		return nil, err
	}

	return conn, nil
}
