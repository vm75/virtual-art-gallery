// Package version contains build metadata for the gallery binary.
package version

// These values are replaced by release builds with Go linker flags.
var (
	Version   = "dev"
	Revision  = "unknown"
	BuildDate = "unknown"
)
