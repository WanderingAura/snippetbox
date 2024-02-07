package ui

import (
	"embed"
)

// below is a comment directive. this directive instructs the go compiler to store
// the files from ui/html and ui/static folders in an embed.FS embedded filesystem
// referenced by the global variable Files

//go:embed "html" "static"
var Files embed.FS
