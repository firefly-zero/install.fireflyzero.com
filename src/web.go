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
	mux.HandleFunc("POST /upload/{id}", handleFileUpload(devices, logger))
	mux.HandleFunc("GET /download/{id}", handleFileDownload(devices, logger))
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

// GET /
//
// Serve the static index page HTML.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(rawIndex))
}

// POST /upload/{id}
//
// Upload a ROM archive from client to server and then to the connected device.
func handleFileUpload(devices DeviceServer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength < 0 {
			sendError(w, "Content-Length is required")
			return
		}
		if r.ContentLength <= 20 {
			sendError(w, "the file is too small")
			return
		}
		if r.ContentLength > 8*mb {
			sendError(w, "ROM is too big")
			return
		}

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
			var id2 uint32
			for id3 := range devices.devices {
				id2 = id3
			}
			println(id, id2)
			sendError(w, "no connected device with the given session ID")
			return
		}

		defer r.Body.Close()
		rom, err := readROM(r.Body)
		if err != nil {
			sendError(w, err.Error())
			logger.Warn("failed to read ROM", "error", err)
			return
		}

		err = device.writeROM(rom)
		if err != nil {
			sendError(w, "failed to send file to the device")
			logger.Warn("send to the device", "error", err)
			return
		}
	}
}

// GET /download/{id}
//
// Download file from server to the connected client (device).
func handleFileDownload(devices DeviceServer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		id64, err := strconv.ParseUint(rawID, 10, 32)
		if err != nil {
			sendError(w, "invalid session ID")
			return
		}
		if id64 >= 100_000_000 {
			sendError(w, "invalid session ID")
			return
		}
		id := uint32(id64)

		// Register the device connection by its ID.
		done := make(chan struct{})
		devices.mx.Lock()
		devices.devices[id] = DeviceConn{
			w:    w,
			done: done,
		}
		devices.mx.Unlock()

		// Exit if the connection is used or if it timed out.
		select {
		case <-time.After(10 * time.Minute):
		case <-done:
		}
	}
}

func sendError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(msg))
}
