package environment

import (
	"encoding/json"
	"log"
	"os"

	"github.com/three-plus-three/modules/util"
)

func loadMinioConfig(fs FileSystem) map[string]interface{} {
	configFile := fs.FromData("minio", ".minio.sys", "config", "config.json")
	if !util.FileExists(configFile) {

		configFile2 := fs.FromData(".minio", "config.json")
		if !util.FileExists(configFile2) {
			log.Println("[warn] '" + configFile + "' isn't exists.")
			log.Println("[warn] '" + configFile2 + "' isn't exists.")
			return nil
		}
		configFile = configFile2
	}

	r, err := os.Open(configFile)
	if err != nil {
		log.Fatalln("[error] read '"+configFile+"' fail,", err)
	}
	defer r.Close()

	var config map[string]interface{}
	if err := json.NewDecoder(r).Decode(&config); err != nil {
		log.Fatalln("[error] read '"+configFile+"' fail,", err)
	}

	return config
}
