package environment

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kardianos/osext"
	"github.com/three-plus-three/modules/as"
	commons_cfg "github.com/three-plus-three/modules/cfg"
)

type ENV_PROXY_TYPE int

const (
	ENV_REDIS_PROXY_ID ENV_PROXY_TYPE = iota
	ENV_MODELS_PROXY_ID
	ENV_SAMPLING_PROXY_ID
	ENV_SAMPLING_STUB_PROXY_ID
	ENV_POLL_PROXY_ID
	ENV_TSDB_PROXY_ID
	ENV_SCHD_PROXY_ID
	ENV_LCN_PROXY_ID
	ENV_IP_MGR_PROXY_ID
	ENV_DELAYED_JOB_PROXY_ID
	ENV_TERMINAL_PROXY_ID
	ENV_JBRIDGE_PROXY_ID
	ENV_JBRIDGE15_PROXY_ID
	ENV_ES_PROXY_ID
	ENV_REST_PROXY_ID
	ENV_WSERVER_PROXY_ID
	ENV_LUA_BRIDGE_PROXY_ID
	ENV_WEB_PROXY_ID
	ENV_LOGGING_PROXY_ID
	ENV_NFLOW_PROXY_ID
	ENV_MC_PROXY_ID
	ENV_MC_DEV_PROXY_ID
	ENV_INFLUXDB_PROXY_ID
	ENV_INFLUXDB_ADM_PROXY_ID
	ENV_FORK_PROXY_ID
	ENV_IMS_PROXY_ID
	ENV_WSERVER_SSL_PROXY_ID
	ENV_CMDB_PROXY_ID
	ENV_ASSET_MANAGE_PROXY_ID
	ENV_NSM_PROXY_ID
	ENV_MINIO_PROXY_ID
	ENV_MAX_PROXY_ID

	ENV_MIN_PROXY_ID = ENV_REDIS_PROXY_ID
	ENV_DS_PROXY_ID  = ENV_MODELS_PROXY_ID
)

type ServiceOption struct {
	Id   ENV_PROXY_TYPE
	Name string
	Host string
	Port string
	Path string
}

var (
	ServiceOptions = []ServiceOption{
		{Id: ENV_REDIS_PROXY_ID, Name: "redis", Host: "127.0.0.1", Port: "36379"},
		{Id: ENV_MODELS_PROXY_ID, Name: "ds", Host: "127.0.0.1", Port: "37071"},
		{Id: ENV_SAMPLING_PROXY_ID, Name: "sampling", Host: "127.0.0.1", Port: "37072"},
		{Id: ENV_SAMPLING_STUB_PROXY_ID, Name: "sampling_stub", Host: "127.0.0.1", Port: "37081"},
		{Id: ENV_POLL_PROXY_ID, Name: "poll", Host: "127.0.0.1", Port: "37073"},
		{Id: ENV_TSDB_PROXY_ID, Name: "tsdb", Host: "127.0.0.1", Port: "37074"},
		{Id: ENV_SCHD_PROXY_ID, Name: "schd", Host: "127.0.0.1", Port: "37075"},
		{Id: ENV_LCN_PROXY_ID, Name: "lcn", Host: "127.0.0.1", Port: "37076"},
		{Id: ENV_IP_MGR_PROXY_ID, Name: "ip_mgr", Host: "127.0.0.1", Port: "37077"},
		{Id: ENV_DELAYED_JOB_PROXY_ID, Name: "delayed_jobs", Host: "127.0.0.1", Port: "37078"},
		{Id: ENV_TERMINAL_PROXY_ID, Name: "terminal", Host: "127.0.0.1", Port: "37079"},
		{Id: ENV_JBRIDGE_PROXY_ID, Name: "jbridge", Host: "127.0.0.1", Port: "37080"},
		{Id: ENV_ES_PROXY_ID, Name: "es_old", Host: "127.0.0.1", Port: "39300"},
		{Id: ENV_REST_PROXY_ID, Name: "rest", Host: "127.0.0.1", Port: "39301"},
		{Id: ENV_WSERVER_PROXY_ID, Name: "wserver", Host: "127.0.0.1", Port: "37070"},
		{Id: ENV_WSERVER_SSL_PROXY_ID, Name: "daemon_ssl", Host: "127.0.0.1", Port: "37090"},
		{Id: ENV_LUA_BRIDGE_PROXY_ID, Name: "lua_bridge", Host: "127.0.0.1", Port: "37082"},
		{Id: ENV_WEB_PROXY_ID, Name: "web", Host: "127.0.0.1", Port: "39000"},
		{Id: ENV_LOGGING_PROXY_ID, Name: "es", Host: "127.0.0.1", Port: "37083"},
		{Id: ENV_NFLOW_PROXY_ID, Name: "nflow", Host: "127.0.0.1", Port: "37084"},
		{Id: ENV_MC_PROXY_ID, Name: "mc", Host: "127.0.0.1", Port: "37085"},
		{Id: ENV_MC_DEV_PROXY_ID, Name: "mc_dev", Host: "127.0.0.1", Port: "9000"},
		{Id: ENV_INFLUXDB_PROXY_ID, Name: "influxdb", Host: "127.0.0.1", Port: "37086"},
		{Id: ENV_INFLUXDB_ADM_PROXY_ID, Name: "influxdb_adm", Host: "127.0.0.1", Port: "39183"},
		{Id: ENV_FORK_PROXY_ID, Name: "fork", Host: "127.0.0.1", Port: "37087"},
		{Id: ENV_JBRIDGE15_PROXY_ID, Name: "jbridge15", Host: "127.0.0.1", Port: "37088"},
		{Id: ENV_IMS_PROXY_ID, Name: "ims", Host: "127.0.0.1", Port: "37089"},
		{Id: ENV_CMDB_PROXY_ID, Name: "cmdb", Host: "127.0.0.1", Port: "37091"},
		{Id: ENV_ASSET_MANAGE_PROXY_ID, Name: "am", Host: "127.0.0.1", Port: "37092"},
		{Id: ENV_NSM_PROXY_ID, Name: "nsm", Host: "127.0.0.1", Port: "37093"},
		{Id: ENV_MINIO_PROXY_ID, Name: "minio", Host: "127.0.0.1", Port: "37094"},
	}
)

type Options struct {
	ConfDir              string
	FlagSet              *flag.FlagSet
	Name                 string
	PrintIfFilesNotFound bool
	Args                 []string
	IsTest               bool
}

type EngineConfig struct {
	IsEnabled       bool
	IsMasterHost    bool
	Name            string
	IsRemoteBlocked bool
	RemoteHost      string
	RemotePort      string
}

func (self EngineConfig) IsMaster() bool {
	return strings.ToLower(strings.TrimSpace(self.Name)) == "default"
}

type DbConfig struct {
	DbType   string
	Address  string
	Port     string
	Schema   string
	Username string
	Password string
}

func (db *DbConfig) Host() string {
	if "" != db.Port && "0" != db.Port {
		return db.Address + ":" + db.Port
	}
	switch db.DbType {
	case "postgresql":
		return db.Address + ":35432"
	default:
		panic(errors.New("unknown db type - " + db.DbType))
	}
}

func (db *DbConfig) Url() (string, string) {
	switch db.DbType {
	case "postgresql":
		if db.Port == "" {
			db.Port = "5432"
		}
		return "postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			db.Address, db.Port, db.Schema, db.Username, db.Password)
	case "mysql":
		if db.Port == "" {
			db.Port = "3306"
		}
		return "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			db.Username, db.Password, db.Address, db.Port, db.Schema)
	default:
		panic(errors.New("unknown db type - " + db.DbType))
	}
}

type Environment struct {
	Fs FileSystem

	Name   string
	Config Config

	Db struct {
		Models DbConfig
		Data   DbConfig
	}

	RawDaemonUrlPath string
	DaemonUrlPath    string
	serviceOptions   []ServiceConfig

	Engine EngineConfig
}

func (self *Environment) Clone() *Environment {
	var copyed = &Environment{}
	*copyed = *self
	copyed.serviceOptions = make([]ServiceConfig, len(self.serviceOptions))
	for idx := range self.serviceOptions {
		copyed.serviceOptions[idx].copyFrom(&self.serviceOptions[idx])
	}

	for idx := range self.serviceOptions {
		copyed.serviceOptions[idx].env = copyed
	}
	return copyed
}

func (self *Environment) RemoveAllListener() {
	for idx := range self.serviceOptions {
		self.serviceOptions[idx].RemoveAllListener()
	}
}

func (self *Environment) GetServiceConfig(id ENV_PROXY_TYPE) *ServiceConfig {
	for idx := range self.serviceOptions {
		if self.serviceOptions[idx].Id == id {
			return &self.serviceOptions[idx]
		}
	}
	return UnknownServiceConfig
}

func (self *Environment) GetEngineConfig() *EngineConfig {
	return &self.Engine
}

func NewEnvironment(opt Options) (*Environment, error) {
	var fs FileSystem
	if runtime.GOOS == "windows" {
		var rootDir string
		if "" == opt.ConfDir {
			if cwd, e := os.Getwd(); nil == e && FileExists(filepath.Join(cwd, "conf", "app.properties")) {
				rootDir = cwd
			} else if nil == e && FileExists(filepath.Join(cwd, "..", "conf", "app.properties")) {
				rootDir = filepath.Clean(filepath.Join(cwd, ".."))
			} else if exeDir, e := osext.ExecutableFolder(); nil == e && FileExists(filepath.Join(exeDir, "conf", "app.properties")) {
				rootDir = exeDir
			} else if nil == e && FileExists(filepath.Join(exeDir, "..", "conf", "app.properties")) {
				rootDir = filepath.Clean(filepath.Join(exeDir, ".."))
			} else if opt.IsTest {
				rootDir, _ = os.Getwd()
			} else {
				found := false
				for _, s := range []string{"../../../../cn/com/hengwei",
					"../../../../../cn/com/hengwei",
					"../../../../../../cn/com/hengwei",
					"../../../../../../../cn/com/hengwei"} {
					abs, _ := filepath.Abs(s)
					abs = filepath.Clean(abs)
					if DirExists(abs) {
						rootDir = abs
						found = true
						break
					}
				}
				if !found {
					//if "<default>" == opt.rootDir { // "<default>" 作为一个特殊的字符，自动使用当前目录
					if cwd, e := os.Getwd(); nil == e {
						rootDir = cwd
					} else {
						rootDir = "."
					}
					// } else {
					// 	return nil, errors.New("root directory is not found")
					// }
				}
				opt.IsTest = true
			}
		} else {
			if "<default>" == opt.ConfDir { // "<default>" 作为一个特殊的字符，自动使用当前目录
				if cwd, e := os.Getwd(); nil == e {
					rootDir = cwd
				} else {
					rootDir = "."
				}
			} else {
				rootDir = filepath.Join(opt.ConfDir, "..")
			}
		}
		fs = &winFs{rootDir: rootDir}
	} else {
		fs = &linuxFs{
			installDir: "/usr/local/tpt",
			binDir:     "/usr/local/tpt/bin",
			logDir:     "/var/log/tpt",
			dataDir:    "/var/lib/tpt",
			confDir:    "/etc/tpt",
			tmpDir:     "/tmp/tpt",
			runDir:     "/var/run/tpt",
		}
	}

	cfg := ReadConfigs(fs, opt.Name, opt.PrintIfFilesNotFound)
	if !opt.IsTest {
		if e := InitConfig(opt.FlagSet, cfg, opt.PrintIfFilesNotFound); nil != e {
			return nil, e
		}
	}

	env := &Environment{
		//rootDir: opt.rootDir,
		Fs:     fs,
		Name:   opt.Name,
		Config: Config{settings: map[string]interface{}{}},
	}
	for k, v := range cfg {
		env.Config.Set(k, v)
	}
	env.Db.Models = ReadDbConfig("models.", cfg, db_defaults)
	env.Db.Data = ReadDbConfig("data.", cfg, db_defaults)

	if opt.IsTest {
		env.Db.Models.Port = "5432"
		env.Db.Models.Schema = "tpt_models_test"

		env.Db.Data.Port = "5432"
		env.Db.Data.Schema = "tpt_data_test"
	}

	env.serviceOptions = make([]ServiceConfig, len(ServiceOptions))
	for idx, so := range ServiceOptions {
		loadServiceConfig(cfg, so, &env.serviceOptions[idx])
	}
	for idx := range env.serviceOptions {
		env.serviceOptions[idx].env = env
		env.serviceOptions[idx].listeners.Init()
	}

	var tsdbConfigFile string
	var tsdbConfig map[string]interface{}
	var err error

	if runtime.GOOS == "windows" {
		tsdbConfigFile = env.Fs.FromConfig("tsdb_config.win.conf")
	} else {
		tsdbConfigFile = env.Fs.FromConfig("tsdb_config.conf")
	}

	_, err = toml.DecodeFile(tsdbConfigFile, &tsdbConfig)
	if err != nil {
		if opt.PrintIfFilesNotFound {
			log.Println("[warn] load tsdb config fail from", tsdbConfigFile)
		}
	} else if nil != tsdbConfig {
		tsdbHTTP, _ := as.Object(tsdbConfig["http"])
		if nil != tsdbHTTP {
			if _, port, err := net.SplitHostPort(fmt.Sprint(tsdbHTTP["bind-address"])); err == nil {
				env.GetServiceConfig(ENV_INFLUXDB_PROXY_ID).SetPort(port)
				if opt.PrintIfFilesNotFound {
					log.Println("[warn] load tsdb http port ("+port+") from", tsdbConfigFile)
				}
			}
		}

		tsdbAdmin, _ := as.Object(tsdbConfig["admin"])
		if nil != tsdbAdmin {
			if _, port, err := net.SplitHostPort(fmt.Sprint(tsdbAdmin["bind-address"])); err == nil {
				env.GetServiceConfig(ENV_INFLUXDB_ADM_PROXY_ID).SetPort(port)
				if opt.PrintIfFilesNotFound {
					log.Println("[warn] load tsdb admin port ("+port+") from", tsdbConfigFile)
				}
			}
		}
	}

	if minioConfig := loadMinioConfig(env.Fs); minioConfig != nil {
		env.Config.Set("minio_config", minioConfig)
	}
	for _, nm := range []string{env.Fs.FromWebConfig("application.conf"),
		env.Fs.FromDataConfig("web/application.conf"),
		env.Fs.FromInstallRoot("web/conf/application.conf")} {
		if props, _ := commons_cfg.ReadProperties(nm); nil != props {
			if secret := props["application.secret"]; "" != secret {
				secret = strings.TrimPrefix(strings.TrimSuffix(secret, "\""), "\"")
				env.Config.Set("app.secret", secret)
				if opt.PrintIfFilesNotFound {
					log.Println("[warn] load app.secret from", nm, secret)
				}
			}
		}
	}

	env.RawDaemonUrlPath = stringWith(cfg, "daemon.urlpath", "hengwei")
	env.DaemonUrlPath = env.RawDaemonUrlPath
	if !strings.HasPrefix(env.DaemonUrlPath, "/") {
		env.DaemonUrlPath = "/" + env.DaemonUrlPath
	}
	if !strings.HasSuffix(env.DaemonUrlPath, "/") {
		env.DaemonUrlPath = env.DaemonUrlPath + "/"
	}

	env.Engine = loadEngineRegistry(&env.Config)
	return env, nil
}

func loadServiceConfig(cfg map[string]string, so ServiceOption, sc *ServiceConfig) *ServiceConfig {
	sc.Id = so.Id
	sc.Name = so.Name

	if so.Id == ENV_WSERVER_PROXY_ID {
		sc.Host = hostWith(cfg, so.Name+".host", stringWith(cfg, "daemon.host", so.Host))
		sc.Port = portWith(cfg, so.Name+".port", stringWith(cfg, "daemon.port", so.Port))
	} else {
		sc.Host = hostWith(cfg, so.Name+".host", so.Host)
		sc.Port = portWith(cfg, so.Name+".port", so.Port)
	}

	if ENV_MC_DEV_PROXY_ID == so.Id {
		if mcDevPort := os.Getenv("mc_dev_port"); "" != mcDevPort {
			sc.Port = mcDevPort
		}
	}

	return sc
}

func loadEngineRegistry(cfg *Config) EngineConfig {
	engine := EngineConfig{IsEnabled: cfg.BoolWithDefault("engine.is_enabled", false),
		Name:            strings.TrimSpace(cfg.StringWithDefault("engine.name", "default")),
		IsRemoteBlocked: cfg.BoolWithDefault("engine.remote_blocked", false),
		RemoteHost:      strings.TrimSpace(cfg.StringWithDefault("engine.remote_host", "127.0.0.1")),
		RemotePort:      strings.TrimSpace(cfg.StringWithDefault("engine.remote_port", ""))}

	engine.IsMasterHost = engine.IsMaster()
	return engine
}

func boolWith(cfg map[string]string, key string, value bool) bool {
	v, ok := cfg[key]
	if !ok {
		return value
	}
	if "yes" == strings.ToLower(v) || "true" == strings.ToLower(v) {
		return true
	}
	return false
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

type Config struct {
	settings map[string]interface{}
}

func (self *Config) PasswordWithDefault(key, defValue string) string {
	if s, ok := self.settings[key]; ok {
		return as.StringWithDefault(s, defValue)
	}
	return defValue
}

func (self *Config) StringWithDefault(key, defValue string) string {
	if s, ok := self.settings[key]; ok {
		return as.StringWithDefault(s, defValue)
	}
	return defValue
}

func (self *Config) IntWithDefault(key string, defValue int) int {
	if s, ok := self.settings[key]; ok {
		return as.IntWithDefault(s, defValue)
	}
	return defValue
}

func (self *Config) BoolWithDefault(key string, defValue bool) bool {
	if s, ok := self.settings[key]; ok {
		return as.BoolWithDefault(s, defValue)
	}
	return defValue
}

func (self *Config) DurationWithDefault(key string, defValue time.Duration) time.Duration {
	if s, ok := self.settings[key]; ok {
		return as.DurationWithDefault(s, defValue)
	}
	return defValue
}

func (self *Config) Set(key string, value interface{}) {
	self.settings[key] = value
}

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

func (self *Config) GetAsString(keys []string, defaultValue string) string {
	o := self.Get(keys[0], keys[1:]...)
	return as.StringWithDefault(o, defaultValue)
}

func (self *Config) GetAsInt(keys []string, defaultValue int) int {
	o := self.Get(keys[0], keys[1:]...)
	return as.IntWithDefault(o, defaultValue)
}

func (self *Config) GetAsBool(keys []string, defaultValue bool) bool {
	o := self.Get(keys[0], keys[1:]...)
	return as.BoolWithDefault(o, defaultValue)
}

func (self *Config) GetAsDuration(keys []string, defaultValue time.Duration) time.Duration {
	o := self.Get(keys[0], keys[1:]...)
	return as.DurationWithDefault(o, defaultValue)
}

func (self *Config) GetAsTime(keys []string, defaultValue time.Time) time.Time {
	o := self.Get(keys[0], keys[1:]...)
	return as.TimeWithDefault(o, defaultValue)
}

func (self *Config) ForEach(cb func(key string, value interface{})) {
	for k, v := range self.settings {
		cb(k, v)
	}
}

func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func DirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}
