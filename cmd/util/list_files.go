package util

import (
	"os"
	"path/filepath"
)

// ListFiles returns all files contained in the given path.
// Directory are excluded.
func ListFiles(path string) ([]string, error) {
	filenames := []string{}
	return filenames, filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filenames = append(filenames, path)
		return nil
	})
}
