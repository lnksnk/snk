package serversent

import (
	"time"

	snkio "github.com/lnksnk/snk/io"
)

type sendEvent struct {
	event    string
	buf      snkio.BufferWriter
	id       string
	createat time.Time
}

func (se *sendEvent) Close() (err error) {
	buf := se.buf
	se.buf = nil
	if buf != nil {
		buf.Close()
	}
	return
}

func nextEvent(id string, event string, a ...any) (se *sendEvent, err error) {
	var buf snkio.BufferWriter
	if len(a) > 0 {
		if buf, err = snkio.WriterBuffer(a...); err != nil {
			return
		}
		se = &sendEvent{event: event, id: id, buf: buf, createat: time.Now()}
		return
	}
	buf, _ = snkio.WriterBuffer()
	se = &sendEvent{event: event, buf: buf, createat: time.Now()}
	return
}
