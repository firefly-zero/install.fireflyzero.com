package main

import (
	"log"

	"github.com/firefly-zero/install.fireflyzero.com/src"
)

func main() {
	err := src.Run()
	if err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}
