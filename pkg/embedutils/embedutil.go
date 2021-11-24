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