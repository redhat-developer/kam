package ioutils

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// NewFileSystem returns a local filesystem based afero FS implementation.
func NewFilesystem() afero.Fs {
	return afero.NewOsFs()
}

// NewMemoryFilesystem returns an in-memory afero FS implementation.
func NewMemoryFilesystem() afero.Fs {
	return afero.NewMemMapFs()
}

// IsExisting returns bool whether path exists
func IsExisting(fs afero.Fs, path string) (bool, error) {
	fileInfo, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return true, fmt.Errorf("%q: Dir already exists at %s", filepath.Base(path), path)
	}
	return true, fmt.Errorf("%q: File already exists at %s", filepath.Base(path), path)
}
