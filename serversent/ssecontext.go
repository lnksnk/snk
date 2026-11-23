package serversent

import (
	"sync"
	"time"
	"uuid"

	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/shortning"
)

type SSEContext struct {
	lastevtid  string
	lstssnid   string
	accessedat time.Time
	mtx        *sync.RWMutex
	events     []*sendEvent
}

func (s *SSEContext) Enqueue(id string, event string, a ...any) (err error) {
	if mtx := s.mtx; mtx != nil {
		var se *sendEvent
		if se, err = nextEvent(id, event, a...); err == nil {
			mtx.Lock()
			defer mtx.Unlock()
			s.events = append(s.events, se)
			return
		}
		return
	}
	return
}

func (s *SSEContext) LastEventID() (eventid string, gotit bool) {
	if mtx := s.mtx; mtx != nil {
		mtx.RLock()
		defer mtx.RUnlock()
		eventid = s.lastevtid
	}
	return eventid, eventid != ""
}

func (s *SSEContext) NextEventID() (eventid string, gotit bool) {
	if mtx := s.mtx; mtx != nil {
		mtx.Lock()
		defer mtx.Unlock()
		eventid = s.lstssnid + ":" + shortning.Encode(shortning.NextUnique())
		s.lastevtid = eventid
	}
	return eventid, eventid != ""
}

func (s *SSEContext) NextEvents(ssef func(se *sendEvent) bool) {
	var events []*sendEvent
	if mtx := s.mtx; mtx != nil {
		mtx.RLock()
		if events = s.events; len(events) == 0 {
			s.mtx.RUnlock()
			return
		}
		mtx.RUnlock()
		mtx.Lock()
		s.events = nil
		mtx.Unlock()
	}

	sei, sel := 0, len(events)
	for sei < sel {
		se := events[sei]
		sei++
		events = events[sei:]
		if !ssef(se) {
			se.Close()
			for sei < sel {
				se = events[sei]
				se.Close()
				sei++
				events = events[sei:]
			}
			return
		}
		se.Close()
	}
}

func (s *SSEContext) Close() {
	if mtx, events := s.mtx, s.events; mtx != nil {
		mtx.Lock()
		defer mtx.Unlock()
		s.lstssnid, s.lastevtid, s.events, s.mtx = "", "", nil, nil
		for ei := range events {
			events[ei].Close()
		}
	}
}

func InvokeRetrieveSSEContext(lastssnid string) (ssectx *SSEContext) {
	if lastssnid != "" {
		if ssectx, _ = GlobalSSContext.Load(lastssnid); ssectx != nil {
			return
		}
		ssectx = &SSEContext{lstssnid: lastssnid, mtx: &sync.RWMutex{}}
		GlobalSSContext.Store(lastssnid, ssectx)
		return
	}
	lastssnid = uuid.NewV7().String()
	ssectx = &SSEContext{lstssnid: lastssnid, mtx: &sync.RWMutex{}}
	GlobalSSContext.Store(lastssnid, ssectx)
	return
}

var GlobalSSContext = snkio.NewMap[string, *SSEContext](snkio.MapDeletedFunc[string, *SSEContext](func(ssectxkey string, a ...*SSEContext) {
	for ai := range a {
		a[ai].Close()
	}
}))

func PublishSSEContextEvent(sseid string, id, event string, a ...any) (err error) {
	if sseid != "" {
		if ssectx, ssestck := GlobalSSContext.Load(sseid); ssestck {
			err = ssectx.Enqueue(id, event, a...)
			return
		}
	}
	return
}
