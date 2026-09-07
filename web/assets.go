package web

import (
	"embed"
)

// Static contains the browser assets served by the application.
//
//go:embed static
var Static embed.FS
