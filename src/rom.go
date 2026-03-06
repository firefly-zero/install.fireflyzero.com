package src

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	kb = 1024
	mb = 1024 * kb
)

type Meta struct{}

type ROM struct {
	totalSize uint32
	files     []*zip.File
	meta      Meta
}

func readROM(r io.Reader) (ROM, error) {
	archiveFile, err := os.CreateTemp(os.TempDir(), "rom")
	if err != nil {
		return ROM{}, fmt.Errorf("create temp file: %w", err)
	}
	archiveSize, err := io.Copy(archiveFile, r)
	if err != nil {
		return ROM{}, fmt.Errorf("save ROM: %w", err)
	}
	_, err = archiveFile.Seek(0, 0)
	if err != nil {
		return ROM{}, fmt.Errorf("re-open archive: %w", err)
	}
	archive, err := zip.NewReader(archiveFile, archiveSize)
	if err != nil {
		return ROM{}, fmt.Errorf("open archive: %w", err)
	}

	var totalSize uint64
	for _, file := range archive.File {
		if file.UncompressedSize64 > 8*mb {
			return ROM{}, fmt.Errorf("file %s is too big", file.Name)
		}
		totalSize += file.UncompressedSize64
	}
	if totalSize > 64*mb {
		return ROM{}, errors.New("the ROM file is too big")
	}

	metaFile, err := archive.Open("_meta")
	defer func() { _ = metaFile.Close() }()
	if err != nil {
		return ROM{}, fmt.Errorf("open app metadata: %w", err)
	}
	rawMeta, err := io.ReadAll(metaFile)
	if err != nil {
		return ROM{}, fmt.Errorf("read app metadata: %w", err)
	}
	meta, err := parseMeta(rawMeta)
	if err != nil {
		return ROM{}, fmt.Errorf("parse app metadata: %w", err)
	}
	rom := ROM{
		totalSize: uint32(totalSize),
		files:     archive.File,
		meta:      meta,
	}
	return rom, nil
}

func parseMeta(rawMeta []byte) (Meta, error) {
	panic("todo")
}
