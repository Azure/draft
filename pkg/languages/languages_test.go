package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

func TestLanguagesCreateDockerfileFileMap(t *testing.T) {
	templateWriter := &writers.FileMapWriter{}
	l := CreateLanguagesFromEmbedFS(template.Dockerfiles, "/test/dest/dir")
	err := l.CreateDockerfileForLanguage("go", map[string]string{
		"PORT": "8080",
	}, templateWriter)

	assert.Nil(t, err)
	assert.NotNil(t, templateWriter.FileMap)
	assert.NotNil(t, templateWriter.FileMap["/test/dest/dir/Dockerfile"])
}
