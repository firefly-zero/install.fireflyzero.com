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

// Send the given ROM to the device.
//
// Only one ROM can be sent in a single connection.
func (c *DeviceConn) writeROM(rom ROM) error {
	const writeTimeout = 10 * time.Minute

	defer close(c.done)
	defer rom.close()

	now := time.Now()
	err := c.conn.SetWriteDeadline(now.Add(writeTimeout))
	if err != nil {
		return fmt.Errorf("set write deadline: %v", err)
	}

	// Write header.
	_, err = c.conn.Write([]byte{1})
	if err != nil {
		return fmt.Errorf("write protocol version: %v", err)
	}
	err = c.writeU32(rom.totalSize)
	if err != nil {
		return fmt.Errorf("write ROM size: %v", err)
	}
	err = c.writeStr(rom.meta.authorID)
	if err != nil {
		return fmt.Errorf("write author ID: %v", err)
	}
	err = c.writeStr(rom.meta.appID)
	if err != nil {
		return fmt.Errorf("write app ID: %v", err)
	}

	// Write body.
	for _, file := range rom.files {
		err = c.writeStr(file.Name)
		if err != nil {
			return fmt.Errorf("write file name: %v", err)
		}
		err := c.writeU32(uint32(file.UncompressedSize64))
		if err != nil {
			return fmt.Errorf("write file size: %v", err)
		}
		stream, err := file.Open()
		if err != nil {
			return fmt.Errorf("open file %s: %v", file.Name, err)
		}
		_, err = io.Copy(c.conn, stream)
		_ = stream.Close()
		if err != nil {
			return fmt.Errorf("send file: %v", err)
		}
	}

	return nil
}

func (c *DeviceConn) writeStr(v string) error {
	return c.writeBytes([]byte(v))
}

func (c *DeviceConn) writeBytes(v []byte) error {
	err := c.writeU32(uint32(len(v)))
	if err != nil {
		return err
	}
	_, err = c.conn.Write(v)
	return err
}

func (c *DeviceConn) writeU32(v uint32) error {
	buf := make([]uint8, 4)
	binary.LittleEndian.PutUint32(buf, v)
	_, err := c.conn.Write(buf)
	return err
}
