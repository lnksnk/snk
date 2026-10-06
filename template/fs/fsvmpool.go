package fs

import snkio "github.com/lnksnk/snk/io"

func VMPool[R any](newvm snkio.PoolNewFunc[R], cleanupvm snkio.PoolCleanupFunc[R]) snkio.Pool[R] {
	return snkio.NewPool(newvm, cleanupvm)
}
