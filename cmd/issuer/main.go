package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/moov-io/ftdc-from-tap-to-auth/internal/config"
	"github.com/moov-io/ftdc-from-tap-to-auth/issuer"
	"github.com/moov-io/ftdc-from-tap-to-auth/log"
)

func main() {
	cfg := &issuer.Config{}

	err := config.NewFromFile("configs/issuer.yaml", cfg)
	if err != nil {
		log.New().Error("Error loading config", "err", err)
		os.Exit(1)
	}

	logger := log.New()
	app := issuer.NewApp(logger, cfg)

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
