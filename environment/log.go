package environment

import "go.uber.org/zap"

func (env *Environment) initLogger() {
	var err error
	env.LogConfig = zap.NewProductionConfig()
	env.Logger, err = env.LogConfig.Build()
	if err != nil {
		panic(err)
	}
	// env.Logger = env.Logger.Named(name)

	zap.ReplaceGlobals(env.Logger)
	env.SugaredLogger = env.Logger.Sugar()
	env.undoRedirectStdLog = zap.RedirectStdLog(env.Logger)
}
