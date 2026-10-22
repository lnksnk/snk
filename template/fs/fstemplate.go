package fs

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"time"

	snkfs "github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/shortning"
)

type Template[P any, M any] interface {
	Content() *snkio.BufferWriter
	Program() P
	Module() M
	snkfs.FileInfo
	Valid(fstat snkfs.StatFileSystem) (valid bool)
	Eval(evalcntnt EvalContentFunc, evalprgmormod EvalProgramModuleFunc[P, M]) (result any, err error)
	Close() error
}

type CompileCodeFunc[P any, M any] func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error)

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

func (pt *prgrmtmplt[P, M]) Eval(evalcntnt EvalContentFunc, evalprgmormod EvalProgramModuleFunc[P, M]) (result any, err error) {
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
type EvalProgramModuleFunc[P any, M any] func(prgm P, modrec M, dependants map[string]time.Time) (runresult any, runerr error)

func EvalTemplate[P any, M any](
	templates snkio.Map[string, Template[P, M]],
	compile CompileCodeFunc[P, M], rtfi snkfs.FileInfo, dfltext string, fstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, evalcntnt EvalContentFunc, evalprgmormod EvalProgramModuleFunc[P, M], a ...any) (result any, err error) {
	al := len(a)
	var cancache = true
	var f snkfs.File
	if al > 0 {
		var rtfid snkfs.FileInfo
		if s, sk := a[0].(string); sk {
			if rtfid, _ = fstat.Stat(s); rtfid != nil {
				a = a[1:]
				if al--; al > 0 {
					cancache = false
					if f, err = fsopen.Open(rtfid.Path()); f == nil {
						return
					}
				}
				rtfi = rtfid
			}
		}
		if rtfid == nil {
			if rtfid, _ = a[0].(snkfs.FileInfo); rtfid != nil {
				a = a[1:]
				if al--; al > 0 {
					cancache = false
					if f, err = fsopen.Open(rtfid.Path()); f == nil {
						return
					}
				}
				rtfi = rtfid
			}
		}
		if rtfid == nil {
			cancache = false
			if dfltext = filepath.Ext(dfltext); dfltext == "" {
				return
			}
			if !rtfi.IsDir() {
				var rtext = filepath.Ext(rtfi.Path())
				if rtext == "" {
					rtext = dfltext
				}
				var chkpth = rtfi.Root() + shortning.Encode(time.Now().UnixNano()) + rtext
				rtfi = snkfs.GenFileInfo(chkpth, rtfi.Base(), 0, rtfi.IsDir(), time.Now())
				/*ai := 0
				var bfa []any
				for ai < al {
					if _, argsk := a[ai].(map[string]any); !argsk {
						bfa = append(bfa, a[ai])
						a = append(a[:ai], a[ai+1:]...)
						al--
						continue
					}
					ai++
				}
				var bfw snkio.BufferWriter
				if bfw, err = snkio.WriterBuffer(bfa...); err != nil {
					return
				}
				var bfr, _ = bfw.Reader()
				f = snkfs.NewFile(snkfs.StatFileFunc(func() (snkfs.FileInfo, error) {
					return rtfi, nil
				}), bfr, bfr, snkio.CloseFunc(func() error {
					bfw.Close()
					bfr.Close()
					return nil
				}))*/
			}
		}
	}

	var tmplsrc, tmpltk = Template[P, M](nil), false

	if cancache {
		tmplsrc, tmpltk = templates.Load(rtfi.Path())
		if tmpltk && al == 0 {
			if tmplsrc.Valid(fstat) {
				return tmplsrc.Eval(evalcntnt, evalprgmormod)
			}
			templates.Delete(rtfi.Path())
		}
		if cancache {
			if f, err = fsopen.Open(rtfi.Path()); f == nil {
				return
			}
		}
	}
	if f != nil {
		defer f.Close()
		if al >= 0 {
			a = append([]any{f}, a...)
		}
	}
	if rtfi != nil {
		if tmplsrc, err = LoadTempate(templates, cancache, compile, rtfi, fstat, fsopen, a...); err == nil {
			if !cancache {
				result, err = tmplsrc.Eval(evalcntnt, evalprgmormod)
				tmplsrc.Close()
				return
			}
			return tmplsrc.Eval(evalcntnt, evalprgmormod)
		}
	}
	return
}

func LoadTempate[P any, M any](templates snkio.Map[string, Template[P, M]], storetemplate bool, compile CompileCodeFunc[P, M], rtfi snkfs.FileInfo, fstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, a ...any) (tmpltsrc Template[P, M], err error) {
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
			if prgm, modrec, err = compile(cde, rtfi, dependants); err == nil {
				tmpltsrc = &prgrmtmplt[P, M]{unmatched: unmatched, matched: matched, dependants: dependants, cntnt: cntnt, pgrm: prgm, modrec: modrec, FileInfo: snkfs.GenFileInfo(rtfi.Path(), rtfi.Base(), rtfi.Size(), rtfi.IsDir(), rtfi.ModTime())}
				if storetemplate {
					templates.Store(rtfi.Path(), tmpltsrc)
				}
				return
			}
			cntnt.Close()
			return
		}
		tmpltsrc = &prgrmtmplt[P, M]{unmatched: unmatched, matched: matched, dependants: nil, cntnt: cntnt, pgrm: prgm, modrec: modrec, FileInfo: snkfs.GenFileInfo(rtfi.Path(), rtfi.Base(), rtfi.Size(), rtfi.IsDir(), rtfi.ModTime())}
		if storetemplate {
			templates.Store(rtfi.Path(), tmpltsrc)
		}
		return
	}
	cntnt.Close()
	return
}

type ResolveTemplatModule[P any, M any] func(specifier string, fndspecifierpath func(string, time.Time), compile CompileCodeFunc[P, M], rootfi fs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error)
