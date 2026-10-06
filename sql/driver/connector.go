package driver

import (
	"context"
	"database/sql/driver"
)

type Connector interface {
	driver.Connector
}

type connector struct {
	gocnctr driver.Connector
	dvr     Driver
}

// Connect implements [driver.Connector].
func (c *connector) Connect(ctx context.Context) (conn driver.Conn, err error) {
	if gocnctr := c.gocnctr; gocnctr != nil {
		if conn, err = gocnctr.Connect(ctx); err == nil {
			conn = &internalconn{goconn: conn}
			return
		}
		return
	}
	err = driver.ErrBadConn
	return
}

// Driver implements [driver.Connector].
func (c *connector) Driver() driver.Driver {
	return c.dvr
}
