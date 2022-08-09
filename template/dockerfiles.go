package template

import "embed"

var (
	//go:embed all:dockerfiles
	Dockerfiles embed.FS
)
