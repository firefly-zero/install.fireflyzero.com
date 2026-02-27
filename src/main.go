package src

import (
	"context"
	"os"
	"os/signal"
)

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	logger := newLogger()
	srv := newDeviceServer(logger)
	go srv.run(ctx)
	runWebServer(srv, logger)
	return nil
}
