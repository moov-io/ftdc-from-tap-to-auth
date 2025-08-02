package log

import (
	"os"
	"time"

	"log/slog"

	"github.com/lmittmann/tint"
)

func New() *slog.Logger {
	return slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		}),
	)
}
