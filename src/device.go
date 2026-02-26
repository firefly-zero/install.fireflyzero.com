package src

import (
	"fmt"
	"io"
	"net"
	"time"
)

type DeviceConn struct {
	conn net.Conn
	done chan struct{}
}

func (c *DeviceConn) writeFrom(r io.Reader) error {
	const writeTimeout = 10 * time.Minute

	defer close(c.done)

	now := time.Now()
	err := c.conn.SetWriteDeadline(now.Add(writeTimeout))
	if err != nil {
		return fmt.Errorf("set write deadline: %v", err)
	}
	_, err = io.Copy(c.conn, r)
	if err != nil {
		return fmt.Errorf("copy: %v", err)
	}
	return nil
}
