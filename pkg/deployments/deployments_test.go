package deployments

import (
	"embed"
	"fmt"
	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/fixtures"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"io"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/Azure/draft/template"
	"github.com/stretchr/testify/assert"
)

var testFS embed.FS

func TestCreateDeployments(t *testing.T) {
	dest := "."
	templateWriter := &writers.LocalFSWriter{}
	draftConfig := &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{Name: "APPNAME", Value: "testapp"},
			{Name: "NAMESPACE", Value: "default"},
			{Name: "PORT", Value: "80"},
			{Name: "IMAGENAME", Value: "testimage"},
			{Name: "IMAGETAG", Value: "latest"},
			{Name: "GENERATORLABEL", Value: "draft"},
			{Name: "SERVICEPORT", Value: "80"},
		},
	}

	tests := []struct {
		name         string
		deployType   string
		shouldError  bool
		tempDirPath  string
		tempFileName string
		tempPath     string
		fixturePath  string
		cleanUp      func()
	}{
		{
			name:         "helm",
			deployType:   "helm",
			shouldError:  false,
			tempDirPath:  "charts/templates",
			tempFileName: "charts/templates/deployment.yaml",
			tempPath:     "../../test/templates/helm/charts/templates/deployment.yaml",
			fixturePath:  "../fixtures/deployments/charts/templates/deployment.yaml",
			cleanUp: func() {
				os.Remove(".charts")
			},
		},
		{
			name:         "unsupported",
			deployType:   "unsupported",
			shouldError:  true,
			tempDirPath:  "test/templates/unsupported",
			tempFileName: "test/templates/unsupported/deployment.yaml",
			tempPath:     "test/templates/unsupported/deployment.yaml",
			cleanUp: func() {
				os.Remove("deployments")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("Creating temp file:", tt.tempFileName)
			err := createTempDeploymentFile(tt.tempDirPath, tt.tempFileName, tt.tempPath)
			assert.Nil(t, err)

			deployments := CreateDeploymentsFromEmbedFS(template.Deployments, dest)
			err = deployments.CopyDeploymentFiles(tt.deployType, draftConfig, templateWriter)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				generatedContent, err := os.ReadFile(tt.tempFileName)
				assert.Nil(t, err)

				if _, err := os.Stat(tt.fixturePath); os.IsNotExist(err) {
					t.Errorf("Fixture file does not exist at path: %s", tt.fixturePath)
				}

				err = fixtures.ValidateContentAgainstFixture(generatedContent, tt.fixturePath)
				assert.Nil(t, err)
			}

			tt.cleanUp()
		})
	}
}

func TestLoadConfig(t *testing.T) {
	fakeFS, err := createMockDeploymentTemplatesFS()
	assert.Nil(t, err)

	d, err := createMockDeployments("deployments", fakeFS)
	assert.Nil(t, err)

	cases := []loadConfTestCase{
		{"helm", true},
		{"unsupported", false},
	}

	for _, c := range cases {
		if c.isNil {
			_, err = d.loadConfig(c.deployType)
			assert.Nil(t, err)
		} else {
			_, err = d.loadConfig(c.deployType)
			assert.NotNil(t, err)
		}
	}
}

func TestPopulateConfigs(t *testing.T) {
	fakeFS, err := createMockDeploymentTemplatesFS()
	assert.Nil(t, err)

	d, err := createMockDeployments("deployments", fakeFS)
	assert.Nil(t, err)

	d.PopulateConfigs()
	assert.Equal(t, 3, len(d.configs))

	d, err = createTestDeploymentEmbed("deployments")
	assert.Nil(t, err)

	d.PopulateConfigs()
	assert.Equal(t, 3, len(d.configs))
}

type loadConfTestCase struct {
	deployType string
	isNil      bool
}

func createTempDeploymentFile(dirPath, fileName, path string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	fmt.Printf("file %v\n", file)
	defer file.Close()

	var source *os.File
	source, err = os.Open(path)
	if err != nil {
		return err
	}
	fmt.Printf("source %v\n", source)
	defer source.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return err
	}
	return nil
}

func createMockDeploymentTemplatesFS() (fs.FS, error) {
	rootPath := "deplyments/"
	embedFiles, err := embedutils.EmbedFStoMapWithFiles(template.Deployments, "deployments")
	if err != nil {
		return nil, fmt.Errorf("failed to readDir: %w in embeded files", err)
	}

	mockFS := fstest.MapFS{}

	for path, file := range embedFiles {
		if file.IsDir() {
			mockFS[path] = &fstest.MapFile{Mode: fs.ModeDir}
		} else {
			bytes, err := template.Deployments.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failes to read file: %w", err)
			}
			mockFS[path] = &fstest.MapFile{Data: bytes}
		}
	}

	mockFS[rootPath+"emptyDir"] = &fstest.MapFile{Mode: fs.ModeDir}
	mockFS[rootPath+"corrupted"] = &fstest.MapFile{Mode: fs.ModeDir}
	mockFS[rootPath+"corrupted/draft.yaml"] = &fstest.MapFile{Data: []byte("fake yaml data")}

	return mockFS, nil
}

func createMockDeployments(dirPath string, mockDeployments fs.FS) (*Deployments, error) {
	dest := "."

	deployMap, err := fsToMap(mockDeployments, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed fsToMap: %w", err)
	}

	d := &Deployments{
		deploys:             deployMap,
		dest:                dest,
		configs:             make(map[string]*config.DraftConfig),
		deploymentTemplates: mockDeployments,
	}

	return d, nil
}

func createTestDeploymentEmbed(dirPath string) (*Deployments, error) {
	dest := "."

	deployMap, err := embedutils.EmbedFStoMap(template.Deployments, "deployments")
	if err != nil {
		return nil, fmt.Errorf("failed to create deployMap: %w", err)
	}

	d := &Deployments{
		deploys:             deployMap,
		dest:                dest,
		configs:             make(map[string]*config.DraftConfig),
		deploymentTemplates: template.Deployments,
	}

	return d, nil
}

func fsToMap(fsFS fs.FS, path string) (map[string]fs.DirEntry, error) {
	files, err := fs.ReadDir(fsFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadDir: %w", err)
	}

	mapping := make(map[string]fs.DirEntry)

	for _, f := range files {
		if f.IsDir() {
			mapping[f.Name()] = f
		}
	}

	return mapping, nil
}
