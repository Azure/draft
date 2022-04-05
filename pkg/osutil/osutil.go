package osutil

import (
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"syscall"
	"os/exec"
	

	"github.com/Azure/draftv2/pkg/configs"
	log "github.com/sirupsen/logrus"
)

// Exists returns whether the given file or directory exists or not.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// SymlinkWithFallback attempts to symlink a file or directory, but falls back to a move operation
// in the event of the user not having the required privileges to create the symlink.
func SymlinkWithFallback(oldname, newname string) (err error) {
	err = os.Symlink(oldname, newname)
	if runtime.GOOS == "windows" {
		// If creating the symlink fails on Windows because the user
		// does not have the required privileges, ignore the error and
		// fall back to renaming the file.
		//
		// ERROR_PRIVILEGE_NOT_HELD is 0x522:
		// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681385(v=vs.85).aspx
		if lerr, ok := err.(*os.LinkError); ok && lerr.Err == syscall.Errno(0x522) {
			err = os.Rename(oldname, newname)
		}
	}
	return
}

// EnsureDirectory checks if a directory exists and creates it if it doesn't
func EnsureDirectory(dir string) error {
	if fi, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %s", dir, err)
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("%s must be a directory", dir)
	}

	return nil
}

// EnsureFile checks if a file exists and creates it if it doesn't
func EnsureFile(file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		f, err := os.Create(file)
		if err != nil {
			return fmt.Errorf("could not create %s: %s", file, err)
		}
		defer f.Close()
	} else if fi.IsDir() {
		return fmt.Errorf("%s must not be a directory", file)
	}

	return nil
}

func CopyDir(
	fileSys fs.FS,
	src, dest string,
	config *configs.DraftConfig,
	customInputs map[string]string) error {
	files, err := fs.ReadDir(fileSys, src)
	if err != nil {
		return err
	}

	for _, f := range files {

		if f.Name() == "draft.yaml" {
			continue
		}

		srcPath := src + "/" + f.Name()
		destPath := dest + "/" + f.Name()

		if f.IsDir() {
			if err = EnsureDirectory(destPath); err != nil {
				return err
			}
			if err = CopyDir(fileSys, srcPath, destPath, config, customInputs); err != nil {
				return err
			}
		} else {
			file, err := fs.ReadFile(fileSys, srcPath)
			if err != nil {
				return err
			}

			fileString := string(file)

			for oldString, newString := range customInputs {
				log.Debugf("replacing %s with %s", oldString, newString)
				fileString = strings.ReplaceAll(fileString, "{{"+oldString+"}}", newString)
			}

			fileName := f.Name()

			if config != nil {
				log.Debugf("checking name override for srcPath: %s, destPath: %s, destPrefix: %s/",
					srcPath, destPath, dest)
				if prefix := config.GetNameOverride(fileName); prefix != "" {
					log.Debugf("overriding file: %s with prefix: %s", destPath, prefix)
					fileName = fmt.Sprintf("%s%s", prefix, fileName)
				}
			}

			if err = os.WriteFile(fmt.Sprintf("%s/%s", dest, fileName), []byte(fileString), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func CheckAzCliInstalled()  {
	azCmd := exec.Command("az")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: AZ cli not installed. Find installation instructions at this link: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
	}
}

func IsLoggedInToAz() bool {
	azCmd := exec.Command("az", "ad", "signed-in-user", "show", "--only-show-errors", "--query", "objectId")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}


func HasGhCli() bool {
	ghCmd := exec.Command("gh")
	_, err := ghCmd.CombinedOutput()
	if err != nil {
		// TODO: install gh cli?
		log.Fatal("Error: The github cli is required to complete this process. Find installation instructions at this link: https://cli.github.com/manual/installation")
		return false
	}

	return true
}

func IsLoggedInToGh() bool {
	ghCmd := exec.Command("gh", "auth", "status")
	out, err := ghCmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		return false
	}

	return true

}