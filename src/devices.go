package src

import (
	"sync"
)

type DeviceServer struct {
	mx      *sync.Mutex
	devices map[uint32]DeviceConn
}

func newDeviceServer() DeviceServer {
	return DeviceServer{
		mx:      &sync.Mutex{},
		devices: map[uint32]DeviceConn{},
	}
}

func (srv DeviceServer) popDevice(id uint32) (DeviceConn, bool) {
	srv.mx.Lock()
	device, found := srv.devices[id]
	if found {
		delete(srv.devices, id)
	}
	srv.mx.Unlock()
	return device, found
}
