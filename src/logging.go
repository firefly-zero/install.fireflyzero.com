package src

import (
	"log/slog"
	"os"
)

func newLogger() *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, nil)
	return slog.New(h)
}
