package kamsuite

import (
	"fmt"
	"os"
)

// DirectoryShouldExist checks existing directory, throws error if not found
func DirectoryShouldExist(dirName string) error {
	if _, err := os.Stat(dirName); os.IsExist(err) {
		return nil
	}

	return fmt.Errorf("directory %s exists", dirName)
}
