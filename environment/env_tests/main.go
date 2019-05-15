package env_tests

import (
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/netutil"
	"github.com/three-plus-three/modules/util"
)

var EnvFile = flag.String("env_file", "test_postgres.properties", "")

func Clone(env *environment.Environment) *environment.Environment {
	var copyed *environment.Environment
	if env == nil {
		file := *EnvFile
		if !filepath.IsAbs(file) {
			var files = []string{file}

			paths := filepath.SplitList(os.Getenv("GOPATH"))
			for _, pa := range paths {
				files = append(files, filepath.Join(pa, file))
				files = append(files, filepath.Join(pa, "src/cn/com/hengwei/commons/env_tests", file))
			}

			for _, s := range files {
				if util.FileExists(s) {
					file = s
					break
				}
			}
		}

		log.Println("load --", file)

		var err error
		opt := environment.Options{Name: "env_test",
			ConfigFiles: []string{file},
			IsTest:      true}
		env, err = environment.NewEnvironment(opt)
		if nil != err {
			panic(err)
		}

		copyed = env
		//		for id := environment.ENV_MIN_PROXY_ID; id < environment.ENV_MAX_PROXY_ID; id++ {
		//			copyed.GetServiceConfig(id).SetUrl("")
		//		}
	} else {
		copyed = env.Clone()
	}
	// copyed.RemoveAllListener()
	return copyed
}

func SetURL(t testing.TB, cfg *environment.ServiceConfig, urlStr string) {
	t.Helper()

	if urlStr == "" {
		cfg.UrlPath = ""
		return
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		t.Fatal(err)
	}

	cfg.Type = "tcp"
	cfg.Host = u.Host
	if host, port, err := net.SplitHostPort(u.Host); err == nil {
		cfg.Host = host
		cfg.Port = port
		if host == netutil.UNIXSOCKET {
			cfg.Type = "unix"
		}
	} else if u.Scheme == "http" {
		cfg.Port = "80"
	} else if u.Scheme == "https" {
		cfg.Port = "443"
	}

	cfg.UrlPath = u.Path
	if cfg.UrlPath == "/" {
		cfg.UrlPath = ""
	}
}

var AbsoluteToImport = util.AbsoluteToImport
var ImportToAbsolute = util.ImportToAbsolute
var CleanImport = util.CleanImport
