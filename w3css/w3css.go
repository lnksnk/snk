package w3css

import (
	_ "embed"
)

//go:embed w3.css
var W3CSS string

//go:embed index.html
var IndexHTML string

func LoadW3(fsroot string, fsmap func(fsroot string, fslocalroot string) (err error), fsset func(name string, a ...any) error) {
	if fsroot != "" && fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/' {
		fsmap(fsroot, "")
		fsset(fsroot+"w3.css", W3CSS)
		fsset(fsroot+"index.html", IndexHTML)
	}

}
