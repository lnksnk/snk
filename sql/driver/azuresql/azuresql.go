package azuresql

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/microsoft/go-mssqldb/azuread"

	snksql "github.com/lnksnk/snk/sql"
	snksqldvr "github.com/lnksnk/snk/sql/driver"
)

func ParseSqlParam(name string, idx int) (stmntname, uniqname string) {
	/*stmntname = fmt.Sprintf("@%d", idx+1)
	if uniqname = name; uniqname == "" {
		uniqname = strconv.Itoa(idx + 1)
	}*/
	return fmt.Sprintf("@A%d", idx+1), fmt.Sprintf("A%d", idx+1)
}

var azuredvr = &azuread.Driver{}

var AZUREDriver = snksqldvr.NewDriver(snksqldvr.DriverContextFunc(func(datasource string) (driver.Connector, error) {
	var tlsversion = ""
	for _, dtasrc := range strings.Split(datasource, ";") {
		if strings.HasPrefix(dtasrc, "tlsmin=") {
			if tlsversion = strings.TrimSpace(dtasrc[len("tlsmin="):]); tlsversion == "" {
				tlsversion = "1.0"
				datasource = strings.Replace(datasource, "tlsmin=", "tlsmin="+tlsversion, 1)
			}
		}
	}
	if tlsversion == "" {
		tlsversion = "1.0"
		datasource += ";" + "tlsmin=" + tlsversion
	}
	return snksqldvr.GoConnector(datasource, azuredvr), nil
}), azuredvr, snksqldvr.DriverParseSqlFunc(ParseSqlParam))

func init() {
	snksql.Register("azuresql", AZUREDriver, nil)
}
