package environment

import (
	"github.com/runner-mei/log"
	"github.com/three-plus-three/modules/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (env *Environment) initLogger(name string) error {
	var err error
	env.LogConfig = zap.NewProductionConfig()
	env.LogConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := env.LogConfig.Build()
	if err != nil {
		return errors.Wrap(err, "init zap logger fail")
	}
	if name != "" {
		logger = logger.Named(name)
	}

	zap.ReplaceGlobals(logger)
	if !env.notRedirectStdLog {
		env.undoRedirectStdLog = zap.RedirectStdLog(logger)
	}

	env.Logger = log.NewLogger(logger)
	return nil
}

func (env *Environment) ensureLogger(name string) error {
	if env.Logger == nil {
		return env.initLogger(name)
	}

	env.Logger = env.Logger.Named(name)
	zap.ReplaceGlobals(env.Logger.Unwrap())

	if env.undoRedirectStdLog != nil {
		env.undoRedirectStdLog()
	}

	if !env.notRedirectStdLog {
		env.undoRedirectStdLog = zap.RedirectStdLog(env.Logger.Unwrap())
	}
	return nil
}
