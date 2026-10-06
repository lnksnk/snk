package fssql

import (
	"database/sql"

	snkfs "github.com/lnksnk/snk/fs"
	snksql "github.com/lnksnk/snk/sql"
)

type ConnDefListDirFileSystem interface {
	snkfs.ListFileSystem
	QueryList(string, ...any) ([]snkfs.FileInfo, error)
}

type ConnDefReadDirFileSystem interface {
	snkfs.ReadDirFileSystem
	QueryReadDir(string, ...any) ([]snkfs.DirEntry, error)
}

type ConnDefStatFileSystem interface {
	snkfs.StatFileSystem
	QueryStat(string, ...any) (snkfs.FileInfo, error)
}

type ConnDefOpenFileSystem interface {
	snkfs.OpenFileSystem
	QueryOpen(string, ...any) (snkfs.File, error)
}

type ConnDefSetFileSystem interface {
	snkfs.SetFileSystem
	ExecSet(name string, a ...any) error
}

type ConnDefAppendFileSystem interface {
	snkfs.AppendFileSystem
	ExecAppend(name string, a ...any) error
}

type conndefinition struct {
	conndef snksql.ConnDefinition
	ConnDefListDirFileSystem
	ConnDefStatFileSystem
	ConnDefOpenFileSystem
	ConnDefReadDirFileSystem
	ConnDefSetFileSystem
	ConnDefAppendFileSystem
}

func (cf *conndefinition) Query(query string, a ...any) (rows snksql.SqlRows, err error) {

	return
}

func (cf *conndefinition) Exec(query string, a ...any) (result sql.Result, err error) {

	return
}

func FSys(conndef snksql.ConnDefinition) snkfs.FileSystem {
	return &conndefinition{conndef: conndef}
}
