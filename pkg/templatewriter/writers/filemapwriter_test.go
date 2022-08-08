package writers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/template"
)

func TestCopyDirToFileMap(t *testing.T) {

	templatewriter := &FileMapWriter{}
	err := osutil.CopyDir(template.Dockerfiles, "dockerfiles/javascript", "/test/dir", nil, map[string]string{
		"PORT": "8080",
	}, templatewriter)
	assert.Nil(t, err)
	assert.NotNil(t, templatewriter.FileMap)
	assert.NotNil(t, templatewriter.FileMap["/test/dir/Dockerfile"])
}