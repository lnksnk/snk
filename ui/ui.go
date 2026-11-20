package ui

import (
	_ "embed"
)

//go:embed snk.js
var SnkJS string

//go:embed snksse.js
var SnkSseJS string

//go:embed snksse.html
var SnkSseHTML string

//go:embed index.html
var IndexHTML string

func LoadUi(fsroot string, fsmap func(fsroot string, fslocalroot string) (err error), fsset func(name string, a ...any) error) {
	if fsroot != "" && fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/' {
		fsmap(fsroot, "")
		fsset(fsroot+"snk.js", SnkJS)
		fsset(fsroot+"index.html", IndexHTML)

		fsmap(fsroot+"sse/", "")
		fsset(fsroot+"sse/snksse.js", SnkSseJS)
		fsset(fsroot+"sse/index.html", SnkSseHTML)

	}
}
