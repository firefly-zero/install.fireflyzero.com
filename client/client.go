package main

import (
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
	id := rand.Uint64()
	idStr := fmt.Sprintf("%8d", id%100_000_000)
	fmt.Printf("session ID: %s\n", idStr)
	_, _ = conn.Write([]byte(idStr))

	fmt.Println("waiting for file...")
	file, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("download file: %v", err)
	}
	fmt.Printf("received %d bytes", len(file))

	return nil
}
