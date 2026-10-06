package mysql

import (
	"strconv"

	snksql "github.com/lnksnk/snk/sql"
	snksqldvr "github.com/lnksnk/snk/sql/driver"

	"github.com/go-sql-driver/mysql"
)

func ParseSqlParam(name string, idx int) (string, string) {
	if name == "" {
		name = strconv.Itoa(idx + 1)
	}
	return "?", name
}

var mysqldvr = &mysql.MySQLDriver{}

var MySQLDriver = snksqldvr.NewDriver(snksqldvr.DriverContextFunc(mysqldvr.OpenConnector), mysqldvr, snksqldvr.DriverParseSqlFunc(ParseSqlParam))

func init() {
	snksql.Register("mysql", MySQLDriver, nil)
}
