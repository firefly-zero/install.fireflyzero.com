package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}

func run() error {
	conn, err := net.Dial("tcp", "127.0.0.1:19743")
	if err != nil {
		return fmt.Errorf("dial: %v", err)
	}
	id := rand.Uint32() % 100_000_000
	fmt.Printf("session ID: %8d\n", id)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, id)
	_, _ = conn.Write(buf)

	fmt.Println("waiting for ROM...")
	rom, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("download ROM: %v", err)
	}
	printROM(rom)
	return nil
}

func printROM(rom []byte) {
	fmt.Printf("received:         %d bytes\n", len(rom))
	fmt.Printf("protocol version: %d\n", rom[0])
	fmt.Printf("body size:    	  %d bytes\n", binary.LittleEndian.Uint32(rom[1:]))
	rom = rom[5:]

	size := binary.LittleEndian.Uint32(rom)
	rom = rom[4:]
	fmt.Printf("author ID:        %s\n", string(rom[:size]))
	rom = rom[size:]

	size = binary.LittleEndian.Uint32(rom)
	rom = rom[4:]
	fmt.Printf("app ID:           %s\n", string(rom[:size]))
	rom = rom[size:]

	for len(rom) != 0 {
		size = binary.LittleEndian.Uint32(rom)
		rom = rom[4:]
		name := string(rom[:size])
		rom = rom[size:]

		size = binary.LittleEndian.Uint32(rom)
		rom = rom[4:]
		rom = rom[size:]

		fmt.Printf("%s: %d bytes\n", name, size)
	}
}
