package languages

import (
	"github.com/Azure/draft/pkg/osutil"
	"testing"

	"github.com/Azure/draft/template"
	"github.com/stretchr/testify/assert"
)

func TestLanguagesCreateDockerfileFileMap(t *testing.T) {
	templateWriter := &osutil.FileMapWriter{}
	l := CreateLanguagesFromEmbedFS(template.Dockerfiles, "/test/dest/dir")
	err := l.CreateDockerfileForLanguage("go", map[string]string{
		"PORT": "8080",
	}, templateWriter)

	assert.Nil(t, err)
	assert.NotNil(t, templateWriter.FileMap)
	assert.NotNil(t, templateWriter.FileMap["/test/dest/dir/Dockerfile"])
}
