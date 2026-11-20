package serversent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	snkio "github.com/lnksnk/snk/io"
)

type ssecon interface {
	io.Writer
	http.Flusher
}

type ServeSSEContextFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, ssn Session) error

func UpgradeAndServeContext(ctx context.Context, w http.ResponseWriter, r *http.Request, srvsse ServeSSEContextFunc) (served bool, err error) {
	accepts := r.Header.Get("Accept")
	if served = accepts != "" && strings.Contains(accepts, "text/event-stream"); served {
		var lastEventID = ""
		if hdr := r.Header; hdr != nil {
			if h := hdr["Last-Event-Id"]; len(h) > 0 && h[0] != "" {
				lastEventID = h[0]
			}
		}
		if url := r.URL; url != nil && lastEventID == "" {
			if urlvals := url.Query(); urlvals != nil {
				lastEventID = urlvals.Get("lastEventId")
			}
		}

		var sseid string
		if strings.Contains(lastEventID, ":") {
			sseid = lastEventID[:strings.Index(lastEventID, ":")]
		}
		var ssectx = InvokeRetrieveSSEContext(sseid)
		if ssectx != nil {
			sseid = ssectx.lstssnid
			if lastEventID == "" {
				lastEventID, _ = ssectx.NextEventID()
			}
		}
		// Discover flusher & deadline.

		if sscn, err := sseupgrade(w); err == nil {
			var ssn Session = &session{
				ctx:    ctx,
				sscn:   sscn,
				sseid:  sseid,
				ssnctx: ssectx,
			}
			defer func() { _ = ssn.Close() }()
			var ctxend, ctxcnl = context.WithCancelCause(ssn.Context())
			go func() {
				var sseerr error = context.Canceled
				defer ctxcnl(sseerr)
				sseerr = ssn.ProcessEvents()
			}()
			<-ctxend.Done()
			if ctxerr := ctxend.Err(); ctxerr != nil && ctxerr != context.Canceled {
				err = ctxerr
			}
			return served, err
		}
	}
	return
}

func sseupgrade(w http.ResponseWriter) (sscn ssecon, err error) {
	if sscn, _ = unwrapResponseWriter(w); sscn == nil {
		return nil, fmt.Errorf("sse: ResponseWriter lacks http.Flusher")
	}
	// Mandatory headers.
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	w.WriteHeader(http.StatusOK)
	return
}

type UpgradeSessionFunc func(w http.ResponseWriter, r *http.Request) (Session, error)

var GlobalUpgradeSSE snkio.Map[string, UpgradeSessionFunc]

func init() {
	GlobalUpgradeSSE = snkio.NewMap[string, UpgradeSessionFunc]()
}
