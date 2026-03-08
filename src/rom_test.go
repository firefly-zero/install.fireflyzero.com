package src

import (
	"fmt"
	"os"
	"testing"
)

func eq[T comparable](a, b T) {
	if a != b {
		panic(fmt.Sprintf("%v != %v", a, b))
	}
}

func TestReadRom(t *testing.T) {
	file, err := os.Open("../test_data/lux.gates.zip")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}
	rom, err := readROM(file)
	if err != nil {
		t.Fatalf("read rom: %v", err)
	}
	eq(len(rom.files), 5)
	eq(rom.meta.authorID, "lux")
	eq(rom.meta.authorName, "Lux")
	eq(rom.meta.appID, "gates")
}
