package fontl

import "embed"

//go:embed index.gotpl
var IndexTemplate string

//go:embed static
var StaticFiles embed.FS
