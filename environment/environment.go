package environment

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/gops/agent"
	"github.com/kardianos/osext"
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
	NotRedirectStdLog    bool
	IsTest               bool
}

// EngineConfig 多引擎时的配置
type EngineConfig struct {
	IsEnabled    bool
	IsMasterHost bool
	Name         string
	IsSSL        bool
	RemoteHost   string
	RemotePort   string
}

func (self EngineConfig) isMaster() bool {
	return strings.ToLower(strings.TrimSpace(self.Name)) == "default"
}

// Environment
type Environment struct {
	HeaderTitleText string
	FooterTitleText string

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
	notRedirectStdLog  bool

	Engine EngineConfig
}

func (env *Environment) SetCurrent(current ENV_PROXY_TYPE) *Environment {
	env.CurrentApplication = current
	if !IsValidProxyID(current) {
		return env
	}
	so := env.GetServiceConfig(env.CurrentApplication)
	if err := env.reinitLogger(so.Name); err != nil {
		panic(err)
	}
	return env
}

func (env *Environment) Current() *ServiceConfig {
	if !IsValidProxyID(env.CurrentApplication) {
		panic(errors.New("currnet application is no value"))
	}
	return env.GetServiceConfig(env.CurrentApplication)
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

func (self *Environment) GetMasterConfig() *ServiceConfig {
	if engineCfg := self.GetEngineConfig(); engineCfg.IsEnabled && !engineCfg.IsMasterHost {
		if engineCfg.RemotePort != "" && engineCfg.RemotePort != "0" {
			return self.GetServiceConfig(ENV_WSERVER_SSL_PROXY_ID)
		}
	}

	return self.GetServiceConfig(ENV_WSERVER_PROXY_ID)
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
			} else if rootDir = os.Getenv("hw_root_dir"); rootDir == "" {
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
		notRedirectStdLog:  opt.NotRedirectStdLog,
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

	env.Db.Models = ReadDbConfig("models.", cfg, dbDefaults)
	env.Db.Data = ReadDbConfig("data.", cfg, dbDefaults)
	env.Engine = loadEngineRegistry(&env.Config)
	env.serviceOptions = make([]ServiceConfig, len(ServiceOptions))
	for idx, so := range ServiceOptions {
		loadServiceConfig(cfg, so, &env.serviceOptions[idx])
	}
	for idx := range env.serviceOptions {
		env.serviceOptions[idx].env = env
		env.serviceOptions[idx].listeners.Init()
	}

	if env.CurrentApplication != ENV_MIN_PROXY_ID {
		so := env.GetServiceConfig(env.CurrentApplication)
		if err := env.initLogger(so.Name); err != nil {
			return nil, err
		}
	} else {
		if err := env.initLogger(""); err != nil {
			return nil, err
		}
	}
	if err := env.initTSDB(opt.PrintIfFilesNotFound); err != nil {
		return nil, err
	}

	if minioConfig := loadMinioConfig(env.Fs); minioConfig != nil {
		env.Config.Set("minio_config", minioConfig)
	}

	exists := false
	filenames := []string{env.Fs.FromWebConfig("application.conf"),
		env.Fs.FromDataConfig("web/application.conf"),
		env.Fs.FromInstallRoot("web/conf/application.conf")}
	for _, nm := range filenames {
		if props, _ := commons_cfg.ReadProperties(nm); nil != props {
			if secret := props["application.secret"]; "" != secret && "\"\"" != secret {
				secret = strings.TrimPrefix(strings.TrimSuffix(secret, "\""), "\"")
				env.Config.Set("app.secret", secret)
				exists = true
				break
			}
		}
	}

	if !exists && opt.PrintIfFilesNotFound {
		env.Logger.Warn("no load app.secret from '" + strings.Join(filenames, ",") + "'")
	}

	env.HeaderTitleText = "IT综合运维管理平台"
	if env.Config.BoolWithDefault("is_nflow", false) {
		env.HeaderTitleText = "数据流分析 v1.0"
	}

	env.HeaderTitleText = ReadFileWithDefault([]string{
		env.Fs.FromDataConfig("resources/profiles/header.txt"),
		env.Fs.FromData("resources/profiles/header.txt"),
		filepath.Join(os.Getenv("hw_root_dir"), "data/resources/profiles/header.txt")},
		env.HeaderTitleText)

	env.FooterTitleText = ReadFileWithDefault([]string{
		env.Fs.FromDataConfig("resources/profiles/footer.txt"),
		env.Fs.FromData("resources/profiles/footer.txt"),
		filepath.Join(os.Getenv("hw_root_dir"), "data/resources/profiles/footer.txt")},
		"© 2019 恒维信息技术(上海)有限公司, 保留所有版权。")

	if err := agent.Listen(agent.Options{}); err != nil {
		env.Logger.Warn("启动调试代理失败", zap.Error(err))
	}

	return env, callHooks(env)
}

func ReadFileWithDefault(files []string, defaultValue string) string {
	for _, s := range files {
		content, e := ioutil.ReadFile(s)
		if nil == e {
			if content = bytes.TrimSpace(content); len(content) > 0 {
				return string(content)
			}
		}
	}
	return defaultValue
}

func loadServiceConfig(cfg map[string]string, so ServiceOption, sc *ServiceConfig) *ServiceConfig {
	sc.Id = so.ID
	sc.Name = so.Name

	sc.IsSSL = boolWith(cfg, so.Name+".is_ssl", so.IsSSL)
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
		Name:       strings.TrimSpace(cfg.StringWithDefault("engine.name", "default")),
		IsSSL:      cfg.BoolWithDefault("engine.is_ssl", false),
		RemoteHost: strings.TrimSpace(cfg.StringWithDefault("engine.remote_host", "127.0.0.1")),
		RemotePort: strings.TrimSpace(cfg.StringWithDefault("engine.remote_port", ""))}

	engine.IsMasterHost = engine.isMaster()
	return engine
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
