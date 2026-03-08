package src

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type DeviceConn struct {
	conn net.Conn
	done chan struct{}
}

func (c *DeviceConn) writeROM(rom ROM) error {
	const writeTimeout = 10 * time.Minute

	defer close(c.done)

	now := time.Now()
	err := c.conn.SetWriteDeadline(now.Add(writeTimeout))
	if err != nil {
		return fmt.Errorf("set write deadline: %v", err)
	}

	_, err = c.conn.Write([]byte{1})
	if err != nil {
		return fmt.Errorf("write protocol version: %v", err)
	}
	err = c.writeU32(rom.totalSize)
	if err != nil {
		return fmt.Errorf("write ROM size: %v", err)
	}

	return nil
}

func (c *DeviceConn) writeU32(v uint32) error {
	buf := make([]uint8, 4)
	binary.LittleEndian.PutUint32(buf, v)
	_, err := c.conn.Write(buf)
	return err
}
