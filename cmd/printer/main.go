package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/moov-io/ftdc-from-tap-to-auth/printer"
)

func main() {
	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		}),
	)

	var prntr printer.Printer
	var err error

	prntr, err = printer.NewThermalPrinter()
	if err != nil {
		logger.Warn("Failed to initialize thermal printer:", slog.String("error", err.Error()))
		logger.Info("Falling back to mock printer...")
		prntr, err = printer.NewMockPrinter()
		if err != nil {
			logger.Error("Failed to initialize mock printer:", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
	defer prntr.Close()

	service := printer.NewService(logger, prntr)
	service.Start()

	srv := printer.NewServer(logger, service)
	srv.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	srv.Shutdown()
	service.Stop()
}
