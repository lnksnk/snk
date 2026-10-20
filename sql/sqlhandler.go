package sql

import (
	"context"

	"github.com/lnksnk/snk/parameters"
)

type SQLHandler interface {
	Query(name string, query string, a ...any) (records func(func(Record, int64) bool), err error)
	Exec(name string, query string, a ...any) (any, error)
	Stats(name string) (stats any)
	Conns() []string
	DefineConn(name string, datasource string) (conndef ConnDefinition, err error)
}

func InvokeSQLHandler(ctx context.Context, formatQuery FormatQueryFunc, params parameters.Parameters) SQLHandler {
	return &sqlhandler{ctx: ctx, formatQuery: formatQuery, params: params}
}

type sqlhandler struct {
	ctx         context.Context
	formatQuery FormatQueryFunc
	params      parameters.Parameters
}

func (dh *sqlhandler) Query(name string, query string, a ...any) (records func(func(Record, int64) bool), err error) {
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
	var rws, rwserr = QueryContext(dh.ctx, name, query, a...)
	if err = rwserr; err != nil {
		records = NumberedRecords(nil)
		return
	}
	records = NumberedRecords(rws)
	return
}

func (dh *sqlhandler) Exec(name string, query string, a ...any) (any, error) {
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
	return ExecContext(dh.ctx, name, query, a...)
}

func (dh *sqlhandler) Stats(name string) (stats any) {
	return ConnStats(name)
}

func (dh *sqlhandler) Conns() []string {
	return ConnDefinitions()
}

func (dh *sqlhandler) DefineConn(name string, datasource string) (conndef ConnDefinition, err error) {
	return DefineConn(name, datasource)
}
