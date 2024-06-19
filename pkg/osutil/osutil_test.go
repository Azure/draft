package osutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	file, err := ioutil.TempFile("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()

	exists, err := Exists(name)
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that exists, got %v", err)
	}
	if !exists {
		t.Error("expected tempfile to exist")
	}
	// on Windows, we need to close all open handles to a file before we remove it.
	file.Close()
	os.Remove(name)
	stillExists, err := Exists(name)
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that does not exist, got %v", err)
	}
	if stillExists {
		t.Error("expected tempfile to NOT exist after removing it")
	}
}

func TestSymlinkWithFallback(t *testing.T) {
	const (
		oldFileName = "foo.txt"
		newFileName = "bar.txt"
	)
	tmpDir, err := ioutil.TempDir("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldFileNamePath := filepath.Join(tmpDir, oldFileName)
	newFileNamePath := filepath.Join(tmpDir, newFileName)

	oldFile, err := os.Create(filepath.Join(tmpDir, oldFileName))
	if err != nil {
		t.Fatal(err)
	}
	oldFile.Close()

	if err := SymlinkWithFallback(oldFileNamePath, newFileNamePath); err != nil {
		t.Errorf("expected no error when calling SymlinkWithFallback() on a file that exists, got %v", err)
	}
}

func TestEnsureDir(t *testing.T) {
	validDir := "./../../test/templates"
	assert.DirExists(t, validDir)

	err := EnsureDirectory(validDir)
	assert.Nil(t, err)

	invalidDir := "./../../test/EnsureDirTest"
	err = EnsureDirectory(invalidDir)

	assert.Nil(t, err)
	assert.DirExists(t, invalidDir)

	os.Remove(invalidDir)
}

func TestEnsureFile(t *testing.T) {
	validFile := "./../../test/templates/ensure_file.yaml"
	assert.FileExists(t, validFile)

	err := EnsureFile(validFile)
	assert.Nil(t, err)

	invalidFile := "./../../test/templates/ensure_file_create.yaml"
	err = EnsureFile(invalidFile)

	assert.Nil(t, err)
	assert.FileExists(t, invalidFile)

	os.Remove(invalidFile)
}

func TestAllVariablesSubstituted(t *testing.T) {
	tests := []struct {
		String      string
		ExpectError bool
	}{
		{"{{WITH SPACE}}", false},
		{"{{ WithEndSpaces }}", false},
		{"{{.helm.values.style}}", false},
		{"{{.Values.helm.style}}", false},
		{"{{VARIABLE1}}", true},
		{"{{WITH_UNDERSCORE}}", true},
		{"{{mIxEdCase}}", true},
		{"{{lowercase}}", true},
		{"{{snake_case}}", true},
	}

	for _, test := range tests {
		t.Run(test.String, func(t *testing.T) {
			err := CheckAllVariablesSubstituted(test.String)
			didError := err != nil
			assert.Equal(t, test.ExpectError, didError)
		})
	}
}
