package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer"
	"github.com/moov-io/ftdc-from-tap-to-auth/log"
)

func main() {
	logger := log.New()
	app := cardpersonalizer.NewApp(logger, cardpersonalizer.DefaultConfig())

	err := app.Start()
	if err != nil {
		logger.Error("Error starting app", "err", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	app.Shutdown()
}
