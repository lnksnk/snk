package mime

import (
	"embed"
	_ "embed"
	"encoding/csv"
	"io"

	"path/filepath"
	"strings"
	"sync"

	snkio "github.com/lnksnk/snk/io"
)

var UTF8exts = map[string]bool{}

func init() {
	utf8WebExtensions := []string{
		".html", ".htm", // Web pages
		".css",        // Stylesheets
		".js", ".mjs", // JavaScript
		".ts", ".tsx", // TypeScript
		".jsx",  // React JSX
		".json", // Data exchange
		".svg",  // Vector graphics
		".xml",  // Configuration/Data
		".md",   // Markdown documentation
		".txt",  // Plain text

		".m3u", //
	}
	for ui := range utf8WebExtensions {
		UTF8exts[utf8WebExtensions[ui]] = true
	}
}

type MimeInfo []string

func (mio MimeInfo) Media() bool {
	return !(len(mio) >= 3 && UTF8exts[mio[2]]) && len(mio) >= 2 && (strings.Contains(mio[1], "audio/") || strings.Contains(mio[1], "video/"))
}

func (mio MimeInfo) Audio() bool {
	return len(mio) >= 2 && strings.Contains(mio[1], "audio/")
}

func (mio MimeInfo) Video() bool {
	return len(mio) >= 2 && strings.Contains(mio[1], "video/")
}

func (mio MimeInfo) Name() string {
	if len(mio) >= 1 {
		return mio[0]
	}
	return ""
}

func (mio MimeInfo) Mimetype() (mimetype string) {
	mimetype = mio[1]
	if UTF8exts[mio[2]] {
		mimetype += "; charset=utf-8"
	}
	return mimetype
}

func (mio MimeInfo) Extension() string {
	if len(mio) >= 3 {
		return mio[2]
	}
	return ""
}

func (mio MimeInfo) Description() string {
	if len(mio) >= 4 {
		return mio[3]
	}
	return ""
}

func ExtMimeInfo(ext string) (mimeinfo MimeInfo) {
	if ext = filepath.Ext(ext); ext == "" {
		return strings.Split("Plain	text/plain	.txt	Plain Text", "\t")
	}
	if mimeinfod, mimeinfok := mimexts.Load(ext); mimeinfok {
		return mimeinfod
	}
	wg := &sync.WaitGroup{}
	wg.Go(func() {
		if mmfi, _ := MimeTypes.Open("mimetypes.csv"); mmfi != nil {
			mimebf := csv.NewReader(mmfi)
			mimebf.Comma = 9
			var rec []string
			var recerr error
			for {
				if rec, recerr = mimebf.Read(); (recerr == nil || recerr == io.EOF) && len(rec) == 4 {
					for n := range rec {
						rec[n] = strings.TrimSpace(rec[n])
					}
					if rec[2] == ext {
						mimeinfo = rec
						mimexts.Store(ext, mimeinfo)
						return
					}
					continue
				}
				break
			}
		}
	})
	wg.Wait()
	if ext != "" && len(mimeinfo) == 0 {
		mimeinfo = strings.Split("Plain	text/plain	.txt	Plain Text", "\t")
		mimeinfo[2] = ext
		mimexts.Store(ext, mimeinfo)
	}
	return
}

var mimexts = snkio.NewSyncMap[string, MimeInfo]()

//go:embed mimetypes.csv
var MimeTypes embed.FS
