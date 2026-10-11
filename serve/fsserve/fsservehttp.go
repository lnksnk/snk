package fsserve

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lnksnk/snk/fs"
	"github.com/lnksnk/snk/mime"
	"github.com/lnksnk/snk/parameters"
	"github.com/lnksnk/snk/serve"
	"github.com/lnksnk/snk/serversent"
	"github.com/lnksnk/snk/websocket"
	"github.com/medama-io/go-useragent"
)

type ServerWS interface {
	ServeWS(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeWSFunc func(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (sws ServeWSFunc) ServeWS(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return sws(fsstat, fsopen, path, params, ua, w, r)
}

type ServeSSE interface {
	ServeSSE(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeSSEFunc func(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (ssse ServeSSEFunc) ServeSSE(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return ssse(fsstat, fsopen, path, params, ua, w, r)
}

type ServeUTF interface {
	ServeUTF(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeUTFFunc func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, imeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (sutf ServeUTFFunc) ServeUTF(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return sutf(fi, fsstat, fsopen, mimeinfo, params, ua, w, r)
}

type ServeSTD interface {
	ServeSTD(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeSTDFunc func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (sstd ServeSTDFunc) ServeSTD(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return sstd(fi, fsstat, fsopen, mimeinfo, params, ua, w, r)
}

type ServeMEDIA interface {
	ServeMEDIA(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)
}

type ServeMediaFunc func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error)

func (smd ServeMediaFunc) ServeMEDIA(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
	return smd(fi, fsstat, fsopen, mimeinfo, params, ua, w, r)
}

type UTF8Ext string

func ServeHttp(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, pathnotfound func(string), a ...any) (hndlr http.Handler, err error) {
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
	if srvmdia == nil {
		srvmdia = ServeMediaFunc(func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			var f fs.File
			if f, err = fsopen.Open(fi.Path()); err == nil {
				defer f.Close()
				var rskrclsr, _ = f.(io.ReadSeekCloser)
				serve.ServeHttpRange(w, r, rskrclsr, fi.Size())
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", 0))
			return

		})
	}
	if srvsse == nil {
		srvsse = ServeSSEFunc(func(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return fmt.Errorf("sent events not supported")
		})
	}
	if srvstd == nil {
		srvstd = ServeSTDFunc(func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return
		})
	}
	if srvutf == nil {
		srvutf = ServeUTFFunc(func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return
		})
	}
	if srvws == nil {
		srvws = ServeWSFunc(func(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
			return fmt.Errorf("websockets not supported")
		})
	}
	if pathnotfound == nil {
		pathnotfound = func(s string) {}
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
	return &HttpServer{utf8exts: utf8exts, UTF8Exts: UTF8Exts, pathnotfound: pathnotfound, fsopen: fsopen, fsstat: fsstat, ServerWS: srvws, ServeSSE: srvsse, ServeUTF: srvutf, ServeSTD: srvstd, ServeMEDIA: srvmdia}, nil
}

type HttpServer struct {
	UTF8Exts     map[string]bool
	utf8exts     []string
	pathnotfound func(string)
	fsstat       fs.StatFileSystem
	fsopen       fs.OpenFileSystem
	ServerWS
	ServeSSE
	ServeUTF
	ServeSTD
	ServeMEDIA
}

func (hs *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	var mimeinfo = mime.ExtMimeInfo(path)
	if fsstat, fsopen, srvws, srvsse, pathnotfound, utf8exts, UTF8Exts := hs.fsstat, hs.fsopen, hs.ServerWS, hs.ServeSSE, hs.pathnotfound, hs.utf8exts, hs.UTF8Exts; fsstat != nil && fsopen != nil {

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
		if served, err = serveWsSse(fsstat, fsopen, path, ua, srvws, srvsse, w, r); served {
			return
		}
		var params, _ = parameters.HTTPRequestParameters(r)
		defer params.ClearAll()
		w.Header().Set("Content-Type", mimeinfo.Mimetype())
		srvstd, srvutf, srvmdia := hs.ServeSTD, hs.ServeUTF, hs.ServeMEDIA
		var fi fs.FileInfo
		if fi, err = fsstat.Stat(path); fi != nil {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			if !fi.IsDir() {
				mimeinfo = mime.ExtMimeInfo(fi.Path())
				w.Header().Set("Content-Type", mimeinfo[1])
				if mimeinfo.Media() {
					err = srvmdia.ServeMEDIA(fi, fsstat, fsopen, mimeinfo, params, ua, w, r)
					return
				}
				if UTF8Exts[filepath.Ext(fi.Path())] {
					goto serveutf
				}
				err = srvstd.ServeSTD(fi, fsstat, fsopen, mimeinfo, params, ua, w, r)
				return
			}
			if fi.IsDir() {
				for _, ext := range utf8exts {
					if fi, err = fsstat.Stat(path + "index" + ext); fi == nil {
						pathnotfound(path + "index" + ext)
						continue
					}
					path += "index" + ext
					goto serveutf
				}
				w.Header().Set("Content-Type", mimeinfo.Mimetype())
				w.Header().Set("Content-Length", "0")
				return
			}
			mimeinfo = mime.ExtMimeInfo(fi.Name())
			w.Header().Set("Content-Type", mimeinfo.Mimetype())
			w.Header().Set("Content-Length", "0")
			return
		serveutf:
			mimeinfo = mime.ExtMimeInfo(fi.Name())
			w.Header().Set("Content-Type", mimeinfo.Mimetype())
			err = srvutf.ServeUTF(fi, fsstat, fsopen, mimeinfo, params, ua, w, r)
			return
		}
	}
	w.Header().Set("Content-Type", mimeinfo[1])
	w.Header().Set("Content-Length", "0")
}

func serveWsSse(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, path string, ua useragent.UserAgent, srvws ServerWS, srvsse ServeSSE, w http.ResponseWriter, r *http.Request) (served bool, err error) {
	if served, err = websocket.UpgradeAndServe(w, r, func(w http.ResponseWriter, r *http.Request) error {
		var params = parameters.NewParameters()
		defer params.ClearAll()
		parameters.LoadParametersFromUrlValues(params, r.URL.Query())
		return srvws.ServeWS(fsstat, fsopen, path, params, ua, w, r)
	}); served {
		return
	}
	served, err = serversent.UpgradeAndServe(w, r, func(w http.ResponseWriter, r *http.Request) error {
		var params = parameters.NewParameters()
		defer params.ClearAll()
		parameters.LoadParametersFromUrlValues(params, r.URL.Query())
		return srvsse.ServeSSE(fsstat, fsopen, path, params, ua, w, r)
	})
	return
}

var mainua = useragent.NewParser()
