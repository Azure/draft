package templates

import "embed"

var (
	//go:embed all:builders
	Builders embed.FS
)
