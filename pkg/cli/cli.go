package cli

import (
	"github.com/applike/gosoline/pkg/cfg"
	"github.com/applike/gosoline/pkg/kernel"
	"time"
)

type kernelSettings struct {
	KillTimeout time.Duration `cfg:"killTimeout" default:"10s"`
}

func Run(module kernel.ModuleFactory) {
	configOptions := []cfg.Option{
		cfg.WithErrorHandlers(defaultErrorHandler),
		cfg.WithConfigFile("./config.dist.yml", "yml"),
		cfg.WithConfigFileFlag("config"),
	}

	config := cfg.New()
	if err := config.Option(configOptions...); err != nil {
		defaultErrorHandler(err, "can not initialize the config")
	}

	logger, err := newCliLogger()
	if err != nil {
		defaultErrorHandler(err, "can not initialize the logger")
	}

	settings := &kernelSettings{}
	config.UnmarshalKey("kernel", settings)

	k := kernel.New(config, logger, kernel.KillTimeout(settings.KillTimeout))
	k.Add("cli", module, kernel.ModuleType(kernel.TypeEssential), kernel.ModuleStage(kernel.StageApplication))
	k.Run()
}
