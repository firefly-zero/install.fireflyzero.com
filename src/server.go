package src

import (
	"net/http"
	"strconv"
)

func handleFileUpload(srv DeviceServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Find an open device connection by the given ID.
		rawID := r.PathValue("id")
		id64, err := strconv.ParseUint(rawID, 10, 32)
		if err != nil {
			// ...
		}
		id := uint32(id64)
		device, found := srv.getDevice(id)
		if !found {
			// ...
		}

		defer srv.disconnectDevice(id)
		defer r.Body.Close()
		err = device.writeFrom(r.Body)
		if err != nil {
			// ...
		}
	}
}
