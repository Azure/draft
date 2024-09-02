package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bfoley13/draft/pkg/config"
	"github.com/bfoley13/draft/pkg/templatewriter/writers"
	"github.com/bfoley13/draft/template"
)

func TestLanguagesCreateDockerfileFileMap(t *testing.T) {
	templateWriter := &writers.FileMapWriter{}
	l := CreateLanguagesFromEmbedFS(template.Dockerfiles, "/test/dest/dir")
	err := l.CreateDockerfileForLanguage("go", &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name:  "PORT",
				Value: "8080",
			},
			{
				Name:  "VERSION",
				Value: "14",
			},
		},
	}, templateWriter)

	assert.Nil(t, err)
	assert.NotNil(t, templateWriter.FileMap)
	assert.NotNil(t, templateWriter.FileMap["/test/dest/dir/Dockerfile"])
}
