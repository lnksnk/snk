package serve

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lnksnk/snk/mime"
	"github.com/lnksnk/snk/parameters"
	"github.com/lnksnk/snk/serversent"
	"github.com/lnksnk/snk/websocket"
	"github.com/medama-io/go-useragent"
)

type ServerWS interface {
	ServeWS(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeWSFunc func(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (sws ServeWSFunc) ServeWS(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return sws(path, params, ua, w, r)
}

type ServeSSE interface {
	ServeSSE(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeSSEFunc func(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (ssse ServeSSEFunc) ServeSSE(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return ssse(path, params, ua, w, r)
}

type ServeUTF interface {
	ServeUTF(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeUTFFunc func(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (sutf ServeUTFFunc) ServeUTF(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return sutf(path, mimeinfo, params, ua, w, r)
}

type ServeSTD interface {
	ServeSTD(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeSTDFunc func(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (sstd ServeSTDFunc) ServeSTD(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return sstd(path, mimeinfo, params, ua, w, r)
}

type ServeMEDIA interface {
	ServeMEDIA(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeMediaFunc func(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (smd ServeMediaFunc) ServeMEDIA(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return smd(path, mimeinfo, params, ua, w, r)
}

type UTF8Ext string

func ServeHttp(a ...any) (hndlr http.Handler, err error) {
	var srvws ServerWS
	var srvsse ServeSSE
	var srvstd ServeSTD
	var srvmdia ServeMEDIA
	var srvutf ServeUTF
	var utf8exts []string
	for ai := range a {
		if srvwsd, srvwsk := a[ai].(ServeWSFunc); srvwsk {
			if srvws == nil {
				srvws = srvwsd
				continue
			}
			continue
		}
		if srvssed, srvssek := a[ai].(ServeSSEFunc); srvssek {
			if srvsse == nil {
				srvsse = srvssed
				continue
			}
			continue
		}
		if srvmdiad, srvmdiak := a[ai].(ServeMediaFunc); srvmdiak {
			if srvmdia == nil {
				srvmdia = srvmdiad
				continue
			}
			continue
		}
		if srvstdd, srvstdk := a[ai].(ServeSTDFunc); srvstdk {
			if srvstd == nil {
				srvstd = srvstdd
				continue
			}
			continue
		}
		if srvutfd, srvutfk := a[ai].(ServeUTFFunc); srvutfk {
			if srvutf == nil {
				srvutf = srvutfd
				continue
			}
			continue
		}
		if srvwsd, srvwsk := a[ai].(ServerWS); srvwsk {
			if srvws == nil {
				srvws = srvwsd
				continue
			}
			continue
		}
		if srvssed, srvssek := a[ai].(ServeSSE); srvssek {
			if srvsse == nil {
				srvsse = srvssed
				continue
			}
			continue
		}
		if srvmdiad, srvmdiak := a[ai].(ServeMEDIA); srvmdiak {
			if srvmdia == nil {
				srvmdia = srvmdiad
				continue
			}
			continue
		}
		if srvstdd, srvstdk := a[ai].(ServeSTD); srvstdk {
			if srvstd == nil {
				srvstd = srvstdd
				continue
			}
			continue
		}
		if srvutfd, srvutfk := a[ai].(ServeUTF); srvutfk {
			if srvutf == nil {
				srvutf = srvutfd
				continue
			}
			continue
		}
		if utf8extd, utf8extk := a[ai].(UTF8Ext); utf8extk {
			utf8exts = append(utf8exts, string(utf8extd))
			continue
		}
	}
	if srvsse == nil {
		srvsse = ServeSSEFunc(func(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return fmt.Errorf("sent events not supported")
		})
	}
	if srvws == nil {
		srvws = ServeWSFunc(func(path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return fmt.Errorf("websockets not supported")
		})
	}

	if srvmdia == nil {
		srvmdia = ServeMediaFunc(func(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return fmt.Errorf("serving media not supported")
		})
	}

	if srvstd == nil {
		srvstd = ServeSTDFunc(func(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return
		})
	}
	if srvutf == nil {
		srvutf = ServeUTFFunc(func(path string, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return
		})
	}

	var UTF8Exts map[string]bool
	if utfl := len(utf8exts); utfl > 0 {
		UTF8Exts = map[string]bool{}
		utfi := 0
		for utfi < utfl {
			if uext := filepath.Ext(utf8exts[utfi]); uext != "" && !UTF8Exts[uext] {
				UTF8Exts[uext] = true
				utfi++
				continue
			}
			utf8exts = append(utf8exts[:utfi], utf8exts[utfi+1:]...)
			utfl--
		}
	}
	return &HttpServer{utf8exts: utf8exts, UTF8Exts: UTF8Exts, ServerWS: srvws, ServeSSE: srvsse, ServeUTF: srvutf, ServeSTD: srvstd, ServeMEDIA: srvmdia}, nil
}

type HttpServer struct {
	UTF8Exts map[string]bool
	utf8exts []string
	ServerWS
	ServeSSE
	ServeUTF
	ServeSTD
	ServeMEDIA
}

func (hs *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	var mimeinfo = mime.ExtMimeInfo(path)
	if srvws, srvsse, utf8exts, UTF8Exts := hs.ServerWS, hs.ServeSSE, hs.utf8exts, hs.UTF8Exts; srvsse != nil && srvws != nil {

		var served, err = false, error(nil)
		defer func() {
			if x := recover(); x != nil {
				if xs, xsk := x.(string); xsk {
					err = fmt.Errorf("%s", xs)
					return
				}
				err, _ = x.(error)
			}
			if err != nil {
				if strings.Contains(err.Error(), "no such file or directory") {
					err = nil
					return
				}
			}
			if flsr, _ := w.(http.Flusher); flsr != nil {
				flsr.Flush()
			}
		}()
		var ua = mainua.Parse(r.UserAgent())
		if ua.IsBot() {
			w.Header().Set("Content-Length", "0")
			return
		}
		if served, err = serveWsSse(path, ua, srvws, srvsse, w, r); served {
			return
		}
		var params, _ = parameters.HTTPRequestParameters(r)
		defer params.ClearAll()
		w.Header().Set("Content-Type", mimeinfo.Mimetype())
		if srvstd, srvutf, srvmdia := hs.ServeSTD, hs.ServeUTF, hs.ServeMEDIA; srvstd != nil && srvutf != nil && srvmdia != nil {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			if mimeinfo.Media() {
				err = srvmdia.ServeMEDIA(path, mimeinfo, params, ua, w, r)
				return
			}
			if ext := filepath.Ext(path); ext != "" {
				for ufi := range utf8exts {
					if ext == utf8exts[ufi] {
						goto serveutf
					}
				}
				if UTF8Exts[ext] {
					goto serveutf
				}
			}
			err = srvstd.ServeSTD(path, mimeinfo, params, ua, w, r)
			return
		serveutf:
			w.Header().Set("Content-Type", mimeinfo.Mimetype())
			err = srvutf.ServeUTF(path, mimeinfo, params, ua, w, r)
			return
		}
	}

	w.Header().Set("Content-Type", mimeinfo[1])
	w.Header().Set("Content-Length", "0")
}

func serveWsSse(path string, ua useragent.UserAgent, srvws ServerWS, srvsse ServeSSE, w http.ResponseWriter, r *http.Request) (served bool, err error) {
	if served, err = websocket.UpgradeAndServe(w, r, func(w http.ResponseWriter, r *http.Request) error {
		var params = parameters.NewParameters()
		defer params.ClearAll()
		parameters.LoadParametersFromUrlValues(params, r.URL.Query())
		return srvws.ServeWS(path, params, ua, w, r)
	}); served {
		return
	}
	served, err = serversent.UpgradeAndServe(w, r, func(w http.ResponseWriter, r *http.Request) error {
		var params = parameters.NewParameters()
		defer params.ClearAll()
		parameters.LoadParametersFromUrlValues(params, r.URL.Query())
		return srvsse.ServeSSE(path, params, ua, w, r)
	})
	return
}

var mainua = useragent.NewParser()
