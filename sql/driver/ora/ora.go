package ora

import (
	"database/sql/driver"
	"fmt"
	"net/url"

	snksql "github.com/lnksnk/snk/sql"
	snksqldvr "github.com/lnksnk/snk/sql/driver"

	goora "github.com/sijms/go-ora/v3"
)

func ParseSqlParam(name string, idx int) (string, string) {
	return fmt.Sprintf(":%d", idx+1), ""
}

var oradvr = goora.NewDriver()

var OracleDriver = snksqldvr.NewDriver(snksqldvr.DriverContextFunc(func(datasource string) (cnctr driver.Connector, err error) {
	if _, err = url.ParseRequestURI(datasource); err != nil {
		return
	}
	return oradvr.OpenConnector(datasource)
}), oradvr, snksqldvr.DriverParseSqlFunc(ParseSqlParam))

func init() {
	snksql.Register("ora", OracleDriver, nil)
	snksql.Register("oracle", OracleDriver, nil)
}
