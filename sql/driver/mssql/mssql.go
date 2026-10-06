package mssql

import (
	"database/sql/driver"
	"fmt"
	"strings"

	snksql "github.com/lnksnk/snk/sql"
	snksqldvr "github.com/lnksnk/snk/sql/driver"

	mssql "github.com/microsoft/go-mssqldb"
)

func ParseSqlParam(name string, idx int) (stmntname, uniqname string) {
	/*stmntname = fmt.Sprintf("@%d", idx+1)
	if uniqname = name; uniqname == "" {
		uniqname = strconv.Itoa(idx + 1)
	}*/
	return fmt.Sprintf("@A%d", idx+1), fmt.Sprintf("A%d", idx+1)
}

var mssqldrvr = &mssql.Driver{}

var MSSqlDriver = snksqldvr.NewDriver(snksqldvr.DriverContextFunc(func(datasource string) (cnctr driver.Connector, err error) {
	var tlsversion = ""
	var multiSubnetFailover = ""
	for _, dtasrc := range strings.Split(datasource, ";") {
		if strings.HasPrefix(dtasrc, "tlsmin=") {
			if tlsversion = strings.TrimSpace(dtasrc[len("tlsmin="):]); tlsversion == "" {
				tlsversion = "1.0"
				datasource = strings.Replace(datasource, "tlsmin=", "tlsmin="+tlsversion, 1)
			}
			continue
		}
		if strings.HasPrefix(dtasrc, "multiSubnetFailover=") {
			multiSubnetFailover = dtasrc
		}
	}
	if tlsversion == "" {
		tlsversion = "1.0"
		datasource += ";" + "tlsmin=" + tlsversion
	}
	if multiSubnetFailover == "" {
		datasource += ";" + "multiSubnetFailover=true"
	}
	var mssqlcnctr *mssql.Connector
	if mssqlcnctr, err = mssql.NewConnector(datasource); err == nil {
		mssqlcnctr.SessionInitSQL = "select 1"
		cnctr = mssqlcnctr
	}
	return
}), mssqldrvr, snksqldvr.DriverParseSqlFunc(ParseSqlParam))

func init() {
	snksql.Register("mssql", MSSqlDriver, nil)
	snksql.Register("sqlserver", MSSqlDriver, nil)
}
