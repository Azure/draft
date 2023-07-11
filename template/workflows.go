package template

import "embed"

var (
	//go:embed all:workflows
	Workflows embed.FS
)
