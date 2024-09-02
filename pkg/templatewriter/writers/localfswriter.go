package writers

import (
	"os"

	"github.com/bfoley13/draft/pkg/osutil"
)

type LocalFSWriter struct {
	WriteMode os.FileMode
}

func (w *LocalFSWriter) WriteFile(path string, data []byte) error {
	mode := w.WriteMode
	if w.WriteMode == 0 {
		mode = 0644
	}

	return os.WriteFile(path, data, mode)
}
func (w *LocalFSWriter) EnsureDirectory(path string) error {
	return osutil.EnsureDirectory(path)
}
