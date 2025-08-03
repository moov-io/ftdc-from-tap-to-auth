package onemorething

import (
	"log/slog"
)

type app struct {
	logger        *slog.Logger
	iso8583Server *server
	config        *Config
}

func NewApp(logger *slog.Logger, config *Config) *app {
	return &app{
		logger: logger,
		config: config,
	}
}

func (a *app) Start() {
	a.logger.Info("Starting One More Thing ...")

	a.iso8583Server = NewServer(a.logger, a.config.ServerAddr)
	err := a.iso8583Server.Start()
	if err != nil {
		a.logger.Error(
			"Error starting ISO 8583 server",
			slog.String("error", err.Error()),
		)
	}
}

func (a *app) Stop() {
	a.logger.Info("Stopping One More Thing ...")

	err := a.iso8583Server.Stop()
	if err != nil {
		a.logger.Error(
			"Error stopping ISO 8583 server",
			slog.String("error", err.Error()),
		)
	}

	a.logger.Info("One More Thing stopped")
}
