package environment

import (
	"strings"
)

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

func loadEngineConfig(cfg *Config) EngineConfig {
	engine := EngineConfig{IsEnabled: cfg.BoolWithDefault("engine.is_enabled", false),
		Name:       strings.TrimSpace(cfg.StringWithDefault("engine.name", "default")),
		IsSSL:      cfg.BoolWithDefault("engine.is_ssl", false),
		RemoteHost: strings.TrimSpace(cfg.StringWithDefault("engine.remote_host", "127.0.0.1")),
		RemotePort: strings.TrimSpace(cfg.StringWithDefault("engine.remote_port", ""))}

	engine.IsMasterHost = engine.isMaster()
	return engine
}
