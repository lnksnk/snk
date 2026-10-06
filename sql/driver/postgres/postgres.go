package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	snksql "github.com/lnksnk/snk/sql"
	snksqldvr "github.com/lnksnk/snk/sql/driver"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

func ParseSqlParam(name string, idx int) (string, string) {
	return fmt.Sprintf("$%d", idx+1), "A" + strconv.Itoa(idx+1)
}

func Connector(datasource string) (cnctr driver.Connector, err error) {
	if !strings.Contains(datasource, "pool_max_conn_lifetime=") {
		datasource += " pool_max_conn_lifetime=10s pool_health_check_period=20s"
	}
	if !strings.Contains(datasource, "pool_health_check_period=") {
		datasource += " pool_health_check_period=20s"
	}
	if !strings.Contains(datasource, "sslmode=") {
		datasource += " sslmode=disable"
	}
	pxcnfg, pxerr := pgxpool.ParseConfig(datasource)
	if pxerr != nil {
		err = pxerr
		return
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, pxcnfg)
	if err != nil {
		return nil, errors.Wrap(err, "create db conn pool")
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, "create db conn pool")
	}
	return stdlib.GetPoolConnector(pool), err
}

var PooledDriver = snksqldvr.NewDriver(snksqldvr.DriverContextFunc(Connector), nil, snksqldvr.DriverParseSqlFunc(ParseSqlParam))

func init() {
	snksql.Register("postgres", PooledDriver, nil)
}
