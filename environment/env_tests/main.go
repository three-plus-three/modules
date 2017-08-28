package env_tests

import (
	"cn/com/hengwei/commons"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/three-plus-three/modules/environment"
)

var DBURL = flag.String("db.url", "host=127.0.0.1 dbname=tpt_models_test user=tpt password=extreme sslmode=disable", "the db url")
var DBDRV = flag.String("db.driver", "postgres", "the db driver")
var data_db_url = flag.String("data_db.url", "host=127.0.0.1 dbname=tpt_data_test user=tpt password=extreme sslmode=disable", "the db url")
var data_db_drv = flag.String("data_db.driver", "postgres", "the db driver")

var env_file = flag.String("env_file", "test_postgres.properties", "")

func Clone(env *environment.Environment) *environment.Environment {
	var copyed *environment.Environment
	if env == nil {
		file := *env_file
		if !filepath.IsAbs(file) {
			var files = []string{file}

			paths := filepath.SplitList(os.Getenv("GOPATH"))
			for _, pa := range paths {
				files = append(files, filepath.Join(pa, file))
				files = append(files, filepath.Join(pa, "src/cn/com/hengwei/commons/env_tests", file))
			}

			for _, s := range files {
				if commons.FileExists(s) {
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
		for id := environment.ENV_MIN_PROXY_ID; id < environment.ENV_MAX_PROXY_ID; id++ {
			copyed.GetServiceConfig(id).SetUrl("")
		}
	} else {
		copyed = env.Clone()
	}
	copyed.RemoveAllListener()
	return copyed
}
