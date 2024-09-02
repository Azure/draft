package writers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bfoley13/draft/pkg/config"
	"github.com/bfoley13/draft/pkg/osutil"
	"github.com/bfoley13/draft/template"
)

func TestCopyDirToFileMap(t *testing.T) {

	templatewriter := &FileMapWriter{}
	err := osutil.CopyDir(template.Dockerfiles, "dockerfiles/javascript", "/test/dir", &config.DraftConfig{
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
	}, templatewriter)
	assert.Nil(t, err)
	assert.NotNil(t, templatewriter.FileMap)
	assert.NotNil(t, templatewriter.FileMap["/test/dir/Dockerfile"])
}
