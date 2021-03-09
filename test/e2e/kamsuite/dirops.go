package kamsuite

import (
	"fmt"
	"os"
	"path/filepath"
)

// BootstrapDirectoryShouldExist checks existing directory, throws error if not found
func BootstrapDirectoryShouldExist(dirName string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	bootstrapDir := filepath.Join(wd, "out", "test-run", dirName)

	if _, err := os.Stat(bootstrapDir); os.IsExist(err) {
		return nil
	}

	return fmt.Errorf("No bootstrap %s directory exists", dirName)
}
