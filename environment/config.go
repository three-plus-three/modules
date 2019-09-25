package environment

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/three-plus-three/modules/as"
	commons_cfg "github.com/three-plus-three/modules/cfg"
)

// Config 配置
type Config struct {
	settings map[string]interface{}
}

// PasswordWithDefault 读配置
func (self *Config) PasswordWithDefault(key, defValue string) string {
	if s, ok := self.settings[key]; ok {
		return as.StringWithDefault(s, defValue)
	}
	return defValue
}

// StringWithDefault 读配置
func (self *Config) StringWithDefault(key, defValue string) string {
	if s, ok := self.settings[key]; ok {
		return as.StringWithDefault(s, defValue)
	}
	return defValue
}

// StringsWithDefault 读配置
func (self *Config) StringsWithDefault(key string, defValue []string) []string {
	if o, ok := self.settings[key]; ok {
		if s, ok := o.(string); ok {
			return strings.Split(s, ",")
		}
		return as.StringsWithDefault(o, defValue)
	}
	return defValue
}

// UintWithDefault 读配置
func (self *Config) UintWithDefault(key string, defValue uint) uint {
	if s, ok := self.settings[key]; ok {
		return as.UintWithDefault(s, defValue)
	}
	return defValue
}

// IntWithDefault 读配置
func (self *Config) IntWithDefault(key string, defValue int) int {
	if s, ok := self.settings[key]; ok {
		return as.IntWithDefault(s, defValue)
	}
	return defValue
}

// BoolWithDefault 读配置
func (self *Config) BoolWithDefault(key string, defValue bool) bool {
	if s, ok := self.settings[key]; ok {
		return as.BoolWithDefault(s, defValue)
	}
	return defValue
}

// DurationWithDefault 读配置
func (self *Config) DurationWithDefault(key string, defValue time.Duration) time.Duration {
	if s, ok := self.settings[key]; ok {
		return as.DurationWithDefault(s, defValue)
	}
	return defValue
}

// Set 写配置
func (self *Config) Set(key string, value interface{}) {
	self.settings[key] = value
}

// Get 读配置
func (self *Config) Get(key string, subKeys ...string) interface{} {
	o := self.settings[key]
	if len(subKeys) == 0 {
		return o
	}

	if o == nil {
		return nil
	}

	for _, subKey := range subKeys {
		m, ok := o.(map[string]interface{})
		if !ok {
			return nil
		}
		o = m[subKey]
		if o == nil {
			return nil
		}
	}
	return o
}

// GetAsString 读配置
func (self *Config) GetAsString(keys []string, defaultValue string) string {
	o := self.Get(keys[0], keys[1:]...)
	return as.StringWithDefault(o, defaultValue)
}

// GetAsInt 读配置
func (self *Config) GetAsInt(keys []string, defaultValue int) int {
	o := self.Get(keys[0], keys[1:]...)
	return as.IntWithDefault(o, defaultValue)
}

// GetAsBool 读配置
func (self *Config) GetAsBool(keys []string, defaultValue bool) bool {
	o := self.Get(keys[0], keys[1:]...)
	return as.BoolWithDefault(o, defaultValue)
}

// GetAsDuration 读配置
func (self *Config) GetAsDuration(keys []string, defaultValue time.Duration) time.Duration {
	o := self.Get(keys[0], keys[1:]...)
	return as.DurationWithDefault(o, defaultValue)
}

// GetAsTime 读配置
func (self *Config) GetAsTime(keys []string, defaultValue time.Time) time.Time {
	o := self.Get(keys[0], keys[1:]...)
	return as.TimeWithDefault(o, defaultValue)
}

// DurationWithDefault 读配置
func (self *Config) ForEach(cb func(key string, value interface{})) {
	for k, v := range self.settings {
		cb(k, v)
	}
}

func boolWith(cfg map[string]string, key string, defaultValue bool) bool {
	if v, ok := cfg[key]; ok && v != "" {
		switch strings.ToLower(v) {
		case "true", "on", "enabled":
			return true
		case "false", "off", "disabled":
			return false
		}
	}
	return defaultValue
}

func hostWith(cfg map[string]string, key, value string) string {
	v := stringWith(cfg, key, value)
	if ip := net.ParseIP(v); nil == ip {
		panic("'" + key + "' isn't a ip address - '" + v + "'.")
	}
	return v
}

func portWith(cfg map[string]string, key, value string) string {
	v := stringWith(cfg, key, value)
	if _, e := strconv.ParseInt(v, 10, 32); nil != e {
		panic("'" + key + "' isn't a port number - '" + v + "'.")
	}
	return v
}

func LoadConfigs(fs FileSystem, nm string, dumpFilesIfFound, dumpFilesIfNotFound bool) map[string]string {
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
				log.Println("read properties '" + file + "' failed," + e.Error())
			}
			return nil
		}

		if dumpFilesIfFound {
			log.Println("load properties '" + file + "'.")
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

func ReadConfigs(fs FileSystem, files []string, nm string, dumpFilesIfFound, dumpFilesIfNotFound bool) map[string]string {
	cfg := LoadConfigs(fs, "app.properties", dumpFilesIfFound, dumpFilesIfNotFound)
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

			if dumpFilesIfFound {
				log.Println("[info] load config ok: ", file)
			}
			if len(props) > 0 {
				for k, v := range props {
					cfg[k] = v
				}
			}
		}
	}

	localCfg := LoadConfigs(fs, nm+".properties", dumpFilesIfFound, dumpFilesIfNotFound)
	if nil != localCfg {
		for k, v := range localCfg {
			cfg[k] = v
		}
	}
	engineCfg := LoadConfigs(fs, "engine.properties", dumpFilesIfFound, dumpFilesIfNotFound)
	if nil != engineCfg {
		for k, v := range engineCfg {
			cfg[k] = v
		}
	}
	return cfg
}

var dbDefaults = map[string]string{}

func SetDefaults(defaults map[string]string) {
	dbDefaults = defaults
}

func InitConfig(flagSet *flag.FlagSet, cfg map[string]string) error {
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
			redis_address := stringWith(cfg, "redis.host", dbDefaults["redis.host"])
			redis_port := stringWith(cfg, "redis.port", dbDefaults["redis.port"]) // nolint
			flagSet.Set(k, redis_address+":"+redis_port)
		case "data_db.url", "db_url": // for tsdb
			drv, url, e := CreateDBUrl("data.", cfg, dbDefaults)
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
			drv, url, e := CreateDBUrl("models.", cfg, dbDefaults)
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
	return dbConfig.dbUrl()
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
