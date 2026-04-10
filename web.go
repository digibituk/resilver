package resilver

import "embed"

//go:embed all:web
var WebFS embed.FS

//go:embed config.json
var DefaultConfig []byte
