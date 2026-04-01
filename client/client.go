package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}

func run() error {
	idUint := rand.Uint32() % 100_000_000
	idStr := fmt.Sprintf("%8d", idUint)
	fmt.Println("session ID:", idStr)

	url := "http://127.0.0.1:19742/download/" + idStr
	fmt.Println("waiting for ROM...")
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("send request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.Status)
	}
	fmt.Println("downloading ROM...")

	defer resp.Body.Close()
	rom, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("download ROM: %v", err)
	}
	printROM(resp.Header, rom)
	return nil
}

func printROM(headers http.Header, body []byte) {
	fmt.Println()
	fmt.Printf("protocol version: %s\n", headers.Get("X-F0-Protocol"))
	fmt.Printf("body size:        %d bytes\n", len(body))
	fmt.Printf("content size:     %s bytes\n", headers.Get("X-F0-Size"))
	fmt.Printf("author ID:        %s\n", headers.Get("X-F0-Author"))
	fmt.Printf("app ID:           %s\n", headers.Get("X-F0-App"))
	fmt.Println()

	for len(body) != 0 {
		size := binary.LittleEndian.Uint32(body)
		body = body[4:]
		name := string(body[:size])
		body = body[size:]

		size = binary.LittleEndian.Uint32(body)
		body = body[4:]
		body = body[size:]

		fmt.Printf("%-16s %6d bytes\n", name, size)
	}
}
