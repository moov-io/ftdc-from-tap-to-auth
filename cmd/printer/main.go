package main

import (
	"log"
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

	termPrinter, err := printer.NewThermalPrinter()
	if err != nil {
		log.Fatalf("Failed to initialize printer: %v", err)
	}
	defer termPrinter.Close()

	service := printer.NewService(logger, termPrinter) // Replace nil with actual printer initialization if needed
	service.Start()

	srv := printer.NewServer(logger, service)
	srv.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	srv.Shutdown()
	service.Stop()
}
