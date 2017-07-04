package environment

import (
	"encoding/json"
	"os"

	"log"
)

func loadMinioConfig(fs FileSystem) map[string]interface{} {
	configFile := fs.FromData(".minio", "config.json")
	if !FileExists(configFile) {
		log.Println("[warn] '" + configFile + "' isn't exists.")
		return nil
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