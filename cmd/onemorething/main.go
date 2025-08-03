package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/moov-io/ftdc-from-tap-to-auth/internal/config"
	"github.com/moov-io/ftdc-from-tap-to-auth/log"
	"github.com/moov-io/ftdc-from-tap-to-auth/onemorething"
)

func main() {
	logger := log.New()

	err := run(logger)
	if err != nil {
		logger.Error("Error running one more thing",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := &onemorething.Config{}

	err := config.NewFromFile("configs/one.yaml", cfg)
	if err != nil {

	}
	app := onemorething.NewApp(logger, cfg)
	app.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	app.Stop()

	return nil
}
