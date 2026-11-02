package path

import (
	"os"
	"path/filepath"
)

// GetProjectRoot looks for the project root directory by the presence of the go.work file.
func GetProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("failed to get the working directory: " + err.Error())
	}

	for {
		_, err = os.Stat(filepath.Join(dir, "go.work"))
		if err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("failed to find the project root (go.work)")
		}

		dir = parent
	}
}
