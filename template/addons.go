package template

import "embed"

var (
	//go:embed all:addons
	Addons embed.FS
)
