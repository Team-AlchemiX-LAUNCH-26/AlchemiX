package models

import (
	"embed"
	"fmt"
	"path/filepath"
)

// EmbeddedFiles makes model artifacts available in both development and
// packaged Wails builds, where the working directory may be different.
//
//go:embed *.json
var EmbeddedFiles embed.FS

func Read(name string) ([]byte, error) {
	clean := filepath.Base(name)
	data, err := EmbeddedFiles.ReadFile(clean)
	if err != nil {
		return nil, fmt.Errorf("read embedded model %q: %w", clean, err)
	}
	return data, nil
}

func MustRead(name string) []byte {
	data, err := Read(name)
	if err != nil {
		panic(err)
	}
	return data
}
