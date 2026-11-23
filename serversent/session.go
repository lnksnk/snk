package serversent

import (
	"context"
	"net/http"
	"time"
)

type Session interface {
	Context() context.Context
	SessionID() string
	Trigger(id string, event string, a ...any) error
	Print(id string, event string, a ...any) error
	Close() error
	ProcessEvents() error
}

func unwrapResponseWriter(w http.ResponseWriter) (flush ssecon, dl interface {
	SetWriteDeadline(t time.Time) error
}) {
	currentWriter := w
	var fslhr http.Flusher
	for currentWriter != nil {
		// Check for http.Flusher
		if f, ok := currentWriter.(http.Flusher); ok {
			fslhr = f // Found a Flusher, should be overwritten by a deeper one
		}

		// Check for writeDeadliner
		if d, ok := currentWriter.(interface {
			SetWriteDeadline(t time.Time) error
		}); ok {
			dl = d // Found a Deadliner, should be overwritten by a deeper one
		}

		// Attempt to unwrap further
		unwrapper, ok := currentWriter.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			// Not an unwrapper, so this is as deep as we can go
			break
		}

		nextWriter := unwrapper.Unwrap()

		// If Unwrap() returns the same writer, break to prevent an infinite loop.
		// If nextWriter is nil, the loop condition (currentWriter != nil)
		// will handle termination when currentWriter is updated.
		if nextWriter == currentWriter {
			break
		}
		currentWriter = nextWriter
	}
	if fslhr != nil {
		flush, _ = currentWriter.(ssecon)
	}
	return flush, dl
}

func (s *session) ProcessEvents() (err error) {
	if ssnctx := s.ssnctx; ssnctx != nil {
		s.Trigger("", "", ssnctx.lstssnid)
		var t = time.Now().Add(15 * time.Second)
		var didevt = false
		for s.sscn != nil {
			for se := range ssnctx.NextEvents {
				if err = s.sendevt(se); err != nil {
					return
				}
				if !didevt {
					didevt = true
				}
			}
			if didevt {
				didevt = false
				t = t.Add(15 * time.Second)
			}
			if tn := time.Now(); tn.After(t) {
				if err = s.Print("", "", "ping"); err != nil {
					return
				}
				t = tn.Add(15 * time.Second)
				continue
			}
			time.Sleep(time.Millisecond * 20)
		}
		return
	}
	return
}

func (s *session) sendevt(se *sendEvent) (err error) {
	if buf := se.buf; buf != nil && !buf.Empty() {
		err = s.Print(se.id, se.event, buf)
		se.Close()
		return
	}
	err = s.Print(se.id, se.event)
	se.Close()
	return
}

type session struct {
	ctx    context.Context
	sscn   ssecon
	sseid  string
	ssnctx *SSEContext
}

// Trigger implements [Session].
func (s *session) Trigger(id string, event string, a ...any) error {
	if ssnctx := s.ssnctx; ssnctx != nil {
		return ssnctx.Enqueue(id, event, a...)
	}
	return nil
}

const maxdatalen int = 65536

// Print implements [Session].
func (s *session) Print(id string, event string, d ...any) (err error) {
	if sscn, ssnctx := s.sscn, s.ssnctx; sscn != nil && ssnctx != nil {
		if id == "" {
			id, _ = ssnctx.NextEventID()
		}
		if err = SseFPrint(sscn, id, event, d...); err == nil {
			sscn.Flush()
			return
		}
	}
	return
}

// Context implements [Session].
func (s *session) Context() context.Context {
	if ctx := s.ctx; ctx != nil {
		return ctx
	}
	return context.Background()
}

// Close implements [Session].
func (s *session) Close() (err error) {
	ssnctx := s.ssnctx
	s.sscn, s.ssnctx = nil, nil
	if ssnctx != nil {
		//ssnctx.accessedat = time.Now()
		if sseid := ssnctx.lstssnid; sseid != "" {
			GlobalSSContext.CompareAndDelete(sseid, ssnctx)
		}
	}
	return
}

// SessionID implements [Session].
func (s *session) SessionID() string {
	return s.sseid
}
