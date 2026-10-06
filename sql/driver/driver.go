package driver

import (
	"context"
	"database/sql/driver"

	snkio "github.com/lnksnk/snk/io"
)

type DriverParseSql interface {
	ParseSqlParam(name string, idx int) (stmntname, uniqname string)
}

type DriverParseSqlFunc func(name string, idx int) (stmntname, uniqname string)

func (dpsfnc DriverParseSqlFunc) ParseSqlParam(name string, idx int) (stmntname, uniqname string) {
	return dpsfnc(name, idx)
}

type Driver interface {
	driver.Driver
	driver.DriverContext
	DriverParseSql
}

type internaldriver struct {
	dvr    driver.Driver
	dvrctx driver.DriverContext
	cnctrs snkio.SyncMap[string, driver.Connector]
	DriverParseSql
}

// Open implements [Driver].
func (i *internaldriver) Open(name string) (conn driver.Conn, err error) {
	if cnctrs, dvrctx := i.cnctrs, i.dvrctx; cnctrs != nil && dvrctx != nil && name != "" {
		if cnctr, ok := cnctrs.Load(name); ok {
			return cnctr.Connect(context.Background())
		}
	}
	err = driver.ErrBadConn
	return
}

// OpenConnector implements [Driver].
func (i *internaldriver) OpenConnector(name string) (cnctr driver.Connector, err error) {
	if cnctrs, dvrctx := i.cnctrs, i.dvrctx; cnctrs != nil && dvrctx != nil && name != "" {
		if cnctr, _ = cnctrs.Load(name); cnctr != nil {
			return
		}
		if cnctr, err = dvrctx.OpenConnector(name); err == nil {
			cnctr = &connector{gocnctr: cnctr}
			cnctrs.Store(name, cnctr)
		}
		return
	}
	err = driver.ErrBadConn
	return
}

type DriverContextFunc func(name string) (driver.Connector, error)

func (dvrctxfnc DriverContextFunc) OpenConnector(name string) (driver.Connector, error) {
	return dvrctxfnc(name)
}

type GoConnectorContext interface {
	Connect(context.Context) (driver.Conn, error)
	Driver() driver.Driver
}

type goconnector struct {
	datasource string
	dvr        driver.Driver
}

func (gocntr *goconnector) Connect(ctx context.Context) (driver.Conn, error) {
	if dvr, datasource := gocntr.dvr, gocntr.datasource; dvr != nil && datasource != "" {
		return dvr.Open(datasource)
	}
	return nil, driver.ErrBadConn
}

func (gocntr *goconnector) Driver() driver.Driver {
	return gocntr.dvr
}

func GoConnector(datasource string, dvr driver.Driver) GoConnectorContext {
	return &goconnector{dvr: dvr}
}

func NewDriver(dvrctx driver.DriverContext, dvr driver.Driver, dvrparssql DriverParseSql) Driver {
	if dvr == nil && dvrctx != nil {
		dvr, _ = dvrctx.(driver.Driver)
	}
	if dvrparssql == nil {
		dvrparssql = DriverParseSqlFunc(func(name string, idx int) (string, string) {
			return "?", ""
		})
	}
	return &internaldriver{dvr: dvr, dvrctx: dvrctx, cnctrs: snkio.NewSyncMap[string, driver.Connector](snkio.MapDeletedFunc[string, driver.Connector](func(key string, value ...driver.Connector) {

	})), DriverParseSql: dvrparssql}
}
