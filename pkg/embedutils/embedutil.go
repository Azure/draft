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
	files, err := fs.ReadDir(embedFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to readDir: %w", err)
	}

	mapping := make(map[string]fs.DirEntry)

	for _, f := range files {
		mapping[f.Name()] = f
		if f.IsDir() {
			add, err := EmbedFStoMapWithFiles(embedFS, path+"/"+f.Name())
			if err != nil {
				return nil, err
			}
			for k, v := range add {
				mapping[f.Name()+"/"+k] = v
			}
		} else {
			mapping[f.Name()] = f
		}
	}

	return mapping, nil
}
