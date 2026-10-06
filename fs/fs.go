package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	snkio "github.com/lnksnk/snk/io"
)

// FileSystem provides acces to a general filesystem
// The FileSystem interface is the minimum implementation required of the filesystem.
type FileSystem interface {
	//Root() string
	//LocalRoot() string
	Stat(name string) (FileInfo, error)
	Set(name string, a ...any) error
	Append(name string, a ...any) error
	Open(name string) (File, error)
	ReadDir(name string) ([]DirEntry, error)
	List(path string) ([]FileInfo, error)
}

// Root interface of a [FileSystem]
type RootFileSystem interface {
	Root() string
}

type RootFileSystemFunc func() string

func (rtfsfunc RootFileSystemFunc) Root() string {
	return rtfsfunc()
}

// Local root interface of a [FileSystem]
// Usually point to physical path (local path) on os
type LocalRootFileSystem interface {
	LocalRoot() string
}

type LocalRootFileSystemFunc func() string

func (lclrtfsfunc LocalRootFileSystemFunc) LocalRoot() string {
	return lclrtfsfunc()
}

// Stat interface of a [FileSystem]
type StatFileSystem interface {
	Stat(string) (FileInfo, error)
}

type StatFileSystemFunc func(string) (FileInfo, error)

func (sttfsfnc StatFileSystemFunc) Stat(name string) (FileInfo, error) {
	return sttfsfnc(name)
}

// Open interface of a [FileSystem]
type OpenFileSystem interface {
	Open(string) (File, error)
}

type OpenFileSystemFunc func(string) (File, error)

func (opnfsfnc OpenFileSystemFunc) Open(name string) (File, error) {
	return opnfsfnc(name)
}

// ReadDir interface of a [FileSystem]
type ReadDirFileSystem interface {
	ReadDir(string) ([]DirEntry, error)
}

type ReadDirFileSystemFunc func(string) ([]DirEntry, error)

func (rddirfsfnc ReadDirFileSystemFunc) ReadDir(name string) ([]DirEntry, error) {
	return rddirfsfnc(name)
}

// List interface of a [FileSystem]
type ListFileSystem interface {
	List(path string) ([]FileInfo, error)
}

type ListFileSystemFunc func(string) ([]FileInfo, error)

func (lstfsfunc ListFileSystemFunc) List(path string) ([]FileInfo, error) {
	return lstfsfunc(path)
}

// Set interface of a [FileSystem]
type SetFileSystem interface {
	Set(name string, a ...any) error
}

type SetFileSystemFunc func(string, ...any) error

func (stfsfnc SetFileSystemFunc) Set(name string, a ...any) error {
	return stfsfnc(name, a...)
}

// Append interface of a [FileSystem]
type AppendFileSystem interface {
	Append(name string, a ...any) error
}

type AppendFileSystemFunc func(string, ...any) error

func (appndfsfnc AppendFileSystemFunc) Append(name string, a ...any) error {
	return appndfsfnc(name, a...)
}

type filesystem struct {
	fappndmp  snkio.Map[string, AppendFileSystem]
	fstmp     snkio.Map[string, SetFileSystem]
	fstsmp    snkio.Map[string, StatFileSystem]
	fopnmap   snkio.Map[string, OpenFileSystem]
	fsrddirmp snkio.Map[string, ReadDirFileSystem]
	fslklroot snkio.Map[string, string]
	LocalRootFileSystem
	RootFileSystem
}

// Append implements [FileSystem].
func (f *filesystem) Append(name string, a ...any) (err error) {
	if f == nil {
		err = &fs.PathError{Op: "Append", Path: name, Err: fmt.Errorf("nil [FileSystem]")}
		return
	}
	var fsroot, fslocalroot, foundroot = globallocalroot(f, name)
	if !foundroot {
		err = &fs.PathError{Op: "Append", Path: name, Err: fmt.Errorf("no root found for [FileSystem] path")}
		return
	}
	return appendFSFile(f.fappndmp, f.fstsmp, fsroot, fslocalroot, name, a...)
}

func appendFSFile(appendFileSystem snkio.Map[string, AppendFileSystem], statFileSystem snkio.Map[string, StatFileSystem], fsroot, fslocalroot, name string, a ...any) (err error) {
	if appendFileSystem == nil {
		err = &fs.PathError{Op: "appendFSFile", Path: name, Err: fmt.Errorf("nil snkio.Map[string, AppendFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "appendFSFile", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}
	if statFileSystem == nil {
		err = &fs.PathError{Op: "statFileSystem", Path: name, Err: fmt.Errorf("nil snkio.Map[string, StatFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "appendFileSystem", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}
	if name == "" {
		name = "/"
	}
	if name[0] != '/' {
		name = fsroot + name
	}

	if strings.HasPrefix(name, fsroot) && name[len(fsroot)-1] == '/' {
		var fsapndcaller, _ = appendFileSystem.Load(fsroot)
		if fsapndcaller == nil {
			if fslocalroot == "" {
				err = &fs.PathError{Op: "appendFileSystem", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
				return
			}
			fsapndcaller = OsAppendFS(fsroot, fslocalroot)
			appendFileSystem.Store(fsroot, fsapndcaller)
		}

		if err = fsapndcaller.Append(name, a...); err != nil {
			return
		}
		return
	}
	err = &fs.PathError{Op: "appendFileSystem", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
	return
}

// Set implements [FileSystem].
func (f *filesystem) Set(name string, a ...any) (err error) {
	if f == nil {
		err = &fs.PathError{Op: "Set", Path: name, Err: fmt.Errorf("nil [FileSystem]")}
		return
	}
	var fsroot, fslocalroot, foundroot = globallocalroot(f, name)
	if !foundroot {
		err = &fs.PathError{Op: "Set", Path: name, Err: fmt.Errorf("no root found for [FileSystem] path")}
		return
	}
	return setFSFile(f.fstmp, f.fstsmp, fsroot, fslocalroot, name, a...)
}

func setFSFile(setFileSystem snkio.Map[string, SetFileSystem], statFileSystem snkio.Map[string, StatFileSystem], fsroot, fslocalroot, name string, a ...any) (err error) {
	if setFileSystem == nil {
		err = &fs.PathError{Op: "setFSFile", Path: name, Err: fmt.Errorf("nil snkio.Map[string, SetFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "setFSFile", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}
	if statFileSystem == nil {
		err = &fs.PathError{Op: "setFileSystem", Path: name, Err: fmt.Errorf("nil snkio.Map[string, StatFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "setFileSystem", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}
	if name == "" {
		name = "/"
	}
	if name[0] != '/' {
		name = fsroot + name
	}

	if strings.HasPrefix(name, fsroot) && name[len(fsroot)-1] == '/' {
		var fsstcaller, _ = setFileSystem.Load(fsroot)
		if fsstcaller == nil {
			if fslocalroot == "" {
				err = &fs.PathError{Op: "setFileSystem", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
				return
			}
			fsstcaller = OsSetFS(fsroot, fslocalroot)
			setFileSystem.Store(fsroot, fsstcaller)
		}

		if err = fsstcaller.Set(name, a...); err != nil {
			return
		}
		return
	}
	err = &fs.PathError{Op: "statFileSystem", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
	return
}

// List implements [FileSystem].
func (f *filesystem) List(path string) ([]FileInfo, error) {
	return ListFileInfos(f, path)
}

func ListFileInfos(fs FileSystem, name string) (fios []FileInfo, err error) {
	var direntries []DirEntry
	if direntries, err = fs.ReadDir(name); err != nil {
		return
	}
	var fi FileInfo
	var fipths = map[string]bool{}
	defer clear(fipths)
	for _, die := range direntries {
		if fi, err = die.Info(); err == nil {
			if fipth := fi.Path(); !fipths[fipth] {
				fipths[fipth] = true
				fios = append(fios, fi)
				continue
			}
			continue
		}
		return
	}
	return
}

func globallocalroot(f *filesystem, name string) (fsroot, fslocalroot string, foundroot bool) {
	if nml := len(name); f != nil && nml > 0 {
		if lkproot, lkplocal := f.RootFileSystem, f.LocalRootFileSystem; lkproot != nil && lkplocal != nil {
			return lkproot.Root(), lkplocal.LocalRoot(), true
		}
		if gblrt := GlobalFSLocalRoot; gblrt != nil {
			if si := strings.LastIndex(name, "/"); si >= 0 {
				var mtchroot, mtchdlclroot = "", ""
				var mtchdl = 0
				gblrt.Range(func(root, localroot string) bool {
					rtl := len(root)
					if nml >= rtl && rtl > mtchdl && si >= rtl-1 && root == name[:rtl] {
						mtchdl = rtl
						mtchroot = root
						mtchdlclroot = localroot
					}
					return true
				})
				return mtchroot, mtchdlclroot, mtchdl > 0 && mtchroot != ""
			}
		}
	}
	return
}

// ReadDir implements [FileSystem].
func (f *filesystem) ReadDir(name string) (dirs []DirEntry, err error) {
	if f == nil {
		return
	}
	var fsroot, fslocalroot, foundroot = globallocalroot(f, name)
	if !foundroot {
		return
	}
	if !strings.HasPrefix(name, fsroot) {
		return
	}
	var fltr = ""
	if nmi := strings.LastIndex(name, "/"); nmi > -1 && strings.ContainsAny(name[nmi+1:], "*?,") {
		fltr = name[nmi+1:]
		name = name[:nmi+1]
	}
	return readDirFS(f.fsrddirmp, fsroot, fslocalroot, name, fltr)
}

func readDirFS(readDirFileSystem snkio.Map[string, ReadDirFileSystem], fsroot, fslocalroot, name, fltr string) (dirs []DirEntry, err error) {
	if readDirFileSystem == nil {
		err = &fs.PathError{Op: "readDirFS", Path: name, Err: fmt.Errorf("nil snkio.Map[string, ReadDirFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "readDirFS", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}

	if name == "" {
		name = "/"
	}
	if name != "/" {
		if name[0] != '/' {
			name = fsroot + name
		}
	}
	if strings.HasPrefix(name, fsroot) && name[len(fsroot)-1] == '/' {
		var fsreaddircaller, _ = readDirFileSystem.Load(fsroot)
		if fsreaddircaller == nil {
			if fslocalroot == "" {
				err = &fs.PathError{Op: "readDirFS", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
				return
			}
			fsreaddircaller = OsReadDirFS(fsroot, fslocalroot)
			readDirFileSystem.Store(fsroot, fsreaddircaller)
		}
		if dirs, err = fsreaddircaller.ReadDir(name); err == nil && fltr != "" {
			sbflrs := strings.Split(fltr, ",")
			di, dl := 0, len(dirs)
		nxtdi:
			for di < dl {
				for _, sbf := range sbflrs {
					if vld, _ := filepath.Match(sbf, dirs[di].Name()); vld {
						di++
						goto nxtdi
					}
				}
				dirs = append(dirs[:di], dirs[di+1:]...)
				dl--
			}
		}
		return
	}
	err = &fs.PathError{Op: "readDirFS", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
	return
}

// Open implements [FileSystem].
func (f *filesystem) Open(name string) (fl File, err error) {
	if f == nil {
		err = &fs.PathError{Op: "Open", Path: name, Err: fmt.Errorf("nil [FileSystem]")}
		return
	}
	var fsroot, fslocalroot, foundroot = globallocalroot(f, name)
	if !foundroot {
		err = &fs.PathError{Op: "Open", Path: name, Err: fmt.Errorf("no root found for [FileSystem] path")}
		return
	}
	return openFSFile(f.fopnmap, fsroot, fslocalroot, name)
}

func openFSFile(openFileSystem snkio.Map[string, OpenFileSystem], fsroot string, fslocalroot string, name string) (f File, err error) {
	if openFileSystem == nil {
		err = &fs.PathError{Op: "openFSFile", Path: name, Err: fmt.Errorf("nil snkio.Map[string, OpenFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "openFSFile", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}

	if name == "" {
		name = "/"
	}
	if name != "/" {
		if name[0] != '/' {
			name = fsroot + name
		}
	}

	if strings.HasPrefix(name, fsroot) && name[len(fsroot)-1] == '/' {
		var fsstscaller, _ = openFileSystem.Load(fsroot)
		if fsstscaller == nil {
			if fslocalroot == "" {
				err = &fs.PathError{Op: "openFSFile", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
				return
			}
			fsstscaller = OsOpenFS(fsroot, fslocalroot)
			openFileSystem.Store(fsroot, fsstscaller)
		}
		if f, err = fsstscaller.Open(name); err != nil {
			return
		}
		return
	}
	err = &fs.PathError{Op: "openFSFile", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
	return
}

// Stat implements [FileSystem].
func (f *filesystem) Stat(name string) (fi FileInfo, err error) {
	if f == nil {
		err = &fs.PathError{Op: "Stat", Path: name, Err: fmt.Errorf("nil [FileSystem]")}
		return
	}
	var fsroot, fslocalroot, foundroot = globallocalroot(f, name)
	if !foundroot {
		err = &fs.PathError{Op: "Open", Path: name, Err: fmt.Errorf("no root found for [FileSystem] path")}
		return
	}

	return statFSFileInfo(f.fstsmp, fsroot, fslocalroot, name)
}

func Stripelips(pth string) string {
	for {
		pei := strings.Index(pth, "/../")
		if pei > -1 {
			var prepth, postpth = pth[:pei], pth[pei+3:]
			if pei = strings.LastIndex(prepth, "/"); pei > -1 {
				pth = prepth[:pei] + postpth
				continue
			}
			pth = postpth
			continue
		}
		return pth
	}

}

func statFSFileInfo(statFileSystem snkio.Map[string, StatFileSystem], fsroot string, fslocalroot string, name string) (fi FileInfo, err error) {
	if statFileSystem == nil {
		err = &fs.PathError{Op: "statFileSystem", Path: name, Err: fmt.Errorf("nil snkio.Map[string, StatFileSystem]")}
		return
	}
	if fsroot == "" || (fsroot != "" && !(fsroot[0] == '/' && fsroot[len(fsroot)-1] == '/')) {
		err = &fs.PathError{Op: "statFileSystem", Path: name, Err: fmt.Errorf("invalid [RootFileSystem]")}
		return
	}
	if name == "" {
		name = "/"
	}
	if name[0] != '/' {
		name = fsroot + name
	}

	if strings.HasPrefix(name, fsroot) && name[len(fsroot)-1] == '/' {
		var fsstscaller, _ = statFileSystem.Load(fsroot)
		if fsstscaller == nil {
			if fslocalroot == "" {
				err = &fs.PathError{Op: "statFileSystem", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
				return
			}
			fsstscaller = OsStatFS(fsroot, fslocalroot)
			statFileSystem.Store(fsroot, fsstscaller)
		}

		fi, err = fsstscaller.Stat(name)
		return
	}
	err = &fs.PathError{Op: "statFileSystem", Path: name, Err: fmt.Errorf("unknown [RootFileSystem]")}
	return
}

func FSys(a ...any) FileSystem {
	var lclrtfs LocalRootFileSystem
	var rtfs RootFileSystem
	var fstmp snkio.Map[string, SetFileSystem]
	var fappndmp snkio.Map[string, AppendFileSystem]
	var fstsmp snkio.Map[string, StatFileSystem]
	var fsrddirmp snkio.Map[string, ReadDirFileSystem]
	var fopnmap snkio.Map[string, OpenFileSystem]
	var fslklroot snkio.Map[string, string]
	for ai := range a {
		if lclrtfsfuncd, lclrtfsfunck := a[ai].(LocalRootFileSystem); lclrtfsfunck {
			if lclrtfs == nil {
				lclrtfs = lclrtfsfuncd
			}
			continue
		}
		if rtfsfuncd, rtfsfunck := a[ai].(RootFileSystemFunc); rtfsfunck {
			if rtfs == nil {
				rtfs = rtfsfuncd
			}
			continue
		}
		if fstmpd, fstmpk := a[ai].(snkio.Map[string, SetFileSystem]); fstmpk {
			if fstmp == nil {
				fstmp = fstmpd
			}
			continue
		}
		if fappndmpd, fappndmpk := a[ai].(snkio.Map[string, AppendFileSystem]); fappndmpk {
			if fappndmp == nil {
				fappndmp = fappndmpd
			}
			continue
		}
		if fstsmpd, fstsmpk := a[ai].(snkio.Map[string, StatFileSystem]); fstsmpk {
			if fstsmp == nil {
				fstsmp = fstsmpd
			}
			continue
		}
		if fopnmapd, fopnmapk := a[ai].(snkio.Map[string, OpenFileSystem]); fopnmapk {
			if fopnmap == nil {
				fopnmap = fopnmapd
			}
			continue
		}
		if fsrddirmapd, fsrddirmapk := a[ai].(snkio.Map[string, ReadDirFileSystem]); fsrddirmapk {
			if fsrddirmp == nil {
				fsrddirmp = fsrddirmapd
			}
			continue
		}
		if fslklrootd, fslklrootk := a[ai].(snkio.Map[string, string]); fslklrootk {
			if fslklroot == nil {
				fslklroot = fslklrootd
			}
			continue
		}
	}
	if fstsmp == nil {
		fstsmp = GlobalFSStat
	}
	if fopnmap == nil {
		fopnmap = GlobalFSOpen
	}
	if fsrddirmp == nil {
		fsrddirmp = GlobalFSReadDir
	}
	if fslklroot == nil {
		fslklroot = GlobalFSLocalRoot
	}
	if fappndmp == nil {
		fappndmp = GlobalFSAppend
	}
	if fstmp == nil {
		fstmp = GlobalFSSet
	}
	return &filesystem{fappndmp: fappndmp, fstmp: fstmp, fstsmp: fstsmp, fopnmap: fopnmap, fsrddirmp: fsrddirmp, fslklroot: fslklroot, LocalRootFileSystem: lclrtfs, RootFileSystem: rtfs}
}

// Global FileSystem ReadDir [[]DirEntry] [snkio.Map]
var GlobalFSReadDir = snkio.NewMap[string, ReadDirFileSystem]()

// Global FileSystem Stat [FileInfo] [snkio.Map]
var GlobalFSStat = snkio.NewMap[string, StatFileSystem]()

// Global FileSystem Open [File] [snkio.Map]
var GlobalFSOpen = snkio.NewMap[string, OpenFileSystem]()

// Global FileSystem LocalRoot [snkio.Map]
var GlobalFSLocalRoot = snkio.NewMap[string, string]()

// Global Filesystem Set [snkio.Map]
var GlobalFSSet = snkio.NewMap[string, SetFileSystem]()

// Global Filesystem Append [snkio.Map]
var GlobalFSAppend = snkio.NewMap[string, AppendFileSystem]()

// Global MemFile [snkio.Map]
var GlobalMemFile = snkio.NewMap[string, MemFile]()
