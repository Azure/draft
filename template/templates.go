package template

import "embed"

var (
	//go:embed all:*
	Templates embed.FS
)