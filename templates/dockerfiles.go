package templates

import "embed"

var (
	//go:embed all:dockerfiles
	DockerfileTemplates embed.FS
)
