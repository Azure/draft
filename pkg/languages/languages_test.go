package languages

import (
	"testing"

	"github.com/Azure/draft/template"
	"github.com/stretchr/testify/assert"
)

func TestLanguagesCreateDockerfileFileMap(t *testing.T) {
	l := CreateLanguagesFromEmbedFS(template.Dockerfiles, "/test/dest/dir")
	fileMap, err := l.GenerateDockerfileFileMapForLanguage("go", map[string]string{
		"PORT": "8080",
	})

	assert.Nil(t, err)
	assert.NotNil(t, fileMap)
	assert.NotNil(t, fileMap["/test/dest/dir/Dockerfile"])
}
