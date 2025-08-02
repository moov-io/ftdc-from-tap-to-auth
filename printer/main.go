package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		}),
	)

	printer, err := NewThermalPrinter()
	if err != nil {
		log.Fatalf("Failed to initialize printer: %v", err)
	}
	defer printer.Close()

	service := NewService(logger, printer) // Replace nil with actual printer initialization if needed
	service.Start()

	srv := NewServer(logger, service)
	srv.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	srv.Shutdown()
	service.Stop()
}
