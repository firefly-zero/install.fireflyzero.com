package src

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type DeviceConn struct {
	conn net.Conn
	done chan struct{}
}

func (c *DeviceConn) writeFrom(cl int64, r io.Reader) error {
	const writeTimeout = 10 * time.Minute

	defer close(c.done)

	now := time.Now()
	err := c.conn.SetWriteDeadline(now.Add(writeTimeout))
	if err != nil {
		return fmt.Errorf("set write deadline: %v", err)
	}

	buf := make([]uint8, 4)
	binary.LittleEndian.PutUint32(buf, uint32(max(cl, 0)))
	_, err = c.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("write file size: %v", err)
	}

	_, err = io.Copy(c.conn, r)
	if err != nil {
		return fmt.Errorf("copy file: %v", err)
	}
	return nil
}
