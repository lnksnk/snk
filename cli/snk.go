package main

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"uuid"

	"github.com/lnksnk/snk/fs"
	"github.com/lnksnk/snk/htmx"
	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/mime"
	"github.com/lnksnk/snk/parameters"
	"github.com/lnksnk/snk/serve"
	"github.com/lnksnk/snk/serversent"
	snksql "github.com/lnksnk/snk/sql"
	_ "github.com/lnksnk/snk/sql/driver/mssql"
	_ "github.com/lnksnk/snk/sql/driver/mysql"
	_ "github.com/lnksnk/snk/sql/driver/ora"
	_ "github.com/lnksnk/snk/sql/driver/postgres"
	_ "github.com/lnksnk/snk/sql/driver/sqlite"
	snksqlfs "github.com/lnksnk/snk/sql/fs"
	fst "github.com/lnksnk/snk/template/fs"
	"github.com/lnksnk/snk/template/fs/fssobek"
	"github.com/lnksnk/snk/ui"
	"github.com/lnksnk/snk/ui/db/qry"
	"github.com/lnksnk/snk/w3css"
	"github.com/lnksnk/snk/websocket"

	"strings"
	"syscall"

	"github.com/grafana/sobek"
	"github.com/medama-io/go-useragent"
	_ "github.com/medama-io/go-useragent"
)

var sbkh = fssobek.SobekHandler()

func ESDBEnv(ctx context.Context, vm *sobek.Runtime, params parameters.Parameters, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem) map[string]any {
	var frmtsqlqry = snksql.FormatQueryFunc(func(out io.Writer, name, driver, query string, a ...any) (err error) {
		return sbkh.FormatQuery(vm, fsstat, fsopen, out, name, driver, query)
	})
	return map[string]any{
		"query": func(name string, query string, a ...any) (records func(func(snksql.Record, int64) bool), err error) {

			if len(a) > 0 {
				a = append([]any{frmtsqlqry}, a...)
			}
			if len(a) == 0 {
				a = append(a, frmtsqlqry)
			}
			a = append(a, params)
			var rws, rwserr = snksql.QueryContext(ctx, name, query, a...)
			if err = rwserr; err != nil {
				records = snksql.NumberedRecords(nil)
				return
			}
			records = snksql.NumberedRecords(rws)
			return
		},
		"exec": func(name string, query string, a ...any) (any, error) {
			if len(a) > 0 {
				a = append([]any{frmtsqlqry}, a...)
			}
			if len(a) == 0 {
				a = append(a, frmtsqlqry)
			}
			a = append(a, params)
			return snksql.ExecContext(ctx, name, query, a...)
		},
		"stats": func(name string) (stats any) {
			return snksql.ConnStats(name)
		},
		"conns": func() []string {
			return snksql.ConnDefinitions()
		},
		"conndef": snksql.DefineConn,
	}
}

func main() {

	var apppath = strings.ReplaceAll(os.Args[0], "\\", "/")
	var appname = apppath[strings.LastIndex(apppath, "/")+1:]
	apppath = apppath[:strings.LastIndex(apppath, "/")+1]
	if appext := filepath.Ext(appname); appext != "" {
		appname = appname[:len(appname)-len(appext)]
	}
	if strings.HasPrefix(appname, "__debug_bin") {
		appname = "__debug_bin"
	}
	var appconfpath = apppath
	var osappconfpath = ""
	var sourcepath = ""
	var serveaddr = ""
	for _, osarg := range os.Args[:] {
		if strings.HasPrefix(osarg, "conf-path=") {
			if osarg = strings.ReplaceAll(osarg[len("-conf-path="):], "\\", "/"); osarg != "" && osarg[len(osarg)-1] == '/' {
				if fi, _ := os.Stat(osarg); fi != nil && fi.IsDir() {
					appconfpath = osarg
				}
			}
			continue
		}
		if strings.HasPrefix(osarg, "serve-addr=") {
			if osarg = osarg[len("serve-addr="):]; osarg != "" {
				if serveaddr == "" {
					serveaddr = osarg
				}
			}
			continue
		}
		if strings.HasPrefix(osarg, "source-path=") {
			if osarg = strings.ReplaceAll(osarg[len("source-path="):], "\\", "/"); osarg != "" && osarg[len(osarg)-1] == '/' {
				if fi, _ := os.Stat(osarg); fi != nil && fi.IsDir() {
					sourcepath = osarg
				}
			}
			continue
		}
	}
	var defaultlocalpath = "./"
	if sourcepath != "" {
		if osappconfpath == "" {
			osappconfpath = sourcepath
		}
		defaultlocalpath = sourcepath
	}

	var fsys = fs.FSys()
	fs.FSMap("/w3css/", "")
	fsys.Set("/w3css/w3.css", w3css.W3CSS)
	fsys.Set("/w3css/index.html", w3css.IndexHTML)

	fs.FSMap("/parsing/", "")
	fsys.Set("/parsing/parser.js", ui.ParserJS)
	fsys.Set("/parsing/index.html", ui.IndexHTML)
	qry.LoadQry("/snk/qry/", fs.FSMap, fsys.Set)
	fs.FSMap("/htmx/", "")
	fsys.Set("/htmx/htmax.js", htmx.HtmaxJS)
	fsys.Set("/htmx/index.html", htmx.IndexHTML)
	var f, _ = os.Open(appconfpath + appname + "-conf.json")
	if f == nil {
		if f, _ = os.Open(osappconfpath + appname + "-conf.json"); f == nil && sourcepath != "" {
			f, _ = os.Open(sourcepath + appname + "-conf.json")
		}
	}
	fmt.Println("loading config file", f.Name())
	var confbfw snkio.BufferWriter
	var lstnrs []net.Listener
	if f != nil {
		confbfw, _ = snkio.WriterBuffer(f)
		defer f.Close()
	}
	if confbfw != nil {
		var confrdr, _ = confbfw.Reader()
		var conf = &Config{Listen: []ConfigListener{}, SqlConns: []ConfigSqlConn{}, Resources: []ConfigResource{}}
		json.UnmarshalRead(confrdr, conf)
		if serveaddr != "" {
			conf.Listen = append(conf.Listen, ConfigListener{Network: "tcp", Addr: serveaddr})
		}
		for _, rc := range conf.Resources {
			if rc.Path == "/" {
				defaultlocalpath = ""
			}
		}
		if defaultlocalpath != "" {
			conf.Resources = append(conf.Resources, ConfigResource{Path: "/", LocalRoot: defaultlocalpath})
		}
		conf.Load(func(cl ConfigListener) {
			if ln, _ := net.Listen(cl.Network, cl.Addr); ln != nil {
				lstnrs = append(lstnrs, ln)
			}
		}, func(cr ConfigResource) {
			fs.FSMap(cr.Path, cr.LocalRoot)
		}, func(csc ConfigSqlConn) {
			snksql.DefineConn(csc.Name, csc.Datasource)
		})
	}

	snksql.GlobalFormatQuery = snksql.FormatQueryFunc(func(out io.Writer, name, driver, query string, a ...any) (err error) {
		var sqlfi fs.FileInfo
		if sqlfi, err = snksqlfs.FSStat(name, driver, query, func(unmtchdpth string) {
			if prgmtmplts := sbkh.ProgramModules(); prgmtmplts != nil && prgmtmplts.Exist(unmtchdpth) {
				go prgmtmplts.Delete(unmtchdpth)
			}
		}, fsys); err == nil {
			if f, _ := fsys.Open(sqlfi.Path()); f != nil {
				defer f.Close()
				_, err = snkio.Fprint(out, f)
			}
			return
		}
		return
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	mainua := useragent.NewParser()
	var hndlr = http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var served, srverr = websocket.UpgradeAndServe(w, r, func(w http.ResponseWriter, r *http.Request) error {

			return nil
		})
		if served {
			return
		}
		if served, srverr = serversent.UpgradeAndServe(w, r, func(w http.ResponseWriter, r *http.Request) error {

			return nil
		}); served {
			return
		}
		var params, _ = parameters.HTTPRequestParameters(r)
		defer params.ClearAll()
		var path = r.URL.Path
		var mimeinfo = mime.ExtMimeInfo(path)
		var fsts, err = fsys.Stat(path)
		if ext := filepath.Ext(path); ext == ".babylon" {
			ext += ""
		}
		var ua = mainua.Parse(r.UserAgent())
		if ua.IsBot() {
			w.Header().Set("Content-Length", "0")
			return
		}
		defer func() {
			if srverr != nil {
				err = srverr
			}
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
		go func() {
			<-r.Context().Done()
			if err != nil {

			}
		}()
		if fsts != nil {
			if fsts.IsDir() {
				for ext := range mime.UTF8exts {
					if fsts, err = fsys.Stat(path + "index" + ext); fsts == nil {
						if prgmtmplts := sbkh.ProgramModules(); prgmtmplts != nil && prgmtmplts.Exist(path) {
							go prgmtmplts.Delete(path)
						}
						continue
					}
					path += "index" + ext
					break
				}
			}
			mimeinfo = mime.ExtMimeInfo(fsts.Name())
			w.Header().Set("Content-Type", mimeinfo.Mimetype())
			if mimeinfo.Media() {
				var f fs.File
				if f, err = fsys.Open(fsts.Path()); err == nil {
					defer f.Close()
					var rskrclsr, _ = f.(io.ReadSeekCloser)
					serve.ServeHttpRange(w, r, rskrclsr, fsts.Size())
					return
				}
				w.Header().Set("Content-Length", fmt.Sprintf("%d", 0))
				return
			}

			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			if !mime.UTF8exts[filepath.Ext(fsts.Path())] {
				var f fs.File
				if f, err = fsys.Open(path); f != nil {
					defer f.Close()
					snkio.Fprint(w, f)
					return
				}
				w.Header().Set("Content-Type", mimeinfo[1])
				w.Header().Set("Content-Length", "0")
				return
			}
			fst.ServeHTTP(sbkh, w, r, fsts, fsys, fsys, func(vm *sobek.Runtime, r *http.Request, vmsetup func(*sobek.Runtime, ...fst.VMItem)) {
				vm.Set("escape", map[string]any{"url": map[string]any{"path": url.PathEscape, "query": url.QueryEscape}, "html": html.EscapeString})
				vm.Set("unescape", map[string]any{"url": map[string]any{"path": url.PathUnescape, "query": url.QueryUnescape}, "html": html.EscapeString})

				vm.Set("hostip", func() string {
					return r.Host
				})
				vmsetup(vm, []fst.VMItem{[]any{"params", params},
					[]any{"ua", ua},
					[]any{"fsys", map[string]any{
						"list":   fsys.List,
						"map":    fs.FSMap,
						"set":    fsys.Set,
						"stat":   fsys.Stat,
						"open":   fsys.Open,
						"append": fsys.Append}},
					[]any{"uuid", func() string { return uuid.NewV7().String() }},
					[]any{"uuid4", func() string { return uuid.NewV4().String() }},
					[]any{"uuid7", func() string { return uuid.NewV7().String() }},
					[]any{"db", ESDBEnv(r.Context(), vm, params, fsys, fsys)}}...)
			})
			return
		}
		w.Header().Set("Content-Type", mimeinfo[1])
		w.Header().Set("Content-Length", "0")
	}))
	for _, ln := range lstnrs {
		go http.Serve(ln, hndlr)
	}
	<-quit
	for _, ln := range lstnrs {
		ln.Close()
	}
	for _, cndef := range snksql.ConnDefinitions() {
		snksql.CloseConnDef(cndef)
	}
}
