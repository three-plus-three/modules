package environment

import (
	"github.com/three-plus-three/modules/errors"
	"go.uber.org/zap"
)

func (env *Environment) initLogger(name string) error {
	var err error
	env.LogConfig = zap.NewProductionConfig()
	env.Logger, err = env.LogConfig.Build()
	if err != nil {
		return errors.Wrap(err, "init zap logger fail")
	}
	env.Logger = env.Logger.Named(name)

	zap.ReplaceGlobals(env.Logger)
	env.SugaredLogger = env.Logger.Sugar()
	env.undoRedirectStdLog = zap.RedirectStdLog(env.Logger)
	return nil
}
