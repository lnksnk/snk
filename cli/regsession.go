package main

import (
	"context"

	"github.com/lnksnk/snk/fs"
	"github.com/lnksnk/snk/parameters"
	snksql "github.com/lnksnk/snk/sql"
	"github.com/medama-io/go-useragent"
)

type Session interface {
	UA() useragent.UserAgent
	DB() DbHandler
	FS() Fsys
	Close() error
}

func InvokeSession(dh DbHandler, fss Fsys, ua useragent.UserAgent) Session {
	return &session{dh: dh, fss: fss, ua: ua}
}

type session struct {
	ua  useragent.UserAgent
	dh  DbHandler
	fss Fsys
}

func (sn *session) Close() (err error) {
	sn.dh, sn.fss = nil, nil
	return
}

func (sn *session) UA() useragent.UserAgent {
	return sn.ua
}

func (sn *session) DB() DbHandler {
	return sn.dh
}

func (sn *session) FS() Fsys {
	return sn.fss
}

type Fsys interface {
	Map(fsroot, fslocalroot string) (err error)
	List(path string) ([]fs.FileInfo, error)
	Stat(path string) (fs.FileInfo, error)
	Open(path string) (fs.File, error)
	Set(name string, a ...any) error
}

func InvokeFSys(fslist fs.ListFileSystem,
	fsmap func(fsroot, fslocalroot string) (err error),
	fsset fs.SetFileSystem,
	fsstat fs.StatFileSystem,
	fsopen fs.OpenFileSystem,
	fsappend fs.AppendFileSystem) Fsys {
	return &fsys{fslist: fslist, fsmap: fsmap, fsset: fsset, fsstat: fsstat, fsopen: fsopen, fsappend: fsappend}
}

type fsys struct {
	fslist   fs.ListFileSystem
	fsmap    func(fsroot, fslocalroot string) (err error)
	fsset    fs.SetFileSystem
	fsstat   fs.StatFileSystem
	fsopen   fs.OpenFileSystem
	fsappend fs.AppendFileSystem
}

func (fss *fsys) Map(fsroot, fslocalroot string) (err error) {
	return fss.fsmap(fsroot, fslocalroot)
}

func (fss *fsys) List(path string) ([]fs.FileInfo, error) {
	return fss.fslist.List(path)
}

func (fss *fsys) Stat(path string) (fs.FileInfo, error) {
	return fss.fsstat.Stat(path)
}

func (fss *fsys) Open(path string) (fs.File, error) {
	return fss.fsopen.Open(path)
}

func (fss *fsys) Set(name string, a ...any) error {
	return fss.fsset.Set(name, a...)
}

func (fss *fsys) Append(name string, a ...any) error {
	return fss.fsappend.Append(name, a...)
}

type DbHandler interface {
	Query(name string, query string, a ...any) (records func(func(snksql.Record, int64) bool), err error)
	Exec(name string, query string, a ...any) (any, error)
	Stats(name string) (stats any)
	Conns() []string
	DefineConn(name string, datasource string) (conndef snksql.ConnDefinition, err error)
}

func InvokeDbHandler(ctx context.Context, formatQuery snksql.FormatQueryFunc, params parameters.Parameters) DbHandler {
	return &dbhandler{ctx: ctx, formatQuery: formatQuery, params: params}
}

type dbhandler struct {
	ctx         context.Context
	formatQuery snksql.FormatQueryFunc
	params      parameters.Parameters
}

func (dh *dbhandler) Query(name string, query string, a ...any) (records func(func(snksql.Record, int64) bool), err error) {
	frmtsqlqry, params := dh.formatQuery, dh.params
	if len(a) > 0 && frmtsqlqry != nil {
		a = append([]any{frmtsqlqry}, a...)
	}
	if len(a) == 0 && frmtsqlqry != nil {
		a = append(a, frmtsqlqry)
	}
	if params != nil {
		a = append(a, params)
	}
	var rws, rwserr = snksql.QueryContext(dh.ctx, name, query, a...)
	if err = rwserr; err != nil {
		records = snksql.NumberedRecords(nil)
		return
	}
	records = snksql.NumberedRecords(rws)
	return
}

func (dh *dbhandler) Exec(name string, query string, a ...any) (any, error) {
	frmtsqlqry, params := dh.formatQuery, dh.params
	if len(a) > 0 && frmtsqlqry != nil {
		a = append([]any{frmtsqlqry}, a...)
	}
	if len(a) == 0 && frmtsqlqry != nil {
		a = append(a, frmtsqlqry)
	}
	if params != nil {
		a = append(a, params)
	}
	return snksql.ExecContext(dh.ctx, name, query, a...)
}

func (dh *dbhandler) Stats(name string) (stats any) {
	return snksql.ConnStats(name)
}

func (dh *dbhandler) Conns() []string {
	return snksql.ConnDefinitions()
}

func (dh *dbhandler) DefineConn(name string, datasource string) (conndef snksql.ConnDefinition, err error) {
	return snksql.DefineConn(name, datasource)
}
