package template

import "embed"

var (
	//go:embed all:azurePipelines
	AzurePipelines embed.FS
)
