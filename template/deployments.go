package template

import "embed"

var (
	//go:embed all:deployments
	Deployments embed.FS
)
