package sql

import (
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"strings"
	"sync"

	snkio "github.com/lnksnk/snk/io"
	snksqldvr "github.com/lnksnk/snk/sql/driver"
)

var drivers = snkio.NewSyncMap[string, snksqldvr.Driver](snkio.MapDeletedFunc[string, snksqldvr.Driver](func(key string, value ...snksqldvr.Driver) {

}))

var cnndefs = snkio.NewSyncMap[string, ConnDefinition](snkio.MapDeletedFunc[string, ConnDefinition](func(key string, value ...ConnDefinition) {
	for vi := range value {
		value[vi].Close()
	}
}))

func Register(name string, driver sqldriver.Driver, dvrparsesql snksqldvr.DriverParseSql) (err error) {
	if name == "" {
		return fmt.Errorf("driver name empty")
	}
	var snkdvr, snkexist = drivers.Load(name)
	if snkexist {
		return
	}
	if driver == nil {
		sql.Register(name, driver)
		var snkdvr = snksqldvr.NewDriver(nil, driver, dvrparsesql)
		drivers.Store(name, snkdvr)
	}
	if snkdvr, _ = driver.(snksqldvr.Driver); snkdvr != nil {
		drivers.Store(name, snkdvr)
	}
	return
}

func DefineConn(name, datasource string) (conndef ConnDefinition, err error) {
	if name == "" || datasource == "" {
		err = fmt.Errorf("No definition name or datasource provided")
		return
	}

	var driver = ""
	var di, dnmi = strings.Index(name, "::"), strings.Index(name, "-")
	if di == -1 && dnmi == -1 {
		err = fmt.Errorf("No definition driver provided")
		return
	}
	if di > -1 {
		if driver, name = name[di+2:], name[:di]; driver == "" || name == "" {
			err = fmt.Errorf("No definition name or driver provided")
			return
		}
	}
	if dnmi > -1 {
		if driver = name[dnmi+1:]; driver == "" {
			err = fmt.Errorf("No definition driver provided")
			return
		}
	}
	var snkcndf, snkok = cnndefs.Load(name)
	var cndf *conndefin
	if snkok {
		if cndf, snkok = snkcndf.(*conndefin); snkok {
			if cndf.datasource == datasource && cndf.driver == driver {
				return
			}
			cnndefs.Delete(name)
		}
	}
	var snkdvr snksqldvr.Driver
	if snkdvr, snkok = drivers.Load(driver); snkok {
		var cnctr sqldriver.Connector
		if cnctr, err = snkdvr.OpenConnector(datasource); err == nil {
			snkcndf = &conndefin{name: name, driver: driver, dvr: snkdvr, datasource: datasource, db: sql.OpenDB(cnctr)}
			cnndefs.Store(name, snkcndf)
			return
		}
		return
	}
	err = fmt.Errorf("no registered driver %s found", driver)
	return
}

func ConnStats(name string) (stats map[string]any) {
	if snkcndf, snkok := cnndefs.Load(name); snkok {
		wg := &sync.WaitGroup{}
		wg.Go(func() {
			var cnstats = snkcndf.Stats()
			stats = map[string]any{}
			stats["idle"] = cnstats.Idle
			stats["inuse"] = cnstats.InUse
			stats["open"] = cnstats.OpenConnections
			stats["max-open"] = cnstats.MaxOpenConnections
			stats["max-idle-closed"] = cnstats.MaxIdleClosed
			stats["max-idle-time-closed"] = cnstats.MaxIdleTimeClosed
			stats["max-lifetime-closed"] = cnstats.MaxLifetimeClosed
			stats["wait-count"] = cnstats.WaitCount
			stats["wait-new"] = cnstats.WaitDuration
		})
		wg.Wait()
	}
	return
}

func ConnDefinitions() (condefs []string) {
	cnndefs.Range(func(key string, value ConnDefinition) bool {
		condefs = append(condefs, key)
		return true
	})
	return
}

func CloseConnDef(name string) (err error) {
	if name == "" {
		return
	}
	if conndef, connloaded := cnndefs.Load(name); connloaded {
		err = conndef.Close()
		cnndefs.Delete(name)
	}
	return
}
