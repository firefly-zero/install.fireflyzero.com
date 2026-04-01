package src

import (
	"archive/zip"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/orsinium-labs/postcard"
)

const (
	kb = 1024
	mb = 1024 * kb
)

type Meta struct {
	appID      string
	appName    string
	authorID   string
	authorName string
	launcher   bool
	sudo       bool
	version    uint32
}

func init() {
	zip.RegisterDecompressor(93, zstd.ZipDecompressor())
}

func (m *Meta) parse(r io.Reader) error {
	parse := postcard.Struct(
		postcard.Str(&m.appID),
		postcard.Str(&m.appName),
		postcard.Str(&m.authorID),
		postcard.Str(&m.authorName),
		postcard.Bool(&m.launcher),
		postcard.Bool(&m.sudo),
		postcard.U32(&m.version),
	)
	return parse(r)
}

type ROM struct {
	// The total uncompressed size of all files in the archive.
	totalSize uint32
	// The parsed app metadata.
	meta Meta
	// List of files in the archive.
	files []*zip.File
	// Close archive file descriptor and remove the temporary archive file.
	close func()
}

func readROM(r io.Reader) (ROM, error) {
	// Create temporary file to save the zip archive to.
	archiveFile, err := os.CreateTemp(os.TempDir(), "rom")
	if err != nil {
		return ROM{}, fmt.Errorf("create temp file: %w", err)
	}
	shouldClose := true
	close := func() {
		_ = archiveFile.Close()
		_ = os.Remove(archiveFile.Name())
	}
	defer func() {
		if shouldClose {
			close()
		}
	}()

	// Write the archive into the temporary file and open it.
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

	// Validate sizes and names.
	var totalSize uint64
	names := map[string]struct{}{}
	for _, file := range archive.File {
		if file.UncompressedSize64 > 8*mb {
			return ROM{}, fmt.Errorf("file %s is too big", file.Name)
		}
		if strings.HasSuffix(file.Name, "/") {
			return ROM{}, errors.New("the archive contains a directory")
		}
		if strings.Contains(file.Name, "/") {
			return ROM{}, errors.New("the archive contains a nested file")
		}
		totalSize += file.UncompressedSize64
		names[file.Name] = struct{}{}
	}
	if totalSize < 10 {
		return ROM{}, errors.New("the ROM is empty")
	}
	if totalSize > 64*mb {
		return ROM{}, errors.New("the ROM file is too big")
	}
	if _, ok := names["_bin"]; !ok {
		return ROM{}, errors.New("file not found: _bin")
	}
	if _, ok := names["_meta"]; !ok {
		return ROM{}, errors.New("file not found: _meta")
	}
	if _, ok := names["_stats"]; !ok {
		return ROM{}, errors.New("file not found: _stats")
	}
	if _, ok := names["_hash"]; !ok {
		return ROM{}, errors.New("file not found: _hash")
	}

	// Parse and validate app metadata.
	metaFile, err := archive.Open("_meta")
	if err != nil {
		return ROM{}, fmt.Errorf("open app metadata: %w", err)
	}
	defer func() {
		_ = metaFile.Close()
	}()
	var meta Meta
	err = meta.parse(metaFile)
	if err != nil {
		return ROM{}, fmt.Errorf("parse app metadata: %w", err)
	}
	if meta.sudo && meta.authorID != "sys" {
		return ROM{}, fmt.Errorf("only system apps can use sudo: %w", err)
	}

	// Sort files by name.
	// The client needs the files sorted so that it can calculate
	// the ROM hash on the fly.
	files := archive.File
	slices.SortStableFunc(files, func(a, b *zip.File) int {
		return cmp.Compare(a.Name, b.Name)
	})

	shouldClose = false
	rom := ROM{
		totalSize: uint32(totalSize),
		files:     files,
		meta:      meta,
		close:     close,
	}
	return rom, nil
}
