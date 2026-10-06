package fs

import (
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	snkfs "github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
)

func FSStat(name, driver, query string, unmatchpath func(string), fsstat snkfs.StatFileSystem) (sqlfi snkfs.FileInfo, err error) {
	if name != "" && driver != "" && query != "" && fsstat != nil {
		if query != "" && query[0] == '/' {
			var qryrt = query[:strings.LastIndex(query, "/")+1]

			if query = query[strings.LastIndex(query, "/")+1:]; query != "" {
				ext := filepath.Ext(query)
				if ext != "" && ext != ".sql" {
					return
				}
				if ext != "" && ext == ".sql" {
					query = query[:len(query)-len(ext)]
				}
				if ext == "" {
					ext = ".sql"
				}
				var path = qryrt + query + "." + driver + ext
				if sqlfi, _ = fsstat.Stat(path); sqlfi == nil {
					if unmatchpath != nil {
						unmatchpath(path)
					}
					path = qryrt + query + "." + name + ext
					if sqlfi, _ = fsstat.Stat(path); sqlfi == nil {
						if unmatchpath != nil {
							unmatchpath(path)
						}
						path = qryrt + query + "." + name + "." + driver + ext
						if sqlfi, _ = fsstat.Stat(path); sqlfi == nil {
							if unmatchpath != nil {
								unmatchpath(path)
							}
							path = qryrt + query + ext
							if sqlfi, _ = fsstat.Stat(path); sqlfi == nil {
								if unmatchpath != nil {
									unmatchpath(path)
								}
								return nil, &fs.PathError{Op: "FSStat", Path: path, Err: fs.ErrNotExist}
							}
							return
						}
						return
					}
					return
				}
				return
			}
		}
	}
	err = &fs.PathError{Op: "FSStat", Path: "", Err: fs.ErrNotExist}
	return
}

func FSFormatQuery(out io.Writer, name, driver, query string, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, unmatchpath func(string), foundsqlfi func(out io.Writer, sqlfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) error) (err error) {
	if fsstat == nil && fsopen == nil {
		return
	}
	var sqlfi snkfs.FileInfo
	if sqlfi, err = FSStat(name, driver, query, unmatchpath, fsstat); sqlfi != nil && err == nil && foundsqlfi != nil {
		err = foundsqlfi(out, sqlfi, fsstat, fsopen)
		return
	}
	_, err = snkio.Fprint(out, query)
	return
}
