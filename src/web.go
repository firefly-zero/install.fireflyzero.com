package src

import (
	_ "embed"
	"log/slog"
	"net/http"
	"strconv"
)

//go:embed index.html
var rawIndex string

func runWebServer(devices DeviceServer, logger *slog.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("POST /upload", handleFileUpload(devices, logger))
	err := http.ListenAndServe(":19742", mux)
	if err != http.ErrServerClosed {
		logger.Error("fatal server error", "error", err)
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
		device, found := devices.getDevice(id)
		if !found {
			sendError(w, "no connected device with the given session ID")
			return
		}

		defer devices.disconnectDevice(id)
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
