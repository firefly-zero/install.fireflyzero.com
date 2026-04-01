package src

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type DeviceConn struct {
	w    http.ResponseWriter
	done chan<- struct{}
}

// Send the given ROM to the device.
//
// Only one ROM can be sent in a single connection.
func (c *DeviceConn) writeROM(rom ROM) error {
	defer close(c.done)
	defer rom.close()

	// Write headers.
	headers := c.w.Header()
	headers.Add("X-F0-Protocol", "1")
	headers.Add("X-F0-Author", rom.meta.authorID)
	headers.Add("X-F0-App", rom.meta.appID)
	headers.Add("X-F0-Today", time.Now().Format(time.DateOnly))
	headers.Add("X-F0-Size", strconv.FormatUint(uint64(rom.totalSize), 10))
	c.w.WriteHeader(http.StatusOK)

	// Write body.
	for _, file := range rom.files {
		err := c.writeStr(file.Name)
		if err != nil {
			return fmt.Errorf("write file name: %v", err)
		}
		err = c.writeU32(uint32(file.UncompressedSize64))
		if err != nil {
			return fmt.Errorf("write file size: %v", err)
		}
		stream, err := file.Open()
		if err != nil {
			return fmt.Errorf("open file %s: %v", file.Name, err)
		}
		_, err = io.Copy(c.w, stream)
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
	_, err = c.w.Write(v)
	return err
}

func (c *DeviceConn) writeU32(v uint32) error {
	buf := make([]uint8, 4)
	binary.LittleEndian.PutUint32(buf, v)
	_, err := c.w.Write(buf)
	return err
}
