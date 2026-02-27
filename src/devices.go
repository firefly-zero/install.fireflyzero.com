package src

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

type DeviceServer struct {
	mx      *sync.Mutex
	devices map[uint32]DeviceConn
	logger  *slog.Logger
}

func newDeviceServer(logger *slog.Logger) DeviceServer {
	return DeviceServer{
		mx:      &sync.Mutex{},
		devices: map[uint32]DeviceConn{},
		logger:  logger,
	}
}

func (srv DeviceServer) run(ctx context.Context) {
	srv.logger.Info("listening for TCP connections...")
	listener, err := net.Listen("tcp", ":19743")
	if err != nil {
		srv.logger.Error("failed to start device server", "error", err)
	}

	go func() {
		<-ctx.Done()
		srv.logger.Info("draining TCP connections...")
		err := listener.Close()
		if err != nil {
			srv.logger.Error("TCP server shutdown error", "error", err)
		}
	}()

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			srv.logger.Warn("failed to accept connection", "error", err)
		}
		wg.Add(1)
		go srv.handleConnection(wg, conn)
	}
}

func (srv DeviceServer) getDevice(id uint32) (DeviceConn, bool) {
	srv.mx.Lock()
	device, found := srv.devices[id]
	srv.mx.Unlock()
	return device, found
}

func (srv DeviceServer) disconnectDevice(id uint32) {
	srv.mx.Lock()
	delete(srv.devices, id)
	srv.mx.Unlock()
}

func (srv DeviceServer) handleConnection(wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	err := srv.handleConnectionInner(conn)
	if err != nil {
		srv.logger.Warn("error handling device connection", "error", err)
	}
}

func (srv DeviceServer) handleConnectionInner(conn net.Conn) error {
	const readTimeout = 10 * time.Minute

	defer conn.Close()

	now := time.Now()
	err := conn.SetReadDeadline(now.Add(readTimeout))
	if err != nil {
		return fmt.Errorf("set read deadline: %v", err)
	}

	// Read the device ID.
	buf := make([]byte, 4)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return fmt.Errorf("read device ID: %v", err)
	}

	// Register the device connection by its ID.
	done := make(chan struct{})
	id := binary.LittleEndian.Uint32(buf)
	srv.mx.Lock()
	srv.devices[id] = DeviceConn{
		conn: conn,
		done: done,
	}
	srv.mx.Unlock()

	// Exit (and close the coonection in `defer`) if it is used or if it timed out.
	select {
	case <-time.After(10 * time.Minute):
	case <-done:
	}
	return nil
}
