//go:build !windows

package writers

import (
	"os"
	"syscall"
)

func openFileNoFollow(path string, mode os.FileMode) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscall.O_NOFOLLOW, mode)
}
