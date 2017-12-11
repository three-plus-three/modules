package environment

import (
	"fmt"
	"log"
	"net"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/three-plus-three/modules/as"
)

func (env *Environment) initTSDB(printIfFilesNotFound bool) error {
	var tsdbConfigFile string
	var tsdbConfig map[string]interface{}

	if runtime.GOOS == "windows" {
		tsdbConfigFile = env.Fs.FromConfig("tsdb_config.win.conf")
	} else {
		tsdbConfigFile = env.Fs.FromConfig("tsdb_config.conf")
	}

	_, err := toml.DecodeFile(tsdbConfigFile, &tsdbConfig)
	if err != nil {
		if printIfFilesNotFound {
			log.Println("[warn] load tsdb config fail from", tsdbConfigFile)
		}
	} else if nil != tsdbConfig {
		tsdbHTTP, _ := as.Object(tsdbConfig["http"])
		if nil != tsdbHTTP {
			if _, port, err := net.SplitHostPort(fmt.Sprint(tsdbHTTP["bind-address"])); err == nil {
				env.GetServiceConfig(ENV_INFLUXDB_PROXY_ID).SetPort(port)
				if printIfFilesNotFound {
					log.Println("[warn] load tsdb http port ("+port+") from", tsdbConfigFile)
				}
			}
		}

		tsdbAdmin, _ := as.Object(tsdbConfig["admin"])
		if nil != tsdbAdmin {
			if _, port, err := net.SplitHostPort(fmt.Sprint(tsdbAdmin["bind-address"])); err == nil {
				env.GetServiceConfig(ENV_INFLUXDB_ADM_PROXY_ID).SetPort(port)
				if printIfFilesNotFound {
					log.Println("[warn] load tsdb admin port ("+port+") from", tsdbConfigFile)
				}
			}
		}
	}
	return nil
}
