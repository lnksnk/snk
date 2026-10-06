package ui

import (
	_ "embed"
)

//go:embed parser.js
var ParserJS string

//go:embed index.html
var IndexHTML string
