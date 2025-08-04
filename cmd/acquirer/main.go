package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer"
	"github.com/moov-io/ftdc-from-tap-to-auth/internal/config"
	"github.com/moov-io/ftdc-from-tap-to-auth/log"
)

func main() {
	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		}),
	)

	cfg := &acquirer.Config{}

	err := config.NewFromFile("configs/acquirer.yaml", cfg)
	if err != nil {
		log.New().Error("Error loading config", "err", err)
		os.Exit(1)
	}

	app := acquirer.NewApp(logger, cfg)

	err = app.Start()
	if err != nil {
		logger.Error("Error starting app", "err", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	app.Shutdown()
}
