package writers

import (
	"os"

	"github.com/Azure/draft/pkg/osutil"
)

type LocalFSWriter struct {
	WriteMode os.FileMode
}

func (w *LocalFSWriter) WriteFile(path string, data []byte) error {
	mode := w.WriteMode
	if w.WriteMode == 0 {
		mode = 0644
	}

	// Reject a symlink at the destination before it can be truncated.
	f, err := openFileNoFollow(path, mode)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}
func (w *LocalFSWriter) EnsureDirectory(path string) error {
	return osutil.EnsureDirectory(path)
}
