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
	"github.com/runner-mei/log"
	commons_cfg "github.com/three-plus-three/modules/cfg"
	"github.com/three-plus-three/modules/util"
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
	NotPrintIfFilesFound bool
	Args                 []string
	NotRedirectStdLog    bool
	IsTest               bool
}

// Environment
type Environment struct {
	HeaderTitleText string
	FooterTitleText string

	LoginHeaderTitleText, LoginFooterTitleText string

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
	Logger             log.Logger
	undoRedirectStdLog func()
	notRedirectStdLog  bool

	Engine EngineConfig
}

func (env *Environment) EnabledPipe() bool {
	return env.Config.BoolWithDefault("pipe_enabled", false)
}

func (env *Environment) SetCurrent(current ENV_PROXY_TYPE) *Environment {
	env.CurrentApplication = current
	if !IsValidProxyID(current) {
		return env
	}
	so := env.GetServiceConfig(env.CurrentApplication)
	if err := env.ensureLogger(so.Name); err != nil {
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

// func (self *Environment) RemoveAllListener() {
// 	for idx := range self.serviceOptions {
// 		self.serviceOptions[idx].RemoveAllListener()
// 	}
// }

func (self *Environment) GetMasterConfig() *ServiceConfig {
	if !self.Engine.IsMasterHost {
		if self.Engine.RemotePort != "" && self.Engine.RemotePort != "0" {
			return self.GetServiceConfig(ENV_HOME_SSL_PROXY_ID)
		}
	}

	if self.EnabledPipe() {
		return self.GetServiceConfig(ENV_GATEWAY_PROXY_ID)
	}
	return self.GetServiceConfig(ENV_HOME_PROXY_ID)
}

func (self *Environment) GetServiceConfig(id ENV_PROXY_TYPE) *ServiceConfig {
	for idx := range self.serviceOptions {
		if self.serviceOptions[idx].ID == id {
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
		var rootDir = opt.ConfDir
		if rootDir == "" {
			if s := os.Getenv("hw_root_dir"); s != "" {
				rootDir = s
			} else if cwd, e := os.Getwd(); nil == e && util.FileExists(filepath.Join(cwd, "conf", "app.properties")) {
				rootDir = cwd
			} else if nil == e && util.FileExists(filepath.Join(cwd, "..", "conf", "app.properties")) {
				rootDir = filepath.Clean(filepath.Join(cwd, ".."))
			} else if exeDir, e := osext.ExecutableFolder(); nil == e && util.FileExists(filepath.Join(exeDir, "conf", "app.properties")) {
				rootDir = exeDir
			} else if nil == e && util.FileExists(filepath.Join(exeDir, "..", "conf", "app.properties")) {
				rootDir = filepath.Clean(filepath.Join(exeDir, ".."))
			} else if opt.IsTest {
				rootDir, _ = os.Getwd()
			} else if rootDir == "" {
				found := false
				for _, s := range []string{"../../../../cn/com/hengwei",
					"../../../../../cn/com/hengwei",
					"../../../../../../cn/com/hengwei",
					"../../../../../../../cn/com/hengwei"} {
					abs, _ := filepath.Abs(s)
					abs = filepath.Clean(abs)
					if util.DirExists(abs) {
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
		} else if rootDir == "<default>" || rootDir == "." { // "<default>" 作为一个特殊的字符，自动使用当前目录
			if cwd, e := os.Getwd(); nil == e {
				rootDir = cwd
			} else {
				rootDir = "."
			}
		} else {
			rootDir = filepath.Join(opt.ConfDir, "..")
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
	projectName := opt.Name
	if projectName == "" {
		for _, so := range ServiceOptions {
			if so.ID == opt.CurrentApplication {
				projectName = so.Name
			}
		}
	}

	cfg := ReadConfigs(fs, opt.ConfigFiles, projectName, !opt.NotPrintIfFilesFound, opt.PrintIfFilesNotFound)
	if e := InitConfig(opt.FlagSet, cfg); nil != e {
		return nil, e
	}

	env := &Environment{
		notRedirectStdLog:  opt.NotRedirectStdLog,
		CurrentApplication: opt.CurrentApplication,
		//rootDir: opt.rootDir,
		Fs:     fs,
		Name:   projectName,
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
	env.Engine = loadEngineConfig(&env.Config)
	env.serviceOptions = make([]ServiceConfig, len(ServiceOptions))
	for idx, so := range ServiceOptions {
		env.serviceOptions[idx].loadConfig(cfg, so)
	}
	for idx := range env.serviceOptions {
		env.serviceOptions[idx].env = env
		// env.serviceOptions[idx].listeners.Init()
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


	env.LoginHeaderTitleText = ReadFileWithDefault([]string{
			fs.FromDataConfig("resources/profiles/login-title.txt"),
			fs.FromData("resources/profiles/login-title.txt")},
			env.HeaderTitleText)

	env.LoginFooterTitleText = ReadFileWithDefault([]string{
			fs.FromDataConfig("resources/profiles/login-footer.txt"),
			fs.FromData("resources/profiles/login-footer.txt")},
			env.FooterTitleText)

	if err := agent.Listen(agent.Options{}); err != nil {
		env.Logger.Warn("启动调试代理失败", log.Error(err))
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
