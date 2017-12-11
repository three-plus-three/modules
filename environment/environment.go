package environment

import (
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/three-plus-three/modules/as"
	commons_cfg "github.com/three-plus-three/modules/cfg"
	"go.uber.org/zap"
)

// Options 初始选项
type Options struct {
	CurrentApplication   ENV_PROXY_TYPE
	ConfigFiles          []string
	ConfDir              string
	FlagSet              *flag.FlagSet
	Name                 string
	PrintIfFilesNotFound bool
	Args                 []string
	IsTest               bool
}

// EngineConfig 多引擎时的配置
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

// Environment
type Environment struct {
	Fs FileSystem

	Name   string
	Config Config

	Db struct {
		Models DbConfig
		Data   DbConfig
	}

	CurrentApplication ENV_PROXY_TYPE
	RawDaemonUrlPath   string
	DaemonUrlPath      string
	serviceOptions     []ServiceConfig

	LogConfig          zap.Config
	Logger             *zap.Logger
	SugaredLogger      *zap.SugaredLogger
	undoRedirectStdLog func()

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
	var fs *linuxFs
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

		fs = &linuxFs{
			installDir: rootDir,
			binDir:     filepath.Join(rootDir, "bin"),
			logDir:     filepath.Join(rootDir, "logs"),
			dataDir:    filepath.Join(rootDir, "data"),
			confDir:    filepath.Join(rootDir, "data", "conf"),
			tmpDir:     filepath.Join(rootDir, "data", "tmp"),
			runDir:     rootDir,
		}
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

	if confDir := os.Getenv("hw_conf_dir"); confDir != "" {
		fs.confDir = confDir
	}
	if dataDir := os.Getenv("hw_data_dir"); dataDir != "" {
		fs.dataDir = dataDir
	}
	return NewEnvironmentWithFS(fs, opt)
}

func NewEnvironmentWithFS(fs FileSystem, opt Options) (*Environment, error) {
	cfg := ReadConfigs(fs, opt.ConfigFiles, opt.Name, opt.PrintIfFilesNotFound)
	if e := InitConfig(opt.FlagSet, cfg, opt.PrintIfFilesNotFound); nil != e {
		return nil, e
	}

	env := &Environment{
		CurrentApplication: opt.CurrentApplication,
		//rootDir: opt.rootDir,
		Fs:     fs,
		Name:   opt.Name,
		Config: Config{settings: map[string]interface{}{}},
	}
	for k, v := range cfg {
		env.Config.Set(k, v)
	}

	env.RawDaemonUrlPath = stringWith(cfg, "daemon.urlpath", "hengwei")
	env.DaemonUrlPath = env.RawDaemonUrlPath
	if !strings.HasPrefix(env.DaemonUrlPath, "/") {
		env.DaemonUrlPath = "/" + env.DaemonUrlPath
	}
	if !strings.HasSuffix(env.DaemonUrlPath, "/") {
		env.DaemonUrlPath = env.DaemonUrlPath + "/"
	}

	env.Db.Models = ReadDbConfig("models.", cfg, db_defaults)
	env.Db.Data = ReadDbConfig("data.", cfg, db_defaults)
	env.Engine = loadEngineRegistry(&env.Config)
	env.serviceOptions = make([]ServiceConfig, len(ServiceOptions))
	for idx, so := range ServiceOptions {
		loadServiceConfig(cfg, so, &env.serviceOptions[idx])
	}
	for idx := range env.serviceOptions {
		env.serviceOptions[idx].env = env
		env.serviceOptions[idx].listeners.Init()
	}

	so := env.GetServiceConfig(env.CurrentApplication)
	if err := env.initLogger(so.Name); err != nil {
		return nil, err
	}
	if err := env.initTSDB(opt.PrintIfFilesNotFound); err != nil {
		return nil, err
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

	return env, nil
}

func loadServiceConfig(cfg map[string]string, so ServiceOption, sc *ServiceConfig) *ServiceConfig {
	sc.Id = so.ID
	sc.Name = so.Name

	if so.ID == ENV_WSERVER_PROXY_ID {
		sc.Host = hostWith(cfg, so.Name+".host", stringWith(cfg, "daemon.host", so.Host))
		sc.Port = portWith(cfg, so.Name+".port", stringWith(cfg, "daemon.port", so.Port))
	} else {
		sc.Host = hostWith(cfg, so.Name+".host", so.Host)
		sc.Port = portWith(cfg, so.Name+".port", so.Port)
	}

	if ENV_MC_DEV_PROXY_ID == so.ID {
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

// FileExists 文件是否存在
func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// DirExists 目录是否存在
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
