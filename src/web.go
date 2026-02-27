package src

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

//go:embed index.html
var rawIndex string

type WebServer struct {
	server *http.Server
	logger *slog.Logger
}

func newWebServer(devices DeviceServer, logger *slog.Logger) WebServer {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("POST /upload", handleFileUpload(devices, logger))
	return WebServer{
		server: &http.Server{
			Addr:    ":19742",
			Handler: mux,
		},
		logger: logger,
	}
}

func (srv WebServer) run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		srv.logger.Info("draining HTTP connections...")
		subCtx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
		defer cancel()
		err := srv.server.Shutdown(subCtx)
		if err != nil {
			srv.logger.Error("HTTP server shutdown error", "error", err)
		}
	}()

	srv.logger.Info("listening for HTTP connections...")
	err := srv.server.ListenAndServe()
	if err != http.ErrServerClosed {
		srv.logger.Error("fatal server error", "error", err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(rawIndex))
}

func handleFileUpload(devices DeviceServer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Find an open device connection by the given ID.
		rawID := r.PathValue("id")
		id64, err := strconv.ParseUint(rawID, 10, 32)
		if err != nil {
			sendError(w, "invalid session ID")
			return
		}
		id := uint32(id64)
		device, found := devices.popDevice(id)
		if !found {
			sendError(w, "no connected device with the given session ID")
			return
		}

		defer r.Body.Close()
		err = device.writeFrom(r.Body)
		if err != nil {
			sendError(w, "failed to send file to the device")
			logger.Warn("send to the device", "error", err)
			return
		}
	}
}

func sendError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(msg))
}
