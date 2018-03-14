package environment

import (
	"github.com/three-plus-three/modules/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (env *Environment) initLogger(name string) error {
	var err error
	env.LogConfig = zap.NewProductionConfig()
	env.LogConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	env.Logger, err = env.LogConfig.Build()
	if err != nil {
		return errors.Wrap(err, "init zap logger fail")
	}
	if name != "" {
		env.Logger = env.Logger.Named(name)
	}

	zap.ReplaceGlobals(env.Logger)
	env.SugaredLogger = env.Logger.Sugar()
	if !env.notRedirectStdLog {
		env.undoRedirectStdLog = zap.RedirectStdLog(env.Logger)
	}
	return nil
}

func (env *Environment) reinitLogger(name string) error {
	if env.Logger == nil {
		return env.initLogger(name)
	}

	env.Logger = env.Logger.Named(name)
	zap.ReplaceGlobals(env.Logger)
	env.SugaredLogger = env.Logger.Sugar()

	if env.undoRedirectStdLog != nil {
		env.undoRedirectStdLog()
	}

	if !env.notRedirectStdLog {
		env.undoRedirectStdLog = zap.RedirectStdLog(env.Logger)
	}
	return nil
}
