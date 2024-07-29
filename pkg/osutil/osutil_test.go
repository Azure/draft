package osutil

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Azure/draft/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		{"", false},
		{"{{WITH SPACE}}", false},
		{"{{ WithEndSpaces }}", false},
		{"{{.helm.values.style}}", false},
		{"{{.Values.helm.style}}", false},
		{"{{VARIABLE1}}", true},
		{"{{WITH_UNDERSCORE}}", true},
		{"{{mIxEdCase}}", true},
		{"{{lowercase}}", true},
		{"{{snake_case}}", true},
		{"{{VAR1}} and {{VAR2}}", true},
		{"{{.Values.name}} and {{.Values.namespace}}", false},
		{"{{.Values.helm.style}} and {{WITH_UNDERSCORE}}", true},
		{"{{nested {{template}}}}", true},
		{"{{nested .Values.template}}", false},
	}

	for _, test := range tests {
		t.Run(test.String, func(t *testing.T) {
			err := checkAllVariablesSubstituted(test.String)
			didError := err != nil
			assert.Equal(t, test.ExpectError, didError)
		})
	}
}

type MockTemplateWriter struct {
	mock.Mock
	directoriesCreated []string
	filesWritten       map[string][]byte
}

func (m *MockTemplateWriter) EnsureDirectory(dirPath string) error {
	m.directoriesCreated = append(m.directoriesCreated, dirPath)
	args := m.Called(dirPath)
	return args.Error(0)
}

func (m *MockTemplateWriter) WriteFile(filePath string, content []byte) error {
	if m.filesWritten == nil {
		m.filesWritten = make(map[string][]byte)
	}
	m.filesWritten[filePath] = content
	args := m.Called(filePath, content)
	return args.Error(0)
}

func TestCopyDirWithTemplates(t *testing.T) {
	tests := []struct {
		name          string
		fileSys       fs.FS
		src           string
		dest          string
		draftConfig   *config.DraftConfig
		expectedFiles map[string]string
		expectedError error
	}{
		{
			name: "successful copy and template replacement",
			fileSys: fstest.MapFS{
				"src/file1.txt":        &fstest.MapFile{Data: []byte("Hello, {{.Name}}!")},
				"src/subdir/file2.txt": &fstest.MapFile{Data: []byte("Welcome to {{.Place}}.")},
			},
			src:  "src",
			dest: "dest",
			draftConfig: &config.DraftConfig{
				Variables: []*config.BuilderVar{
					{Name: "Name", Value: "Joe"},
					{Name: "Place", Value: "Paris"},
				},
			},
			expectedFiles: map[string]string{
				"dest/file1.txt":        "Hello, Joe!",
				"dest/subdir/file2.txt": "Welcome to Paris.",
			},
			expectedError: nil,
		},
		{
			name: "missing variable",
			fileSys: fstest.MapFS{
				"src/file1.txt": &fstest.MapFile{Data: []byte("Hello, {{.Name}}!")},
			},
			src:  "src",
			dest: "dest",
			draftConfig: &config.DraftConfig{
				Variables: []*config.BuilderVar{},
			},
			expectedFiles: nil,
			expectedError: fmt.Errorf("variable map is empty, unable to replace template variables"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateWriter := new(MockTemplateWriter)
			if tt.expectedError == nil {
				for path, content := range tt.expectedFiles {
					templateWriter.On("WriteFile", path, []byte(content)).Return(nil)
				}
				templateWriter.On("EnsureDirectory", mock.Anything).Return(nil)
			}

			err := CopyDirWithTemplates(tt.fileSys, tt.src, tt.dest, tt.draftConfig, templateWriter)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				for path, content := range tt.expectedFiles {
					templateWriter.AssertCalled(t, "WriteFile", path, []byte(content))
				}
			}
		})
	}
}

func TestReplaceGoTemplateVariables(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		variableMap   map[string]string
		expected      string
		expectedError bool
		fileExists    bool
	}{
		{
			name:        "simple template substitution",
			fileContent: "Hello, {{.Name}}!",
			variableMap: map[string]string{
				"Name": "Joe",
			},
			expected:      "Hello, Joe!",
			expectedError: false,
			fileExists:    true,
		},
		{
			name:        "missing variable in map",
			fileContent: "Hello, {{.Joe}}!",
			variableMap: map[string]string{
				"Other": "Joe",
			},
			expected:      "Hello, <no value>!",
			expectedError: false,
			fileExists:    true,
		},
		{
			name:          "syntax error in template",
			fileContent:   "Hello, {{.Name",
			variableMap:   map[string]string{},
			expected:      "",
			expectedError: true,
			fileExists:    true,
		},
		{
			name:        "multiple variables",
			fileContent: "Hello, {{.FirstName}} {{.LastName}}!",
			variableMap: map[string]string{
				"FirstName": "Apple",
				"LastName":  "sauce",
			},
			expected:      "Hello, Apple sauce!",
			expectedError: false,
			fileExists:    true,
		},
		{
			name:        "nested variables",
			fileContent: "{{if .Greeting}}{{.Greeting}}, {{.Name}}!{{end}}",
			variableMap: map[string]string{
				"Greeting": "Hello",
				"Name":     "Joe",
			},
			expected:      "Hello, Joe!",
			expectedError: false,
			fileExists:    true,
		},
		{
			name:          "file not found",
			fileContent:   "",
			variableMap:   map[string]string{},
			expected:      "",
			expectedError: true,
			fileExists:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSys := fstest.MapFS{}
			if tt.fileExists {
				fileSys["template.txt"] = &fstest.MapFile{
					Data: []byte(tt.fileContent),
				}
			}

			result, err := replaceGoTemplateVariables(fileSys, "template.txt", tt.variableMap)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}
