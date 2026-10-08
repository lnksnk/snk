package db

import (
	_ "embed"
)

//go:embed exec.html
var ExecHTML string

//go:embed stats.html
var StatsHTML string

//go:embed qry.html
var QryHTML string

//go:embed data.html
var DataHTML string

//go:embed init.html
var InitHTML string

//go:embed finit.html
var FinitHTML string

//go:embed first.html
var FirstHTML string

//go:embed last.html
var LastHTML string

func LoadQry(fsroot string, fsmap func(fsroot string, fslocalroot string) (err error), fsset func(name string, a ...any) error) {
	if fsroot != "" && fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/' {
		fsmap(fsroot, "")
		fsset(fsroot+"qry.html", QryHTML)
		fsset(fsroot+"init.html", InitHTML)
		fsset(fsroot+"first.html", FirstHTML)
		fsset(fsroot+"data.html", DataHTML)
		fsset(fsroot+"last.html", LastHTML)
		fsset(fsroot+"finit.html", FinitHTML)

		fsset(fsroot+"exec.html", ExecHTML)
		fsset(fsroot+"stats.html", StatsHTML)
	}

}
