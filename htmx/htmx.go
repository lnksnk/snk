package htmx

import (
	_ "embed"
)

//go:embed htmax.js
var HtmaxJS string

//go:embed index.html
var IndexHTML string

func LoadHTMX(fsroot string, fsmap func(fsroot string, fslocalroot string) (err error), fsset func(name string, a ...any) error) {
	if fsroot != "" && fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/' {
		fsmap(fsroot, "")
		fsset(fsroot+"htmax.js", HtmaxJS)
		fsset(fsroot+"index.html", IndexHTML)
	}

}
