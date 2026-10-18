package ui

import (
	_ "embed"
)

//go:embed parser.js
var ParserJS string

//go:embed index.html
var IndexHTML string

func LoadUi(fsroot string, fsmap func(fsroot string, fslocalroot string) (err error), fsset func(name string, a ...any) error) {
	if fsroot != "" && fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/' {
		fsmap(fsroot, "")
		fsset(fsroot+"parser.js", ParserJS)
		fsset(fsroot+"index.html", IndexHTML)
	}

}
