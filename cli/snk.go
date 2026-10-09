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
	"github.com/lnksnk/snk/serve/fsserve"
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
	"github.com/lnksnk/snk/ui/db"
	"github.com/lnksnk/snk/w3css"

	"strings"
	"syscall"

	"github.com/grafana/sobek"
	"github.com/medama-io/go-useragent"
	_ "github.com/medama-io/go-useragent"
)

var sbkh = fssobek.SobekHandler()

func ESVmSetupRequestContext(ctx context.Context, dh DbHandler, fs Fsys, vm *sobek.Runtime, params parameters.Parameters, ua useragent.UserAgent, vmsetup func(*sobek.Runtime, ...fst.VMItem)) {
	vm.Set("escape", map[string]any{"url": map[string]any{"path": url.PathEscape, "query": url.QueryEscape}, "html": html.EscapeString})
	vm.Set("unescape", map[string]any{"url": map[string]any{"path": url.PathUnescape, "query": url.QueryUnescape}, "html": html.EscapeString})
	vm.Set("DB", dh)
	vm.Set("FS", fs)
	vmsetup(vm, []fst.VMItem{[]any{"params", params},
		[]any{"UA", ua},
		[]any{"uuid", func() string { return uuid.NewV7().String() }},
		[]any{"uuid4", func() string { return uuid.NewV4().String() }},
		[]any{"uuid7", func() string { return uuid.NewV7().String() }}}...)
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
	db.LoadQry("/snk/db/", fs.FSMap, fsys.Set)
	fs.FSMap("/htmx/", "")
	fsys.Set("/htmx/htmax.js", htmx.HtmaxJS)
	fsys.Set("/htmx/index.html", htmx.IndexHTML)
	var f, _ = os.Open(appconfpath + appname + "-conf.json")
	if f == nil {
		if f, _ = os.Open(osappconfpath + appname + "-conf.json"); f == nil && sourcepath != "" {
			f, _ = os.Open(sourcepath + appname + "-conf.json")
		}
	}
	if f != nil {
		fmt.Println("loading config file", f.Name())
	}
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
			if ln, lnerr := net.Listen(cl.Network, cl.Addr); ln != nil {
				lstnrs = append(lstnrs, ln)
			} else if lnerr != nil {
				fmt.Println(lnerr.Error())
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

	var hndlr http.Handler
	hndlr, _ = fsserve.ServeHttp(fsys, fsys, func(path string) {
		if prgmtmplts := sbkh.ProgramModules(); prgmtmplts != nil && prgmtmplts.Exist(path) {
			go prgmtmplts.Delete(path)
		}
	}, fsserve.ServeSTDFunc(func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
		var f fs.File
		if f, err = fsopen.Open(fi.Path()); f != nil {
			defer f.Close()
			snkio.Fprint(w, f)
			return
		}
		return
	}), fsserve.ServeUTFFunc(func(fi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, mimeinfo mime.MimeInfo, params parameters.Parameters, ua useragent.UserAgent, w http.ResponseWriter, r *http.Request) (err error) {
		fst.ServeHTTP(sbkh, w, r, fi, fsstat, fsopen, func(vm *sobek.Runtime, r *http.Request, vmsetup func(*sobek.Runtime, ...fst.VMItem)) {
			var db = InvokeDbHandler(r.Context(), snksql.FormatQueryFunc(func(out io.Writer, name, driver, query string, a ...any) (err error) {
				return sbkh.FormatQuery(vm, fsstat, fsopen, out, name, driver, query)
			}), params)
			var fs = InvokeFSys(fsys, fs.FSMap, fsys, fsstat, fsopen, fsys)
			vm.Set("hostip", func() string {
				return r.Host
			})
			ESVmSetupRequestContext(r.Context(), db, fs, vm, params, ua, vmsetup)
		})
		return
	}), fsserve.UTF8Ext(".html"), fsserve.UTF8Ext(".js"), fsserve.UTF8Ext(".css"), fsserve.UTF8Ext(".json"))
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
