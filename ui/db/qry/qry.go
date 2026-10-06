package qry

import (
	_ "embed"
)

//go:embed index.html
var IndexHTML string

func LoadQry(fsroot string, fsmap func(fsroot string, fslocalroot string) (err error), fsset func(name string, a ...any) error) {
	if fsroot != "" && fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/' {
		fsmap(fsroot, "")
		fsset(fsroot+"index.html", IndexHTML)
	}

}
