package ui

import "embed"

//go:embed all:html
var HTMLFiles embed.FS

//go:embed all:static
var StaticFiles embed.FS
