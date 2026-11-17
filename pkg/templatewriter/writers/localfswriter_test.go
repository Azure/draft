package writers

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalFSWriterWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	w := &LocalFSWriter{}
	require.NoError(t, w.WriteFile(path, []byte("original contents")))
	require.NoError(t, w.WriteFile(path, []byte("new")))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "new", string(data))
}

func TestLocalFSWriterWriteFileRejectsSymlink(t *testing.T) {
	for _, existing := range []bool{true, false} {
		name := "dangling target"
		if existing {
			name = "existing target"
		}
		t.Run(name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "outside.txt")
			if existing {
				require.NoError(t, os.WriteFile(target, []byte("keep me"), 0600))
			}
			path := filepath.Join(t.TempDir(), "service.yaml")
			err := os.Symlink(target, path)
			if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
				t.Skip("creating symlinks requires privileges on Windows")
			}
			require.NoError(t, err)

			w := &LocalFSWriter{}
			require.Error(t, w.WriteFile(path, []byte("generated manifest")))
			linkTarget, err := os.Readlink(path)
			require.NoError(t, err)
			require.Equal(t, target, linkTarget)
			data, err := os.ReadFile(target)
			if existing {
				require.NoError(t, err)
				require.Equal(t, "keep me", string(data))
			} else {
				require.ErrorIs(t, err, os.ErrNotExist)
			}
		})
	}
}
