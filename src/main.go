package src

import (
	"context"
	"os"
	"os/signal"
	"sync"
)

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	logger := newLogger()
	devices := newDeviceServer(logger)
	web := newWebServer(devices, logger)

	// Run servers in background
	// and wait for them to stop before exiting.
	// Servers exit when the context is canceled
	// (which is triggered by SIGINT).
	wg := &sync.WaitGroup{}
	wg.Go(func() { devices.run(ctx) })
	wg.Go(func() { web.run(ctx) })
	wg.Wait()
	return nil
}
