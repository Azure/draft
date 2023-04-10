package embedutils

import (
	"embed"
	"fmt"
	"io/fs"
)

func EmbedFStoMap(embedFS embed.FS, path string) (map[string]fs.DirEntry, error) {
	files, err := embedFS.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to readDir: %w", err)
	}

	mapping := make(map[string]fs.DirEntry)

	for _, f := range files {
		if f.IsDir() {
			mapping[f.Name()] = f
		}
	}

	return mapping, nil
}

func EmbedFStoMapWithFiles(embedFS fs.FS, path string) (map[string]fs.DirEntry, error) {
	mapping := make(map[string]fs.DirEntry)
	err := fs.WalkDir(embedFS, path, func(path string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		mapping[path] = f
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walkDir: %w", err)
	}

	return mapping, nil
}
