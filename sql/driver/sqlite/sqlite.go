package sqlite

import (
	"database/sql/driver"
	"strconv"

	snksql "github.com/lnksnk/snk/sql"
	snksqldrv "github.com/lnksnk/snk/sql/driver"

	"modernc.org/sqlite"
)

var sqltdvr = &sqlite.Driver{}

var SQLiteDriver = snksqldrv.NewDriver(snksqldrv.DriverContextFunc(func(datasource string) (driver.Connector, error) {
	if datasource == "" || datasource == ":memory:" {
		datasource = "file::memory:?mode=memory"
	}
	return snksqldrv.GoConnector(datasource, sqltdvr), nil
}), sqltdvr, snksqldrv.DriverParseSqlFunc(func(name string, idx int) (string, string) {
	if name == "" {
		name = strconv.Itoa(idx + 1)
	}
	return "?", name
}))

func init() {
	snksql.Register("sqlite", SQLiteDriver, nil)
}
