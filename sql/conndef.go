package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"maps"
	"net/url"
	"strings"

	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/parameters"
	snkdrvr "github.com/lnksnk/snk/sql/driver"
)

type ConnDefinition interface {
	Query(query string, a ...any) (rows SqlRows, err error)
	QueryContext(ctx context.Context, query string, a ...any) (rows SqlRows, err error)
	Exec(query string, a ...any) (result sql.Result, err error)
	ExecContext(ctx context.Context, query string, a ...any) (result sql.Result, err error)
	Close() error
	Stats() sql.DBStats
}

type conndefin struct {
	name       string
	driver     string
	datasource string
	db         *sql.DB
	dvr        snkdrvr.Driver
}

// Stats implements [ConnDefinition].
func (cf *conndefin) Stats() (stats sql.DBStats) {
	if db := cf.db; db != nil {
		stats = db.Stats()
	}
	return
}

type SqlRows interface {
	Close() error
	ColumnTypes() ([]*sql.ColumnType, error)
	Columns() ([]string, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...any) error
}

type gorows struct {
	cn *sql.Conn
	*sql.Rows
}

func SqlRowsConnCloser(rows *sql.Rows, cn *sql.Conn) SqlRows {
	return &gorows{Rows: rows, cn: cn}
}

func (gorws *gorows) Close() (err error) {
	cn, rws := gorws.cn, gorws.Rows
	gorws.cn = nil
	if rws != nil {
		err = rws.Close()
	}
	if cn != nil {
		cn.Close()
	}
	return
}

func (cf *conndefin) Query(query string, a ...any) (rows SqlRows, err error) {
	return cf.QueryContext(context.Background(), query, a...)
}

func (cf *conndefin) QueryContext(ctx context.Context, query string, a ...any) (rows SqlRows, err error) {
	if db := cf.db; db != nil {

		if rows, err = dbquerycn(ctx, cf.name, cf.driver, cf.dvr, query, db, a...); err != nil {
			return
		}
		return
	}
	err = driver.ErrBadConn
	return
}

func (cf *conndefin) Exec(query string, a ...any) (result sql.Result, err error) {
	return cf.ExecContext(context.Background(), query, a...)
}

func (cf *conndefin) ExecContext(ctx context.Context, query string, a ...any) (result sql.Result, err error) {
	if db := cf.db; db != nil {
		if result, err = dbexeccn(ctx, cf.name, cf.driver, cf.dvr, query, db, a...); err != nil {
			return
		}
		return
	}
	err = driver.ErrBadConn
	return
}

func (cf *conndefin) Close() (err error) {
	db := cf.db
	cf.db = nil
	if db != nil {
		db.Close()
	}
	return
}

func dbquerycn(ctx context.Context, name, driver string, dvr snkdrvr.Driver, query string, db *sql.DB, a ...any) (rows SqlRows, err error) {
	var frmtqry FormatQuery
	if ai, al := 0, len(a); al > 0 {
		for ai < al {
			if frmtqryd, frmtqryk := a[ai].(FormatQuery); frmtqryk {
				if frmtqry == nil {
					frmtqry = frmtqryd
				}
				a = append(a[:ai], a[ai+1:]...)
				al--
				continue
			}
			ai++
		}
	}
	var bfqry, _ = snkio.WriterBuffer()
	defer bfqry.Close()
	if glbfmrqry := GlobalFormatQuery; frmtqry == nil && glbfmrqry != nil {
		frmtqry = glbfmrqry
	}
noformat:
	if frmtqry == nil {
		var stmnts = extractstmnts(bfqry, dvr, a...)
		var stmnstl = len(stmnts)
		var cn *sql.Conn
		if stmnstl > 1 {
			if cn, err = db.Conn(ctx); err != nil {
				return
			}
		}
		for sti := range stmnts {
			if sti == len(stmnts)-1 {
				if stmnstl > 1 {
					if rows, err = cn.QueryContext(ctx, stmnts[sti].query, stmnts[sti].args...); err == nil {
						rows = SqlRowsConnCloser(rows.(*sql.Rows), cn)
						return
					}
					cn.Close()
					return
				}
				if rows, err = db.QueryContext(ctx, stmnts[sti].query, stmnts[sti].args...); err == nil {
					rows = SqlRowsConnCloser(rows.(*sql.Rows), nil)
					return
				}
				return
			}
			if _, err = cn.ExecContext(ctx, stmnts[sti].query, stmnts[sti].args...); err != nil {
				cn.Close()
				return
			}
		}
		return
	}

	if err = frmtqry.FprintQuery(bfqry, name, driver, query, a...); err == nil {
		frmtqry = nil
		goto noformat
	}
	return
}

func dbexeccn(ctx context.Context, name, driver string, dvr snkdrvr.Driver, query string, db *sql.DB, a ...any) (result sql.Result, err error) {
	var frmtqry FormatQuery
	if ai, al := 0, len(a); al > 0 {
		for ai < al {
			if frmtqryd, frmtqryk := a[ai].(FormatQuery); frmtqryk {
				if frmtqry == nil {
					frmtqry = frmtqryd
				}
				a = append(a[:ai], a[ai+1:]...)
				al--
				continue
			}
			ai++
		}
	}
	var bfqry, _ = snkio.WriterBuffer()
	defer bfqry.Close()
	if glbfmrqry := GlobalFormatQuery; frmtqry == nil && glbfmrqry != nil {
		frmtqry = glbfmrqry
	}
noformat:
	if frmtqry == nil {
		var stmnts = extractstmnts(bfqry, dvr, a...)
		var stmntsl = len(stmnts)
		var cn *sql.Conn
		if stmntsl > 1 {
			if cn, err = db.Conn(ctx); err != nil {
				return
			}
		}
		for sti := range stmnts {
			if sti == len(stmnts)-1 {
				if stmntsl > 1 {
					if result, err = cn.ExecContext(ctx, stmnts[sti].query, stmnts[sti].args...); err != nil {
						cn.Close()
						return
					}
					return
				}
				result, err = db.ExecContext(ctx, stmnts[sti].query, stmnts[sti].args...)
				return
			}
			if result, err = cn.ExecContext(ctx, stmnts[sti].query, stmnts[sti].args...); err != nil {
				cn.Close()
				return
			}
		}
		return
	}

	if err = frmtqry.FprintQuery(bfqry, name, driver, query, a...); err == nil {
		frmtqry = nil
		goto noformat
	}
	return
}

type stmntentry struct {
	query string
	args  []any
}

func extractstmnts(bfqry snkio.BufferWriter, dvr snkdrvr.Driver, a ...any) (stmnts []*stmntentry) {
	var prssqlprmd, _ = dvr.(snkdrvr.DriverParseSql)
	al := len(a)
	var aiter func(func(any) bool) = func(f func(any) bool) {
		for al > 0 {
			if aitr, aitrk := a[0].(func(func(any) bool)); aitrk {
				a = a[1:]
				al--
				for ai := range aitr {
					if f(ai) {
						continue
					}
					return
				}
				continue
			}
			if f(a[0]) {
				a = a[1:]
				al--
				continue
			}
			return
		}
	}
	if prssqlprmd == nil {
		prssqlprmd = snkdrvr.DriverParseSqlFunc(func(name string, idx int) (string, string) {
			return "?", ""
		})
	}
	//bfw.Print(query)
	var arsg []any
	var ignrdagrs []any
	var argmp map[string]any
	var argrecord Record
	var argparams parameters.Parameters
	var argurlvals url.Values
	var argnmdargvals map[string]any
	for ai := range aiter {
		if ainmdarg, ainmdargk := ai.(sql.NamedArg); ainmdargk {
			if argnmdargvals == nil {
				argnmdargvals = map[string]any{}
			}
			argnmdargvals[ainmdarg.Name] = ainmdarg.Value
			continue
		}
		if aimp, aimpk := ai.(map[string]any); aimpk {
			ignrdagrs = append(ignrdagrs, ai)
			if len(aimp) > 0 {
				if argmp == nil {
					argmp = map[string]any{}
				}
				var name, oknme = aimp["Name"]
				var _, okordnl = aimp["Ordinal"]
				var val, okvl = aimp["Value"]
				if oknme && okordnl && okvl {
					if nme, _ := name.(string); nme != "" {
						argmp[nme] = val
					}
					continue
				}

				maps.Copy(argmp, aimp)
			}
			continue
		}
		if airec, aireck := ai.(Record); aireck {
			ignrdagrs = append(ignrdagrs, ai)
			if argrecord == nil {
				argrecord = airec
			}
			continue
		}
		if aiprms, aiprmsk := ai.(parameters.Parameters); aiprmsk {
			ignrdagrs = append(ignrdagrs, ai)
			if argparams == nil {
				argparams = aiprms
			}
			continue
		}
		if aiurlvals, aiurlvalsk := ai.(url.Values); aiurlvalsk {
			if argurlvals == nil {
				argurlvals = aiurlvals
			}
			continue
		}
		arsg = append(arsg, ai)
	}
	if !bfqry.Empty() {
		var chkargvals = bfqry.Contains("@")
		var mtchdargvals map[string]any
		var bfr snkio.BufferReader
		bfr, _ = bfqry.Reader()
		bfqry.Reset()
		defer bfr.Close()
		var stmntitr = snkio.IterRunes(nil, bfr)

		psr := rune(0)
		var nmerns []rune
		var crntargs []any
		captureArg := func(name string, val any) {
			prsdsql, unqnme := prssqlprmd.ParseSqlParam(name, len(crntargs))
			if name != "" && strings.Contains(prsdsql, "@"+name+"@") {
				crntargs = append(crntargs, sql.NamedArg{Name: name, Value: val})
				bfqry.WriteRunes([]rune(prsdsql)...)
				nmerns = nil
				return
			}
			if unqnme != "" {
				if mtchdargvals == nil {
					if _, nmdrg := val.(sql.NamedArg); !nmdrg {
						val = sql.Named(unqnme, val)
					}
					mtchdargvals = map[string]any{unqnme: val}
					crntargs = append(crntargs, val)
					bfqry.WriteRunes([]rune(prsdsql)...)
					nmerns = nil
					return
				}
				if mtcdv, ok := mtchdargvals[unqnme]; ok {
					if _, nmdrg := mtcdv.(sql.NamedArg); !nmdrg {
						mtcdv = sql.Named(unqnme, mtcdv)
						mtchdargvals[unqnme] = mtcdv
						crntargs[len(crntargs)-1] = mtcdv
					}
					bfqry.WriteRunes([]rune(prsdsql)...)
					nmerns = nil
					return
				}
				if _, nmdrg := val.(sql.NamedArg); !nmdrg {
					val = sql.Named(unqnme, val)
				}
				mtchdargvals[unqnme] = val
				crntargs = append(crntargs, val)
				bfqry.WriteRunes([]rune(prsdsql)...)
				nmerns = nil
				return
			}
			crntargs = append(crntargs, val)
			bfqry.WriteRunes([]rune(prsdsql)...)
			nmerns = nil
		}
	again:
		for sr := range stmntitr {
			if sr == '\'' {
				psr = sr
				bfqry.WriteRunes(sr)
				for sr = range stmntitr {
					bfqry.WriteRunes(sr)
					if sr == '\'' {
						if psr == sr {
							continue
						}
						goto again
					}
					psr = sr
				}
				continue
			}
			if sr == '@' && chkargvals {
				psr = 0
				for sr = range stmntitr {
					if sr == '@' {
						if tstname := string(nmerns); tstname != "" {
							if len(mtchdargvals) > 0 {
								if argv, ok := mtchdargvals[tstname]; ok {
									captureArg(tstname, argv)
									goto again
								}
							}
							if argparams != nil {
								if argparams.Exist(tstname) {
									if tstv := strings.Join(argparams.Get(tstname), ""); tstv != "" {
										captureArg(tstname, tstv)
										goto again
									}
									if len(argmp) > 0 {
										if argv, ok := argmp[tstname]; ok {
											captureArg(tstname, argv)
											goto again
										}
									}
									captureArg(tstname, "")
									goto again
								}
							}
							if argrecord != nil {
								if rv, rfnd := argrecord.Value(argrecord.Index(tstname)); rfnd {
									captureArg(tstname, rv)
									goto again
								}
							}
							if len(argurlvals) > 0 {
								if argv, ok := argurlvals[tstname]; ok {
									if len(argv) == 0 || argv[0] == "" {
										if len(argmp) > 0 {
											if argv, ok := argmp[tstname]; ok {
												captureArg(tstname, argv)
												goto again
											}
										}
										captureArg(tstname, "")
										goto again
									}
									captureArg(tstname, argv[0])
									goto again
								}
							}
							if len(argnmdargvals) > 0 {
								if argv, ok := argnmdargvals[tstname]; ok {
									captureArg(tstname, argv)
									goto again
								}
							}
							if len(argmp) > 0 {
								if argv, ok := argmp[tstname]; ok {
									captureArg(tstname, argv)
									goto again
								}
							}
							captureArg(tstname, nil)
							goto again
						}
						captureArg("", nil)
						goto again
					}
					nmerns = append(nmerns, sr)
				}
				bfqry.WriteRunes('@')
				bfqry.WriteRunes(nmerns...)
				nmerns = nil
				continue
			}
			bfqry.WriteRunes(sr)
			if sr == ';' {
				psr = 0
				stmnts = append(stmnts, &stmntentry{query: bfqry.String(), args: crntargs})
				bfqry.Reset()
				crntargs = nil
				continue
			}
			psr = sr
		}
		if !bfqry.Empty() {
			stmnt := bfqry.String()
			if stmnt[len(stmnt)-1] != ';' {
				stmnt += ";"
			}
			stmnts = append(stmnts, &stmntentry{query: bfqry.String(), args: crntargs})
			bfqry.Reset()
			crntargs = nil
		}
		return
	}
	return
}

var GlobalFormatQuery FormatQuery
