package sqlite

import (
	"database/sql/driver"
	"strconv"
	"uuid"

	snksql "github.com/lnksnk/snk/sql"
	snksqldrv "github.com/lnksnk/snk/sql/driver"

	"modernc.org/sqlite"
)

var sqltdvr = &sqlite.Driver{}

var SQLiteDriver = snksqldrv.NewDriver(snksqldrv.DriverContextFunc(func(datasource string) (driver.Connector, error) {
	if datasource == "" || datasource == ":memory:" {
		datasource = "file::memory:?mode=memory"
	}
	return sqlite.NewConnector(datasource) //snksqldrv.GoConnector(datasource, sqltdvr), nil
}), sqltdvr, snksqldrv.DriverParseSqlFunc(func(name string, idx int) (string, string) {
	if name == "" {
		name = strconv.Itoa(idx + 1)
	}
	return "?", name
}))

func init() {
	sqlite.RegisterFunction("uuid", &sqlite.FunctionImpl{
		NArgs:         0,    // Accept 0 arguments for uuid()
		Deterministic: true, // Optimizes execution when evaluated within query contexts
		Scalar: func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
			return uuid.NewV7().String(), nil
		},
	})
	snksql.Register("sqlite", SQLiteDriver, nil)
}
