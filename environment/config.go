package environment

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	commons_cfg "github.com/three-plus-three/modules/cfg"
)

func LoadConfigs(fs FileSystem, nm string, dumpFilesIfNotFound bool) map[string]string {
	files := fs.SearchConfig(nm)
	if 0 == len(files) {
		if dumpFilesIfNotFound {
			var buffer bytes.Buffer
			buffer.WriteString("[warn] properties file is not exists, search path is:\r\n")
			buffer.WriteString("    ")
			buffer.WriteString(fs.FromConfig(""))
			buffer.WriteString("\r\n    ")
			buffer.WriteString(fs.FromDataConfig(""))
			buffer.WriteString("\r\n")
			log.Print(buffer.String())
		}
		return nil
	}

	var cfg map[string]string
	for _, file := range files {
		props, e := commons_cfg.ReadProperties(file)
		if nil != e {
			if dumpFilesIfNotFound {
				log.Println("[warn] read properties '" + file + "' failed," + e.Error())
			}
			return nil
		}

		if dumpFilesIfNotFound {
			log.Println("[info] load properties '" + file + "'.")
		}
		if nil == cfg {
			cfg = props
		} else {
			for k, v := range props {
				cfg[k] = v
			}
		}
	}
	return cfg
}

// func LoadConfig(fs FileSystem, nm string, flagSet *flag.FlagSet) error {
// 	return LoadConfigsWith(fs, nm, flagSet, true)
// }

func ReadConfigs(fs FileSystem, files []string, nm string, dumpFilesIfNotFound bool) map[string]string {
	cfg := LoadConfigs(fs, "app.properties", dumpFilesIfNotFound)
	if nil == cfg {
		cfg = map[string]string{}
	}

	if len(files) > 0 {
		for _, file := range files {
			if !filepath.IsAbs(file) {
				file = fs.FromDataConfig(file)
			}

			props, e := commons_cfg.ReadProperties(file)
			if e != nil {
				if dumpFilesIfNotFound || !os.IsNotExist(e) {
					log.Println("[warn] load config fail: ", e)
				}
				continue
			}

			log.Println("[info] load config ok: ", file)
			if len(props) > 0 {
				for k, v := range props {
					cfg[k] = v
				}
			}
		}
	}

	localCfg := LoadConfigs(fs, nm+".properties", dumpFilesIfNotFound)
	if nil != localCfg {
		for k, v := range localCfg {
			cfg[k] = v
		}
	}
	engineCfg := LoadConfigs(fs, "engine.properties", dumpFilesIfNotFound)
	if nil != engineCfg {
		for k, v := range engineCfg {
			cfg[k] = v
		}
	}
	return cfg
}

// func LoadConfigsWith(fs FileSystem, nm string, flagSet *flag.FlagSet, dumpFilesIfNotFound bool) error {
// 	return InitConfig(flagSet, ReadConfigs(fs, nm, dumpFilesIfNotFound), dumpFilesIfNotFound)
// }

var db_defaults = map[string]string{"redis.host": "127.0.0.1",
	"redis.port":       "36379",
	"db.type":          "postgresql",
	"db.address":       "127.0.0.1",
	"db.port":          "35432",
	"data.db.schema":   "tpt_data",
	"models.db.schema": "tpt",
	"db.username":      "tpt",
	"db.password":      "extreme"}

func InitConfig(flagSet *flag.FlagSet, cfg map[string]string, dumpFilesIfNotFound bool) error {
	if nil == flagSet {
		flagSet = flag.CommandLine
	}
	//commons_registry.LoadRegistry(cfg)
	//commons_registry.LoadEngineRegistry(cfg)

	actual := map[string]string{}
	flagSet.Visit(func(f *flag.Flag) {
		actual[f.Name] = f.Name
	})

	formul := map[string]string{}
	flagSet.VisitAll(func(f *flag.Flag) {
		formul[f.Name] = f.DefValue
	})

	for k, k_var := range formul {
		if _, ok := actual[k]; ok {
			continue
		}

		switch k {
		case "daemon":
			daemon_address := stringWith(cfg, "daemon.host", "127.0.0.1")
			daemon_port := stringWith(cfg, "daemon.port", "37072")
			flagSet.Set("daemon", "http://"+daemon_address+":"+daemon_port) // nolint
		case "redis_address", "redis":
			redis_address := stringWith(cfg, "redis.host", db_defaults["redis.host"])
			redis_port := stringWith(cfg, "redis.port", db_defaults["redis.port"]) // nolint
			flagSet.Set(k, redis_address+":"+redis_port)
		case "data_db.url", "db_url": // for tsdb
			drv, url, e := CreateDBUrl("data.", cfg, db_defaults)
			if nil != e {
				return e
			}

			url_name := "data_db.url"
			drv_name := "data_db.driver"
			if "db_url" == k {
				url_name = "db_url"
				drv_name = "db_driver"
			}

			flagSet.Set(url_name, url) // nolint
			flagSet.Set(drv_name, drv) // nolint
		case "db.url": // for ds
			drv, url, e := CreateDBUrl("models.", cfg, db_defaults)
			if nil != e {
				return e
			}

			flagSet.Set("db.url", url)    // nolint
			flagSet.Set("db.driver", drv) // nolint
		case "ds.listen": // for db
			if v, ok := cfg["models.port"]; ok {
				flagSet.Set(k, ":"+v) // nolint
			}
		case "tsdb.listen": // for tsdb
			if v, ok := cfg["repo.port"]; ok {
				flagSet.Set(k, ":"+v) // nolint
			}
		case "sampling.listen": // for sampling
			if v, ok := cfg["sampling.port"]; ok {
				flagSet.Set(k, ":"+v) // nolint
			}
		case "ds.url": // for sampling
			if v, ok := cfg["models.port"]; ok {
				host, ok := cfg["models.host"]
				if !ok {
					host = "127.0.0.1"
				}

				flagSet.Set(k, "http://"+host+":"+v) // nolint
			}

		case "ds": // for poller
			if v, ok := cfg["models.port"]; ok {
				host, ok := cfg["models.host"]
				if !ok {
					host = "127.0.0.1"
				}

				flagSet.Set(k, "http://"+host+":"+v) // nolint
			}

		case "sampling": // for poller
			if v, ok := cfg["sampling.port"]; ok {
				host, ok := cfg["sampling.host"]
				if !ok {
					host = "127.0.0.1"
				}

				flagSet.Set(k, "http://"+host+":"+v+"/batch") // nolint
			}

		case "tsdb.url": // for poller
			if v, ok := cfg["repo.port"]; ok {
				host, ok := cfg["repo.host"]
				if !ok {
					host = "127.0.0.1"
				}

				flagSet.Set(k, "http://"+host+":"+v) // nolint
			}
		case "listen": // delayed_jobs
			if ":39086" == k_var {
				if v, ok := cfg["delayed_jobs.port"]; ok {
					flagSet.Set(k, ":"+v) // nolint
				}
			}
		default:
			if v, ok := cfg[k]; ok {
				flagSet.Set(k, v) // nolint
			}
		}
	}
	return nil
}

func IsSetFlagVar(name string) (ret bool) {
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			ret = true
		}
	})
	return ret
}

func stringWith(props map[string]string, nm, defaultValue string) string {
	if v, ok := props[nm]; ok && 0 != len(v) {
		return v
	}
	return defaultValue
}

func ReadDbConfig(prefix string, props, defaultValues map[string]string) DbConfig {
	db_type := stringWith(props, prefix+"db.type", stringWith(defaultValues, prefix+"db.type", stringWith(defaultValues, "db.type", "")))
	db_address := stringWith(props, prefix+"db.address", stringWith(defaultValues, prefix+"db.address", stringWith(defaultValues, "db.address", "")))
	db_port := stringWith(props, prefix+"db.port", stringWith(defaultValues, prefix+"db.port", stringWith(defaultValues, "db.port", "")))
	db_schema := stringWith(props, prefix+"db.schema", stringWith(defaultValues, prefix+"db.schema", stringWith(defaultValues, "db.schema", "")))
	db_username := stringWith(props, prefix+"db.username", stringWith(defaultValues, prefix+"db.username", stringWith(defaultValues, "db.username", "")))
	db_password := stringWith(props, prefix+"db.password", stringWith(defaultValues, prefix+"db.password", stringWith(defaultValues, "db.password", "")))

	return DbConfig{
		DbType:   db_type,
		Address:  db_address,
		Port:     db_port,
		Schema:   db_schema,
		Username: db_username,
		Password: db_password,
	}
}

func CreateDBUrl(prefix string, props, defaultValues map[string]string) (string, string, error) {
	dbConfig := ReadDbConfig(prefix, props, defaultValues)
	switch dbConfig.DbType {
	case "postgresql":
		return "postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			dbConfig.Address, dbConfig.Port, dbConfig.Schema, dbConfig.Username, dbConfig.Password), nil
	default:
		return "", "", errors.New("unknown db type - " + dbConfig.DbType)
	}
}

func LoadConfigFromJsonFile(nm string, flagSet *flag.FlagSet, isOverride bool) error {
	if nil == flagSet {
		flagSet = flag.CommandLine
	}

	f, e := os.Open(nm)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}

	var res map[string]interface{}
	e = json.NewDecoder(f).Decode(&res)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}

	actual := map[string]string{}
	flagSet.Visit(func(f *flag.Flag) {
		actual[f.Name] = f.Name
	})

	e = assignFlagSet("", res, flagSet, actual, isOverride)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}
	return nil
}

func assignFlagSet(prefix string, res map[string]interface{}, flagSet *flag.FlagSet, actual map[string]string, isOverride bool) error {
	for k, v := range res {
		switch value := v.(type) {
		case map[string]interface{}:
			e := assignFlagSet(combineName(prefix, k), value, flagSet, actual, isOverride)
			if nil != e {
				return e
			}
			continue
		case []interface{}:
		case string:
		case float64:
		case bool:
		case nil:
			continue
		default:
			return fmt.Errorf("unsupported type for %s - %T", combineName(prefix, k), v)
		}
		nm := combineName(prefix, k)

		if !isOverride {
			if _, ok := actual[nm]; ok {
				log.Printf("load flag '%s' from config is skipped.\n", nm)
				continue
			}
		}

		var g *flag.Flag = flagSet.Lookup(nm)
		if nil == g {
			log.Printf("flag '%s' is not defined.\n", nm)
			continue
		}

		err := g.Value.Set(fmt.Sprint(v))
		if nil != err {
			return err
		}
	}
	return nil
}

func combineName(prefix, nm string) string {
	if "" == prefix {
		return nm
	}
	return prefix + "." + nm
}

func SetFlags(cfg map[string]string, flagSet *flag.FlagSet, isOverride bool) {
	actual := map[string]string{}
	flags := make([]*flag.Flag, 0, 10)
	if nil == flagSet {
		flagSet = flag.CommandLine
	}
	if !isOverride {
		flagSet.Visit(func(g *flag.Flag) {
			actual[g.Name] = g.Name
		})
	}
	flagSet.VisitAll(func(g *flag.Flag) {
		if isOverride {
			flags = append(flags, g)
		} else if _, ok := actual[g.Name]; !ok {
			flags = append(flags, g)
		}
	})
	for _, g := range flags {
		if v, ok := cfg[g.Name]; ok {
			flagSet.Set(g.Name, v)
		}
	}
}

func IsFlagInitialized(name string) bool {
	ret := false
	flag.Visit(func(f *flag.Flag) {
		if name == f.Name {
			ret = true
		}
	})
	return ret
}
