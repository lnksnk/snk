package driver

import (
	"context"
	"database/sql/driver"
)

type Conn interface {
	driver.Conn
	driver.QueryerContext
	driver.ExecerContext
	driver.Validator
}

type internalconn struct {
	goconn driver.Conn
}

func (i *internalconn) Ping(ctx context.Context) (err error) {
	return i.ResetSession(ctx)
}

func (i *internalconn) IsValid() bool {
	return i.goconn != nil
}

func (i *internalconn) ResetSession(ctx context.Context) (err error) {
	if goconn := i.goconn; goconn == nil {
		err = driver.ErrBadConn
	}
	return
}

// Begin implements [driver.Conn].
func (i *internalconn) Begin() (driver.Tx, error) {
	panic("unimplemented")
}

// Close implements [driver.Conn].
func (i *internalconn) Close() (err error) {
	goconn := i.goconn
	i.goconn = nil
	if goconn != nil {
		err = goconn.Close()
	}
	return
}

// Prepare implements [driver.Conn].
func (i *internalconn) Prepare(query string) (driver.Stmt, error) {
	panic("unimplemented")
}

type internalrows struct {
	driver.Rows
	icn *internalconn
}

func (irws *internalrows) Close() (err error) {
	icn, rws := irws.icn, irws.Rows
	irws.icn, irws.Rows = nil, nil
	if rws != nil {
		rws.Close()
	}
	if icn != nil {
		icn.Close()
	}
	return
}

func (i *internalconn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	if goconn := i.goconn; goconn != nil {
		if qryer, _ := goconn.(driver.QueryerContext); qryer != nil {
			if rows, err = qryer.QueryContext(ctx, query, args); err == nil {
				rows = &internalrows{Rows: rows, icn: i}
				return
			}
			return
		}
		var stmnt driver.Stmt
		if stmnt, err = goconn.Prepare(query); err == nil {
			defer stmnt.Close()
			var argvals []driver.Value
			if argsl := len(args); argsl > 0 {
				argvals = make([]driver.Value, argsl)
				for ai := range args {
					argvals[ai] = args[ai].Value
				}
			}
			if stmntgryer, _ := stmnt.(driver.StmtQueryContext); stmntgryer != nil {
				rows, err = stmntgryer.QueryContext(ctx, args)
				rows = &internalrows{Rows: rows, icn: i}
				return
			}
			rows, err = stmnt.Query(argvals)
			rows = &internalrows{Rows: rows, icn: i}
			return
		}
		return
	}
	err = driver.ErrBadConn
	return
}

func (i *internalconn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	if goconn := i.goconn; goconn != nil {
		if exer, _ := goconn.(driver.ExecerContext); exer != nil {
			if result, err = exer.ExecContext(ctx, query, args); err == nil {
				return
			}
			return
		}
		var stmnt driver.Stmt
		if stmnt, err = goconn.Prepare(query); err == nil {
			defer stmnt.Close()
			var argvals []driver.Value
			if argsl := len(args); argsl > 0 {
				argvals = make([]driver.Value, argsl)
				for ai := range args {
					argvals[ai] = args[ai].Value
				}
			}
			if stmntexer, _ := stmnt.(driver.StmtExecContext); stmntexer != nil {
				result, err = stmntexer.ExecContext(ctx, args)
				return
			}
			result, err = stmnt.Exec(argvals)
			return
		}
		return
	}
	err = driver.ErrBadConn
	return
}
