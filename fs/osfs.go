package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	snkio "github.com/lnksnk/snk/io"
)

// OsOpenFS return an [OpenFileSystem] handler passing a fsroot and fslocalroot invoking an implementation of [OpenFileSystemFunc]
func OsOpenFS(fsroot, fslocalroot string) OpenFileSystemFunc {
	return func(name string) (File, error) {
		return osopen(name, fsroot, fslocalroot)
	}
}

// OsStatFS return an [StatFileSystem] handler passing a fsroot and fslocalroot invoking an implementation of [StatFileSystemFunc]
func OsStatFS(fsroot, fslocalroot string) StatFileSystemFunc {
	return func(name string) (FileInfo, error) {
		return osstat(name, fsroot, fslocalroot)
	}
}

// OsSetFS return an [SetFileSystem] handler passing a fsroot and fslocalroot invoking an implementation of [SetFileSystemFunc]
func OsSetFS(fsroot, fslocalroot string) SetFileSystemFunc {
	return func(name string, a ...any) error {
		return osset(name, fsroot, fslocalroot, a...)
	}
}

// OsAppendFS return an [AppendFileSystem] handler passing a fsroot and fslocalroot invoking an implementation of [AppendFileSystemFunc]
func OsAppendFS(fsroot, fslocalroot string, a ...any) AppendFileSystemFunc {
	return func(name string, a ...any) error {
		return osappend(name, fsroot, fslocalroot, a...)
	}
}

// OsReadDirFS return an [ReadDirFileSystem] handler passing a fsroot and fslocalroot invoking an implementation of [ReadDirFileSystemFunc]
func OsReadDirFS(fsroot, fslocalroot string) ReadDirFileSystemFunc {
	return func(name string) ([]DirEntry, error) {
		return osreaddir(name, fsroot, fslocalroot)
	}
}

func osreaddir(name, fsroot, fslocalroot string) (dirs []DirEntry, dirserr error) {
	var osdrs []os.DirEntry
	var osfi fs.FileInfo
	var fi FileInfo
	if osdrs, dirserr = os.ReadDir(fslocalroot + name[len(fsroot):]); dirserr == nil {
		for _, osdr := range osdrs {
			if osfi, dirserr = osdr.Info(); dirserr == nil {
				fi = GenFileInfo(fsroot+name[len(fsroot):]+osfi.Name(), fsroot, osfi.Size(), osfi.IsDir(), osfi.ModTime())
				dirs = append(dirs, dirEntry(fi))
				continue
			}
			return
		}
	}
	return
}

type osfile struct {
	fi  FileInfo
	osf fs.File
	skr io.Seeker
}

func (osf *osfile) Close() (err error) {
	if f := osf.osf; f != nil {
		osf.osf = nil
		f.Close()
	}
	return
}

func (osf *osfile) Stat() (FileInfo, error) {
	return osf.fi, nil
}

func (osf *osfile) Read(p []byte) (n int, err error) {
	if f := osf.osf; f != nil {
		if n, err = f.Read(p); err != nil {
			osf.Close()
		}
		return
	}
	return 0, io.EOF
}

func (osf *osfile) Seek(offset int64, whence int) (int64, error) {
	if f, skr := osf.osf, osf.skr; f != nil && skr != nil {
		return skr.Seek(offset, whence)
	}
	return 0, io.EOF
}

func osopen(name string, fsroot, fslocalroot string) (f File, err error) {
	defer func() {
		if f == nil {
			if err == nil {
				err = &fs.PathError{Op: "opopen", Path: name, Err: fmt.Errorf("invalid path")}
			}
		}
	}()
	var fi FileInfo
	if fi, err = osstat(name, fsroot, fslocalroot); err != nil {
		return
	}
	var osf fs.File
	if osf, err = os.Open(strings.ReplaceAll(fslocalroot+name[len(fsroot):], "//", "/")); err != nil {
		return osarchiveopen(name, fsroot, fslocalroot)
	}
	var skr, _ = osf.(io.Seeker)
	f = &osfile{osf: osf, fi: fi, skr: skr}
	return
}

func osstat(name string, fsroot, fslocalroot string) (fi FileInfo, err error) {
	defer func() {
		if fi == nil {
			if err == nil {
				err = &fs.PathError{Op: "osstat", Path: name, Err: fmt.Errorf("invalid path")}
			}
		}
	}()
	var osfi fs.FileInfo
	if osfi, err = os.Stat(fslocalroot + name[len(fsroot):]); err != nil {
		return osarchivestat(name, fsroot, fslocalroot)
	}
	fi = GenFileInfo(name, fsroot, osfi.Size(), osfi.IsDir(), osfi.ModTime())
	return
}

type archstat func(name string, fsroot, fsarchpath, fsarchname string, fsarchmod time.Time) (fi FileInfo, err error)
type archopen func(name string, fsroot, fsarchpath, fsarchname string, fsarchmod time.Time) (fi File, err error)

var ArchExtsStat = map[string]archstat{".zip": func(name, fsroot, fsarchpath, fsarchname string, fsarchmod time.Time) (fi FileInfo, err error) {
	//
	var r *zip.ReadCloser
	if r, err = zip.OpenReader(fsarchpath); err != nil {
		return
	}
	defer r.Close()
	var fsarchroot = func() string {
		if si := strings.LastIndex(fsarchname, "/"); si > -1 {
			return fsarchname[:si+1]
		}
		return "/"
	}()
	if fsarchroot[0] == '/' {
		fsarchroot = fsarchroot[1:]
	}
	for rfi := range r.File {
		fh := r.File[rfi].FileHeader
		if len(fh.Name) >= len(fsarchroot) && fh.Name[:len(fsarchroot)] == fsarchroot {
			if "/"+fh.Name == fsarchname {
				fi = GenFileInfo(name, fsroot, fh.FileInfo().Size(), fh.FileInfo().IsDir(), fh.FileInfo().ModTime())
				return
			}
		}
		if fh = r.File[len(r.File)-(rfi+1)].FileHeader; len(fh.Name) >= len(fsarchroot) && fh.Name[:len(fsarchroot)] == fsarchroot {
			if "/"+fh.Name == fsarchname {
				fi = GenFileInfo(name, fsroot, fh.FileInfo().Size(), fh.FileInfo().IsDir(), fh.FileInfo().ModTime())
				return
			}
		}
	}
	return
}}

var ArchExtsOpen = map[string]archopen{".zip": func(name, fsroot, fsarchpath, fsarchname string, fsarchmod time.Time) (f File, err error) {
	//
	var r *zip.ReadCloser
	if r, err = zip.OpenReader(fsarchpath); err != nil {
		return
	}
	defer r.Close()
	var fsarchroot = func() string {
		if si := strings.LastIndex(fsarchname, "/"); si > -1 {
			return fsarchname[:si+1]
		}
		return "/"
	}()
	if fsarchroot[0] == '/' {
		fsarchroot = fsarchroot[1:]
	}
	if fsarchroot[0] == '/' {
		fsarchroot = fsarchroot[1:]
	}
	for rfi := range r.File {
		fh := r.File[rfi].FileHeader
		if len(fh.Name) >= len(fsarchroot) && fh.Name[:len(fsarchroot)] == fsarchroot {
			if "/"+fh.Name == fsarchname {
				arcsf, arcerr := r.File[rfi].Open()
				if arcsf != nil {
					if bfw, _ := snkio.WriterBuffer(arcsf); bfw != nil {
						if bfr, _ := bfw.Reader(); bfr != nil {
							f = &archfile{bfr: bfr, fi: GenFileInfo(name, fsroot, fh.FileInfo().Size(), fh.FileInfo().IsDir(), fh.FileInfo().ModTime())}
							return
						}
					}
				}
				return nil, arcerr
			}
		}
		if fh = r.File[len(r.File)-(rfi+1)].FileHeader; len(fh.Name) >= len(fsarchroot) && fh.Name[:len(fsarchroot)] == fsarchroot {
			if "/"+fh.Name == fsarchname {
				arcsf, arcerr := r.File[len(r.File)-(rfi+1)].Open()
				if arcsf != nil {
					if bfw, _ := snkio.WriterBuffer(arcsf); bfw != nil {
						if bfr, _ := bfw.Reader(); bfr != nil {
							f = &archfile{bfr: bfr, fi: GenFileInfo(name, fsroot, fh.FileInfo().Size(), fh.FileInfo().IsDir(), fh.FileInfo().ModTime())}
							return
						}
					}
				}
				return nil, arcerr
			}
		}
	}
	return
}}

type archfile struct {
	fi      FileInfo
	bfr     snkio.BufferReader
	archext string
	archpth string
	archnme string
}

func (af *archfile) Stat() (FileInfo, error) {
	return af.fi, nil
}

func (af *archfile) Read(p []byte) (n int, err error) {
	if bfr := af.bfr; bfr != nil {
		if n, err = bfr.Read(p); err != nil {
			af.bfr = nil
			bfr.Close()
		}
		return
	}
	return 0, io.EOF
}

func (af *archfile) Close() (err error) {
	bfr := af.bfr
	af.bfr = nil
	if bfr != nil {
		bfr.Close()
	}
	return
}

func osarchivestat(name string, fsroot, fslocalroot string) (fi FileInfo, err error) {
	var archstt archstat
	var fsfullpath = fslocalroot + name[len(fsroot):]
	var fsfl = len(fsfullpath)
	for ri := range fsfullpath {
		if fsfullpath[ri] == '/' {
			if fsext := filepath.Ext(fsfullpath[:ri]); fsext != "" {
				if fzpfi, _ := os.Stat(fsfullpath[:ri] + fsext); fzpfi != nil {
					if archstt = ArchExtsStat[fsext]; archstt != nil {
						return archstt(name, fsroot, fsfullpath[:ri], fsfullpath[ri:], fzpfi.ModTime())
					}
					return
				}
			}
			for fsext := range ArchExtsStat {
				if fzpfi, _ := os.Stat(fsfullpath[:ri] + fsext); fzpfi != nil {
					if archstt = ArchExtsStat[fsext]; archstt != nil {
						return archstt(name, fsroot, fsfullpath[:ri]+fsext, fsfullpath[ri:], fzpfi.ModTime())
					}
					return
				}
			}
		}
		if tri := fsfl - (ri + 1); tri > ri {
			if fsfullpath[tri] == '/' {
				if fsext := filepath.Ext(fsfullpath[:tri]); fsext != "" {
					if fzpfi, _ := os.Stat(fsfullpath[:tri]); fzpfi != nil {
						if archstt = ArchExtsStat[fsext]; archstt != nil {
							return archstt(name, fsroot, fsfullpath[:tri], fsfullpath[tri:], fzpfi.ModTime())
						}
						return
					}
				}
				for fsext := range ArchExtsStat {
					if fzpfi, _ := os.Stat(fsfullpath[:tri] + fsext); fzpfi != nil {
						if archstt = ArchExtsStat[fsext]; archstt != nil {
							return archstt(name, fsroot, fsfullpath[:tri]+fsext, fsfullpath[tri:], fzpfi.ModTime())
						}
						return
					}
				}
			}
			continue
		}
		return
	}
	return
}

func osarchiveopen(name string, fsroot, fslocalroot string) (f File, err error) {
	var archopn archopen
	var fsfullpath = fslocalroot + name[len(fsroot):]
	var fsfl = len(fsfullpath)
	var fsext string
	for ri := range fsfullpath {
		if fsfullpath[ri] == '/' {
			if fsext = filepath.Ext(fsfullpath[:ri]); fsext != "" {
				if fzpfi, _ := os.Stat(fsfullpath[:ri] + fsext); fzpfi != nil {
					if archopn = ArchExtsOpen[fsext]; archopn != nil {
						return archopn(name, fsroot, fsfullpath[:ri], fsfullpath[ri:], fzpfi.ModTime())
					}
				}
			}
			for fsext := range ArchExtsOpen {
				if fzpfi, _ := os.Stat(fsfullpath[:ri] + fsext); fzpfi != nil {
					if archopn = ArchExtsOpen[fsext]; archopn != nil {
						return archopn(name, fsroot, fsfullpath[:ri]+fsext, fsfullpath[ri:], fzpfi.ModTime())
					}
				}
			}
		}
		if tri := fsfl - (ri + 1); tri > ri {
			if fsfullpath[tri] == '/' {
				if fsext = filepath.Ext(fsfullpath[:tri]); fsext != "" {
					if fzpfi, _ := os.Stat(fsfullpath[:tri]); fzpfi != nil {
						if archopn = ArchExtsOpen[fsext]; archopn != nil {
							return archopn(name, fsroot, fsfullpath[:tri], fsfullpath[tri:], fzpfi.ModTime())
						}
						return
					}
				}
				for fsext := range ArchExtsOpen {
					if fzpfi, _ := os.Stat(fsfullpath[:tri] + fsext); fzpfi != nil {
						if archopn = ArchExtsOpen[fsext]; archopn != nil {
							return archopn(name, fsroot, fsfullpath[:tri]+fsext, fsfullpath[tri:], fzpfi.ModTime())
						}
					}
				}
			}
			continue
		}
		return
	}
	return
}

func osset(name string, fsroot, fslocalroot string, a ...any) (err error) {
	defer func() {
		if err != nil {
			err = &fs.PathError{Op: "osset", Path: name, Err: err}
		}
	}()
	/*var osfi fs.FileInfo
	if osfi, err = os.Stat(fslocalroot + name[len(fsroot):]); err != nil {
		return
	}*/
	//fi = GenFileInfo(name, fsroot, osfi.Size(), osfi.IsDir(), osfi.ModTime())
	return
}

func osappend(name string, fsroot, fslocalroot string, a ...any) (err error) {
	defer func() {
		if err != nil {
			err = &fs.PathError{Op: "osset", Path: name, Err: err}
		}
	}()
	/*var osfi fs.FileInfo
	if osfi, err = os.Stat(fslocalroot + name[len(fsroot):]); err != nil {
		return
	}*/
	//fi = GenFileInfo(name, fsroot, osfi.Size(), osfi.IsDir(), osfi.ModTime())
	return
}
