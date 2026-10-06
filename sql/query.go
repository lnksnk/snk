package sql

import (
	"context"
	"database/sql"
	"io"
	"sync"
)

type FormatQuery interface {
	FprintQuery(out io.Writer, name, driver, query string, a ...any) (err error)
}

type FormatQueryFunc func(out io.Writer, name, driver, query string, a ...any) (err error)

func (fmtqryfnc FormatQueryFunc) FprintQuery(out io.Writer, name, driver, query string, a ...any) (err error) {
	return fmtqryfnc(out, name, driver, query, a...)
}

func Query(name, query string, a ...any) (rows SqlRows, err error) {
	return QueryContext(context.Background(), name, query, a...)
}

func QueryContext(ctx context.Context, name, query string, a ...any) (rows SqlRows, err error) {
	if name != "" && query != "" {
		if cndef, _ := cnndefs.Load(name); cndef != nil {
			wg := &sync.WaitGroup{}
			wg.Go(func() {
				rows, err = cndef.QueryContext(ctx, query, a...)
			})
			wg.Wait()
		}
		return
	}
	return
}

func Exec(name, query string, a ...any) (resut sql.Result, err error) {
	return ExecContext(context.Background(), name, query, a...)
}

func ExecContext(ctx context.Context, name, query string, a ...any) (result sql.Result, err error) {
	if name != "" && query != "" {
		if cndef, _ := cnndefs.Load(name); cndef != nil {
			wg := &sync.WaitGroup{}
			wg.Go(func() {
				result, err = cndef.ExecContext(ctx, query, a...)
			})
			wg.Wait()
		}
		return
	}
	return
}

func NumberedRecords(rows SqlRows) func(func(Record, int64) bool) {
	return func(f func(Record, int64) bool) {
		for rec := range Records(rows) {
			if !f(rec, rec.Nr()) {
				return
			}
		}
	}
}

func Records(rows SqlRows) func(func(Record) bool) {
	return func(f func(Record) bool) {
		if rows == nil {
			return
		}
		defer rows.Close()
		var rec = &record{init: true}
		defer rec.Close()
		var cls, err = rows.Columns()
		if rec.err = err; rec.err == nil {
			rec.cls = cls
			//init
			if !f(rec) {
				return
			}
			rec.init = false
			var clsl = len(cls)
			rec.first = true
			if rows.Next() {
				if rec.err = rows.Err(); rec.err == nil {
					rec.dta = make([]any, clsl)
					var data = make([]any, clsl)
					var dataref = make([]any, clsl)
					for di := range data {
						dataref[di] = &data[di]
					}
				nxtrow:
					if rec.err = rows.Scan(dataref...); rec.err == nil {
						copy(rec.dta, data)
						rec.last, rec.err = !rows.Next(), rows.Err()
						if rec.err == nil {
							rec.nr++
							if !f(rec) {
								return
							}
							rec.first = false
							if rec.last {
								rec.finit = true
								f(rec)
								return
							}
							rec.first = false
							goto nxtrow
						}
						f(rec)
						return
					}
				}
				f(rec)
				return
			}
		}
	}
}

type Record interface {
	Nr() int64
	Columns() []string
	Data() []any
	First() bool
	Last() bool
	Init() bool
	Finit() bool
	Err() error
	Value(int, ...any) (any, bool)
	Index(string) int
}

type record struct {
	nr    int64
	first bool
	last  bool
	init  bool
	finit bool
	err   error
	cls   []string
	dta   []any
}

// Index implements [Record].
func (r *record) Index(col string) int {
	if cls, cl := r.cls, len(r.cls); cl > 0 && col != "" {
		for ci := range cls {
			if cls[ci] == col {
				return ci
			}
			if ci < cl-1 {
				if cls[cl-(ci+1)] == col {
					return cl - (ci + 1)
				}
			}
		}
		return -1
	}
	return -1
}

// Value implements [Record].
func (r *record) Value(idx int, defaultval ...any) (any, bool) {
	if dta := r.dta; idx > 0 && idx < len(dta) {
		return dta[idx], true
	}
	if len(defaultval) > 1 {
		return defaultval[0], false
	}
	return nil, false
}

// Columns implements [Record].
func (r *record) Columns() []string {
	return r.cls
}

// Data implements [Record].
func (r *record) Data() []any {
	return r.dta
}

// Err implements [Record].
func (r *record) Err() (err error) {
	if err = r.err; err == io.EOF {
		return
	}
	return
}

// Finit implements [Record].
func (r *record) Finit() bool {
	return r.finit
}

// First implements [Record].
func (r *record) First() bool {
	return r.first
}

// Init implements [Record].
func (r *record) Init() bool {
	return r.init
}

// Last implements [Record].
func (r *record) Last() bool {
	return r.last
}

// Nr implements [Record].
func (r *record) Nr() int64 {
	return r.nr
}

func (r *record) Close() (err error) {
	err = r.err
	r.cls, r.dta, r.err = nil, nil, nil
	return
}
