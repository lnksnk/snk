package fs

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"time"

	"github.com/lnksnk/snk/fs"
	snkfs "github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
	snksqlfs "github.com/lnksnk/snk/sql/fs"
)

type FSTHandler[R any, P any, M any, O any] interface {
	FSTFormatQuery[R]
	FSTProgramModules[P, M]
	FSTCompileProgram[R, P, M, O]
	FSTResolveHostModule[R, P, M, O]
	FSTExecDependants[R, P, M, O]
	FSTExecDependant[R, P, M, O]
	FSTExecModule[R, P, M, O]
	FSTExecProgram[R, P, M, O]
	FSTRequire[R, P, M, O]
	FSTEval[R, P, M, O]
	FSTSetVMArgs[R, P, M, O]
}

type FSTHANDLER[R any, P any, M any, O any] struct {
	prgmtmplts snkio.Map[string, Template[P, M]]
	vmpool     snkio.Pool[R]
	FSTCompileProgram[R, P, M, O]
	FSTExecDependant[R, P, M, O]
	FSTExecModule[R, P, M, O]
	FSTExecProgram[R, P, M, O]
	FSTRequire[R, P, M, O]
	FSTSetVMArgs[R, P, M, O]
}

func FSTPooledSyncHandler[R any, P any, M any, O any](newvm snkio.PoolNewFunc[R], cleanupvm snkio.PoolCleanupFunc[R], setupfsthander func(*FSTHANDLER[R, P, M, O])) FSTHandler[R, P, M, O] {
	var sfthndlr *FSTHANDLER[R, P, M, O]

	sfthndlr = &FSTHANDLER[R, P, M, O]{vmpool: snkio.NewPool(newvm, cleanupvm), prgmtmplts: snkio.NewSyncMap[string, Template[P, M]](snkio.MapDeletedFunc[string, Template[P, M]](func(key string, value ...Template[P, M]) {
		if len(value) >= 1 {
			if clsed, _ := value[0].(io.Closer); clsed != nil {
				clsed.Close()
			}
		}
	}))}

	if setupfsthander != nil {
		setupfsthander(sfthndlr)
	}
	return sfthndlr
}

func FSTPooledHandler[R any, P any, M any, O any](newvm snkio.PoolNewFunc[R], cleanupvm snkio.PoolCleanupFunc[R], setupfsthander func(*FSTHANDLER[R, P, M, O])) FSTHandler[R, P, M, O] {
	var sfthndlr *FSTHANDLER[R, P, M, O]
	sfthndlr = &FSTHANDLER[R, P, M, O]{vmpool: snkio.NewPool(newvm, cleanupvm), prgmtmplts: snkio.NewMap[string, Template[P, M]](snkio.MapDeletedFunc[string, Template[P, M]](func(key string, value ...Template[P, M]) {
		if len(value) >= 1 {
			if clsed, _ := value[0].(io.Closer); clsed != nil {
				clsed.Close()
			}
		}
	}))}

	if setupfsthander != nil {
		setupfsthander(sfthndlr)
	}
	return sfthndlr
}

// ProgramModules implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ProgramModules() snkio.Map[string, Template[P, M]] {
	return f.prgmtmplts
}

// CompileProgram implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) CompileProgram(cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm P, modrec M, err error) {
	return f.FSTCompileProgram.CompileProgram(cde, rootfi, fsstat, fsopen, dependensies)
}

// Eval implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) Eval(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (err error) {
	if vmpool := f.vmpool; vmpool != nil {
		return Eval(f, rootfi, fsstat, fsopen, vm, w, vmpool.Get, vmpool.Put, vmsetup)
	}
	return Eval(f, rootfi, fsstat, fsopen, vm, w, nil, nil, vmsetup)
}

// ExecDependant implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ExecDependant(vm R, mr M) (err error) {
	return f.FSTExecDependant.ExecDependant(vm, mr)
}

// ExecDependants implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ExecDependants(vm R, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependants map[string]time.Time) (err error) {
	return ExeDependants(f, vm, rootfi, fsstat, fsopen, dependants)
}

func ExeDependants[R any, P any, M any, O any](h FSTHandler[R, P, M, O], vm R, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependants map[string]time.Time) (err error) {
	if !isReflectNil(reflect.ValueOf(vm)) {
		var depmd M
		for depths := range dependants {
			if depmd, err = h.ResolveHostModule(depths, nil, rootfi, fsstat, fsopen); err != nil {
				return
			}
			if err = h.ExecDependant(vm, depmd); err != nil {
				return
			}
		}
	}
	return
}

// ExecModule implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ExecModule(vm R, mr M) (o O, err error) {
	return f.FSTExecModule.ExecModule(vm, mr)
}

// ExecProgram implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ExecProgram(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (err error) {
	return f.FSTExecProgram.ExecProgram(vm, pgrm, rootfi, fsstat, fsopen, canrequire)
}

// Require implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) Require(vm R, specifier string, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (o O, err error) {
	return Require(f, vm, specifier, rootfi, fsstat, fsopen)
}

func Eval[R any, P any, M any, O any](h FSTHandler[R, P, M, O], rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, invokevm func() R, disposevm func(R), vmsetup FSTSetupVM[R, P, M, O]) (err error) {
	if prgmtmplts, compile := h.ProgramModules(), h.CompileProgram; prgmtmplts != nil {
		var evalcntnt EvalContentFunc
		if w != nil {
			evalcntnt = func(cntnt snkio.BufferWriter) (wrterr error) {
				_, wrterr = cntnt.WriteTo(w)
				return
			}
		}
		return EvalTemplate(prgmtmplts, compile, rootfi, fsstat, fsopen, evalcntnt, func(prgm P, modrec M, dependants map[string]time.Time) (runerr error) {
			if isReflectNil(reflect.ValueOf(vm)) {
				if invokevm == nil {
					return
				}
				if vm = invokevm(); isReflectNil(reflect.ValueOf(vm)) {
					return
				}
				if disposevm != nil {
					defer disposevm(vm)
				}
			}
			var prvprint, prvprintln = SetVMprints(h, vm, w)
			if prvprint != nil && prvprintln != nil {
				defer func() {
					if vmset, _ := reflect.ValueOf(vm).Interface().(interface {
						Set(string, any)
					}); vmset != nil {
						vmset.Set("print", prvprint)
						vmset.Set("println", prvprintln)
					}
				}()
			}
			if vmsetup != nil {
				vmsetup(h, vm)
			}

			if !isReflectNil(reflect.ValueOf(modrec)) {
				_, err = h.ExecModule(vm, modrec)
				return
			}
			if err = h.ExecDependants(vm, rootfi, fsstat, fsopen, dependants); err == nil {
				return h.ExecProgram(vm, prgm, rootfi, fsstat, fsopen, len(dependants) == 0)
			}
			return
		})
	}
	err = fmt.Errorf("nil Program Module handler")
	return
}

func Compile[R any, P any, M any, O any](h FSTHandler[R, P, M, O], cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm P, modrec M, err error) {
	return h.CompileProgram(cde, rootfi, fsstat, fsopen, dependensies)
}

func Require[R any, P any, M any, O any](h FSTHandler[R, P, M, O], vm R, specifier string, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (o O, err error) {
	var mr M
	if mr, err = ResolveHostModule(h, specifier, nil, rootfi, fsstat, fsopen); err != nil {
		return
	}
	return h.ExecModule(vm, mr)
}

func ResolveHostModule[R any, P any, M any, O any](h FSTHandler[R, P, M, O], specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error) {
	if prgmtmplts := h.ProgramModules(); prgmtmplts != nil {
		if specifier == "" {
			mderr = fmt.Errorf("no module specifier")
			return
		}
		if fsstat == nil {
			mderr = fmt.Errorf("unable to load module %s", specifier)
			return
		}
		if specifier[0] != '/' {
			if rootfi == nil {
				specifier = "/" + specifier
			}
			if rootfi != nil {
				specifier = rootfi.Root() + specifier
			}
		}
		ext := filepath.Ext(specifier)
		if ext == "" {
			ext = ".js"
			spcl := len(specifier)
			if specifier[spcl-1] != '/' {
				specifier = specifier + ext
			}
			if specifier[spcl-1] == '/' {
				specifier = specifier + "index" + ext
			}
		}
		var mdfi snkfs.FileInfo
		if mdfi, mderr = fsstat.Stat(specifier); mderr != nil {
			return
		}
		if mdfi == nil {
			prgmtmplts.Delete(specifier)
			mderr = fmt.Errorf("%s module not found", specifier)
			return
		}
		mderr = EvalTemplate(prgmtmplts, h.CompileProgram, mdfi, fsstat, fsopen, nil, func(prgm P, modrec M, dependants map[string]time.Time) (runerr error) {
			if md = modrec; !isReflectNil(reflect.ValueOf(md)) {
				if fndspecifierpath != nil {
					fndspecifierpath(mdfi.Path(), mdfi.ModTime())
				}
				return
			}
			return fmt.Errorf("invalid module code %s", specifier)
		})
		return
	}
	mderr = fmt.Errorf("nil Program Module handler")
	return
}

// ResolveHostModule implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ResolveHostModule(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error) {
	return ResolveHostModule(f, specifier, fndspecifierpath, rootfi, fsstat, fsopen)
}

// SetVMArgs implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) SetVMArgs(h FSTHandler[R, P, M, O], vm R, args ...VMItem) {
	f.FSTSetVMArgs.SetVMArgs(f, vm, args...)
}

// FormatQuery implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) FormatQuery(vm R, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, out io.Writer, name, driver, query string, a ...any) error {
	return FSTFormatSql(f, vm, fsstat, fsopen, out, name, driver, query)
}

type VMItem []any

func (vi VMItem) Key() (key string) {
	if len(vi) >= 1 {
		key, _ = vi[0].(string)
	}
	return
}

func (vi VMItem) Value() (val any) {
	if len(vi) >= 2 {
		return vi[1]
	}
	return
}

type FSTInvokeVM[R any] func(vm R) R

type FSTSetupVM[R any, P any, M any, O any] func(h FSTHandler[R, P, M, O], vm R)

type FSTDisposeVM[R any] func(vm R)

type FSTVMPrint func(...any) error
type FSTVMPrintln func(...any) error

type FSTSetVMArgs[R any, P any, M any, O any] interface {
	SetVMArgs(h FSTHandler[R, P, M, O], vm R, args ...VMItem)
}

type FSTSetVMArgsFunc[R any, P any, M any, O any] func(h FSTHandler[R, P, M, O], vm R, args ...VMItem)

func (fststargs FSTSetVMArgsFunc[R, P, M, O]) SetVMArgs(h FSTHandler[R, P, M, O], vm R, args ...VMItem) {
	fststargs(h, vm, args...)
}

type FSTEval[R any, P any, M any, O any] interface {
	Eval(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (err error)
}

type FSTEvalFunc[R any, P any, M any, O any] func(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (err error)

func (fstevlfn FSTEvalFunc[R, P, M, O]) Eval(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (err error) {
	return fstevlfn(rootfi, fsstat, fsopen, vm, w, vmsetup)
}

type FSTProgramModules[P any, M any] interface {
	ProgramModules() snkio.Map[string, Template[P, M]]
}

type FSTProgramModulesFunc[P any, M any] func() snkio.Map[string, Template[P, M]]

func (fstpgrmmdsfnc FSTProgramModulesFunc[P, M]) ProgramModules() snkio.Map[string, Template[P, M]] {
	var pgrmmod snkio.Map[string, Template[P, M]] = fstpgrmmdsfnc()
	return pgrmmod
}

type FSTCompileProgram[R any, P any, M any, O any] interface {
	CompileProgram(cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm P, modrec M, err error)
}

type FSTCompileProgramFunc[R any, P any, M any, O any] func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm P, modrec M, err error)

func (fstcmplfnc FSTCompileProgramFunc[R, P, M, O]) CompileProgram(cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm P, modrec M, err error) {
	return fstcmplfnc(cde, rootfi, fsstat, fsopen, dependensies)
}

type FSTResolveHostModule[R any, P any, M any, O any] interface {
	ResolveHostModule(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error)
}

type FSTResolveHostModuleFunc[R any, P any, M any, O any] func(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error)

func (fstrslvmdfnc FSTResolveHostModuleFunc[R, P, M, O]) ResolveHostModule(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error) {
	return fstrslvmdfnc(specifier, fndspecifierpath, rootfi, fsstat, fsopen)
}

type FSTExecDependants[R any, P any, M any, O any] interface {
	ExecDependants(vm R, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependants map[string]time.Time) (err error)
}

type FSTExecDependantsFunc[R any, P any, M any, O any] func(vm R, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependants map[string]time.Time) (err error)

func (fstdpndntsfnc FSTExecDependantsFunc[R, P, M, O]) ExecDependants(h FSTHandler[R, P, M, O], vm R, compile CompileCodeFunc[P, M], rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependants map[string]time.Time) (err error) {
	return fstdpndntsfnc(vm, rootfi, fsstat, fsopen, dependants)
}

type FSTExecDependant[R any, P any, M any, O any] interface {
	ExecDependant(vm R, mr M) (err error)
}

type FSTExecDependantFunc[R any, P any, M any, O any] func(vm R, mr M) (err error)

func (fstpndntfnc FSTExecDependantFunc[R, P, M, O]) ExecDependant(vm R, mr M) (err error) {
	return fstpndntfnc(vm, mr)
}

type FSTExecModule[R any, P any, M any, O any] interface {
	ExecModule(vm R, mr M) (o O, err error)
}

type FSTExecModuleFunc[R any, P any, M any, O any] func(vm R, mr M) (o O, err error)

func (fstmdfnc FSTExecModuleFunc[R, P, M, O]) ExecModule(vm R, mr M) (o O, err error) {
	return fstmdfnc(vm, mr)
}

type FSTExecProgram[R any, P any, M any, O any] interface {
	ExecProgram(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (err error)
}

type FSTExecProgramFunc[R any, P any, M any, O any] func(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (err error)

func (fsthpgrmfnc FSTExecProgramFunc[R, P, M, O]) ExecProgram(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (err error) {
	return fsthpgrmfnc(vm, pgrm, rootfi, fsstat, fsopen, canrequire)
}

type FSTRequire[R any, P any, M any, O any] interface {
	Require(vm R, specifier string, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (o O, err error)
}

type FSTRequireFunc[R any, P any, M any, O any] func(vm R, specifier string, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (o O, err error)

func (fstrqrfnc FSTRequireFunc[R, P, M, O]) Require(vm R, specifier string, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (o O, err error) {
	return fstrqrfnc(vm, specifier, rootfi, fsstat, fsopen)
}

var emptyprint = func(...any) error {
	return nil
}

func SetVMprints[R any, P any, M any, O any](h FSTHandler[R, P, M, O], vm R, w io.Writer) (prvprint func(...any) error, prvprintln func(...any) error) {
	if h != nil {
		if w == nil {
			h.SetVMArgs(h, vm, VMItem{"print", emptyprint}, VMItem{"println", emptyprint})
			return
		}
		var vmval = reflect.ValueOf(vm)
		var vmget = vmval.MethodByName("Get")
		var vmgetfunc = func(name string) any {
			if vmget.Kind() == reflect.Func {
				var args = []reflect.Value{reflect.ValueOf(name)}
				var rslt = vmget.Call(args)
				if len(rslt) == 1 {
					return rslt[0].Interface()
				}
			}
			return nil
		}

		var rprint, rprintln = vmgetfunc("print"), vmgetfunc("println")
		if rprint != nil {
			prvprint, _ = rprint.(func(...any) error)
		}
		if rprintln != nil {
			prvprintln, _ = rprintln.(func(...any) error)
		}

		h.SetVMArgs(h, vm, VMItem{"print", func(a ...any) (wrterr error) {
			_, wrterr = snkio.Fprint(w, a...)
			return
		}}, VMItem{"println", func(a ...any) (wrterr error) {
			_, wrterr = snkio.Fprintln(w, a...)
			return
		}})
	}
	return
}

type FTSSetupRequestVM[R any] func(vm R, r *http.Request, vmsetup func(R, ...VMItem))

func ServeHTTP[R any, P any, M any, O any](h FSTHandler[R, P, M, O], w http.ResponseWriter, r *http.Request, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, setupRequestVM FTSSetupRequestVM[R]) {
	var vm R
	if err := h.Eval(rootfi, fsstat, fsopen, vm, w, func(h FSTHandler[R, P, M, O], vm R) {
		if setupRequestVM != nil {
			var vmsetup = func(vm R, args ...VMItem) {
				h.SetVMArgs(h, vm, args...)
			}
			setupRequestVM(vm, r, vmsetup)
		}
	}); err != nil {
		snkio.Fprint(w, err.Error())
	}
}

type RVm interface {
	Set(string, any)
	Get(string) RValue
}

type RValue interface {
	ToInteger() int64
	//ToString() RValue
	String() string
	ToFloat() float64
	ToNumber() RValue
	ToBoolean() bool
	//ToObject(*Runtime) *Object
	//SameAs(Value) bool
	//Equals(Value) bool
	//StrictEquals(Value) bool
	Export() interface{}
	ExportType() reflect.Type

	//baseObject(r *Runtime) *Object

	//hash(hasher *maphash.Hash) uint64
}

type FSTFormatQuery[R any] interface {
	FormatQuery(vm R, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, out io.Writer, name, driver, query string, a ...any) error
}

type FSTFormatQueryFunc[R any] func(vm R, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, out io.Writer, name, driver, query string, a ...any) error

func (fstfrmtqryfnc FSTFormatQueryFunc[R]) FormatQuery(vm R, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, out io.Writer, name, driver, query string, a ...any) error {
	return fstfrmtqryfnc(vm, fsstat, fsopen, out, name, driver, query, a...)
}

func FSTFormatSql[R any, P any, M any, O any](h FSTHandler[R, P, M, O], vm R, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, out io.Writer, name, driver, query string, a ...any) (err error) {
	return snksqlfs.FSFormatQuery(out, name, driver, query, fsstat, fsopen, func(unmatchedpath string) {
		if prgmtmplts := h.ProgramModules(); prgmtmplts != nil && prgmtmplts.Exist(unmatchedpath) {
			go prgmtmplts.Delete(unmatchedpath)
		}
	}, func(out io.Writer, sqlfi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem) (preperr error) {
		return h.Eval(sqlfi, fsstat, fsopen, vm, out, nil)
	})
}
