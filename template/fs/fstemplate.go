package fs

import (
	"io/fs"
	"reflect"
	"time"

	snkfs "github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
)

type Template[P any, M any] interface {
	Content() *snkio.BufferWriter
	Program() P
	Module() M
	snkfs.FileInfo
	Valid(fstat snkfs.StatFileSystem) (valid bool)
	Eval(evalcntnt EvalContentFunc, evalprgmormod EvalProgramModuleFunc[P, M]) (err error)
	Close() error
}

type CompileCodeFunc[P any, M any] func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm P, modrec M, err error)

type prgrmtmplt[P any, M any] struct {
	unmatched map[string]bool
	matched   map[string]time.Time
	cntnt     snkio.BufferWriter
	pgrm      P
	modrec    M
	snkfs.FileInfo
	dependants map[string]time.Time
}

func (pt *prgrmtmplt[P, M]) Content() *snkio.BufferWriter {
	return &pt.cntnt
}
func (pt *prgrmtmplt[P, M]) Program() P {
	return pt.pgrm
}
func (pt *prgrmtmplt[P, M]) Module() M {
	return pt.modrec
}

func (pt *prgrmtmplt[P, M]) Name() string {
	return pt.FileInfo.Name()
}

func (pt *prgrmtmplt[P, M]) Path() string {
	return pt.FileInfo.Path()
}

func (pt *prgrmtmplt[P, M]) Root() string {
	return pt.FileInfo.Root()
}

func (pt *prgrmtmplt[P, M]) Base() string {
	return pt.FileInfo.Base()
}

func (pt *prgrmtmplt[P, M]) IsDir() bool {
	return pt.FileInfo.IsDir()
}

func (pt *prgrmtmplt[P, M]) Mode() fs.FileMode {
	return pt.FileInfo.Mode()
}

func (pt *prgrmtmplt[P, M]) Sys() any {
	return pt.FileInfo.Sys()
}

func (pt *prgrmtmplt[P, M]) Size() int64 {
	return pt.FileInfo.Size()
}

func (pt *prgrmtmplt[P, M]) Valid(fstat snkfs.StatFileSystem) (valid bool) {
	if unmatched, matched, dependants, prfi := pt.unmatched, pt.matched, pt.dependants, pt.FileInfo; prfi != nil {
		if fi, err := fstat.Stat(prfi.Path()); err == nil {
			if valid = prfi.ModTime().Equal(fi.ModTime()); valid {
				for unpth := range unmatched {
					if unfi, _ := fstat.Stat(unpth); unfi == nil {
						continue
					}
					return false
				}
				for mpth, mmd := range matched {
					if fi, err = fstat.Stat(mpth); err == nil && fi.ModTime().Equal(mmd) {
						continue
					}
					return false
				}
				for deppth, depmd := range dependants {
					if fi, err = fstat.Stat(deppth); err == nil && fi.ModTime().Equal(depmd) {
						continue
					}
					return false
				}
			}
			return
		}
	}
	return
}

func isReflectNil(v reflect.Value) bool {
	// 1. Check if the reflect.Value itself is valid (handles untyped nil)
	if !v.IsValid() {
		return true
	}

	// 2. Ensure the type's Kind is actually capable of being nil
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}

	return false
}

var nilv = reflect.ValueOf(nil)

func (pt *prgrmtmplt[P, M]) Eval(evalcntnt EvalContentFunc, evalprgmormod EvalProgramModuleFunc[P, M]) (err error) {
	cntnt, prgm, modrec, dependants := pt.cntnt, pt.pgrm, pt.modrec, pt.dependants
	var ispgrm, ismod = !isReflectNil(reflect.ValueOf(prgm)), !isReflectNil(reflect.ValueOf(modrec))
	if evalcntnt != nil && cntnt != nil && !cntnt.Empty() {
		if err = evalcntnt(cntnt); err == nil {
			if evalprgmormod != nil && (ispgrm || ismod) {
				return evalprgmormod(prgm, modrec, dependants)
			}
		}
		return
	}
	if evalprgmormod != nil && (ispgrm || ismod) {
		return evalprgmormod(prgm, modrec, dependants)
	}
	return
}

func (pt *prgrmtmplt[P, M]) Close() (err error) {
	cntnt := pt.cntnt
	pt.cntnt = nil
	if cntnt != nil {
		cntnt.Close()
	}
	return
}

type EvalContentFunc func(cntnt snkio.BufferWriter) (wrterr error)
type EvalProgramModuleFunc[P any, M any] func(prgm P, modrec M, dependants map[string]time.Time) (runerr error)

func EvalTemplate[P any, M any](
	templates snkio.Map[string, Template[P, M]],
	compile CompileCodeFunc[P, M], rtfi snkfs.FileInfo, fstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, evalcntnt EvalContentFunc, evalprgmormod EvalProgramModuleFunc[P, M]) (err error) {
	tmplsrc, tmpltk := templates.Load(rtfi.Path())
	if tmpltk {
		if tmplsrc.Valid(fstat) {
			return tmplsrc.Eval(evalcntnt, evalprgmormod)
		}
		templates.Delete(rtfi.Path())
	}
	var f snkfs.File

	if f, err = fsopen.Open(rtfi.Path()); f != nil {
		defer f.Close()
		if tmplsrc, err = LoadTempate(templates, compile, rtfi, fstat, fsopen, f); err == nil {
			return tmplsrc.Eval(evalcntnt, evalprgmormod)
		}
		return
	}
	return
}

func LoadTempate[P any, M any](templates snkio.Map[string, Template[P, M]], compile CompileCodeFunc[P, M], rtfi snkfs.FileInfo, fstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, a ...any) (tmpltsrc Template[P, M], err error) {
	var rscntnt, _ = snkio.WriterBuffer()
	defer rscntnt.Close()
	var unmatched = map[string]bool{}
	var matched = map[string]time.Time{}
	if err = FParseFS(rscntnt, unmatched, matched, rtfi, fstat, a...); err != nil {
		return
	}
	var cntnt, _ = snkio.WriterBuffer()
	var cde, _ = snkio.WriterBuffer()
	defer cde.Close()
	if err = FParseCbaseFS(cntnt, cde, rtfi, rscntnt); err == nil {
		var prgm P
		var modrec M
		if !cde.Empty() {
			var dependants = map[string]time.Time{}
			if prgm, modrec, err = compile(cde, rtfi, fstat, fsopen, dependants); err == nil {
				tmpltsrc = &prgrmtmplt[P, M]{unmatched: unmatched, matched: matched, dependants: dependants, cntnt: cntnt, pgrm: prgm, modrec: modrec, FileInfo: snkfs.GenFileInfo(rtfi.Path(), rtfi.Base(), rtfi.Size(), rtfi.IsDir(), rtfi.ModTime())}
				templates.Store(rtfi.Path(), tmpltsrc)
				return
			}
			cntnt.Close()
			return
		}
		tmpltsrc = &prgrmtmplt[P, M]{unmatched: unmatched, matched: matched, dependants: nil, cntnt: cntnt, pgrm: prgm, modrec: modrec, FileInfo: snkfs.GenFileInfo(rtfi.Path(), rtfi.Base(), rtfi.Size(), rtfi.IsDir(), rtfi.ModTime())}
		templates.Store(rtfi.Path(), tmpltsrc)
		return
	}
	cntnt.Close()
	return
}

type ResolveTemplatModule[P any, M any] func(specifier string, fndspecifierpath func(string, time.Time), compile CompileCodeFunc[P, M], rootfi fs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error)
