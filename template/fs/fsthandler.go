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
	FSTCompileProgram[R, P, M, O]
	FSTResolveHostModule[R, P, M, O]
	FSTExecProgram[R, P, M, O]
	FSTEval[R, P, M, O]
	REval[R]
}

type FSTHANDLER[R any, P any, M any, O any] struct {
	fss           fs.FileSystem
	prgmtmplts    snkio.Map[string, Template[P, M]]
	vmpool        snkio.Pool[R]
	execModule    ExecModuleFunc[R, M, O]
	execDependant ExecDependantFunc[R, M]
	FSTCompileProgram[R, P, M, O]
	FSTExecProgram[R, P, M, O]
}

func FSTPooledSyncHandler[R any, P any, M any, O any](newvm snkio.PoolNewFunc[R], cleanupvm snkio.PoolCleanupFunc[R], setupfsthander func(*FSTHANDLER[R, P, M, O]), fss fs.FileSystem, execModule ExecModuleFunc[R, M, O], execDependant ExecDependantFunc[R, M]) FSTHandler[R, P, M, O] {
	if execModule == nil {
		execModule = func(vm R, mr M) (o O, err error) {
			err = fmt.Errorf("excuting module not supported")
			return
		}
	}

	if execDependant == nil {
		execDependant = func(vm R, mr M) (err error) {
			err = fmt.Errorf("excuting dependant module not supported")
			return
		}
	}

	var sfthndlr *FSTHANDLER[R, P, M, O]

	sfthndlr = &FSTHANDLER[R, P, M, O]{execModule: execModule, execDependant: execDependant, fss: fss, vmpool: snkio.NewPool(newvm, cleanupvm), prgmtmplts: snkio.NewSyncMap[string, Template[P, M]](snkio.MapDeletedFunc[string, Template[P, M]](func(key string, value ...Template[P, M]) {
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

func FSTPooledHandler[R any, P any, M any, O any](newvm snkio.PoolNewFunc[R], cleanupvm snkio.PoolCleanupFunc[R], setupfsthander func(*FSTHANDLER[R, P, M, O]), fss fs.FileSystem, execDependant ExecDependantFunc[R, M], execModule ExecModuleFunc[R, M, O]) FSTHandler[R, P, M, O] {

	if execModule == nil {
		execModule = func(vm R, mr M) (o O, err error) {
			err = fmt.Errorf("excuting module not supported")
			return
		}
	}

	if execDependant == nil {
		execDependant = func(vm R, mr M) (err error) {
			err = fmt.Errorf("excuting dependant module not supported")
			return
		}
	}

	var sfthndlr *FSTHANDLER[R, P, M, O]
	sfthndlr = &FSTHANDLER[R, P, M, O]{execModule: execModule, execDependant: execDependant, fss: fss, vmpool: snkio.NewPool(newvm, cleanupvm), prgmtmplts: snkio.NewMap[string, Template[P, M]](snkio.MapDeletedFunc[string, Template[P, M]](func(key string, value ...Template[P, M]) {
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

func (f *FSTHANDLER[R, P, M, O]) EvalR(vm R, a ...any) (result any, err error) {
	return
}

// CompileProgram implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) CompileProgram(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error) {
	return f.FSTCompileProgram.CompileProgram(cde, rootfi, dependensies)
}

// Eval implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) Eval(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (result any, err error) {
	if vmpool := f.vmpool; vmpool != nil {
		return Eval(f.prgmtmplts, f.CompileProgram, rootfi, fsstat, fsopen, vm, w, vmpool.Get, vmpool.Put, vmsetup, f.execDependant, f.execModule, f.ExecProgram)
	}
	return Eval(f.prgmtmplts, f.CompileProgram, rootfi, fsstat, fsopen, vm, w, nil, nil, vmsetup, f.execDependant, f.execModule, f.ExecProgram)
}

func ExeDependants[R any, P any, M any, O any](prgmtmplts snkio.Map[string, Template[P, M]], compile func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error), execDependant ExecDependantFunc[R, M], vm R, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependants map[string]time.Time) (err error) {
	if !isReflectNil(reflect.ValueOf(vm)) {
		var depmd M
		for depths := range dependants {
			if depmd, err = ResolveHostModule[R, P, M, O](prgmtmplts, compile, depths, nil, rootfi, fsstat, fsopen); err != nil {
				return
			}
			if err = execDependant(vm, depmd); err != nil {
				return
			}
		}
	}
	return
}

// ExecModule implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ExecModule(vm R, mr M) (o O, err error) {
	if execModule := f.execModule; execModule != nil {
		return execModule(vm, mr)
	}
	return
}

// ExecProgram implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ExecProgram(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (result any, err error) {
	return f.FSTExecProgram.ExecProgram(vm, pgrm, rootfi, fsstat, fsopen, canrequire)
}

// Require implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) Require(vm R, specifier string, rootfi snkfs.FileInfo) (o O, err error) {
	return Require(f.prgmtmplts, vm, f.CompileProgram, f.ExecModule, specifier, rootfi, f.fss, f.fss)
}

func EvalR[R any, P any, M any, O any](prgmtmplts snkio.Map[string, Template[P, M]], compile func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, execDependant ExecDependantFunc[R, M], execModule ExecModuleFunc[R, M, O], execProgram ExecProgramFunc[R, P], a ...any) (result any, err error) {
	if prgmtmplts != nil && compile != nil {
		var evalcntnt EvalContentFunc
		if w != nil {
			evalcntnt = func(cntnt snkio.BufferWriter) (wrterr error) {
				_, wrterr = cntnt.WriteTo(w)
				return
			}
		}
		return EvalTemplate(prgmtmplts, compile, rootfi, ".html", fsstat, fsopen, evalcntnt, func(prgm P, modrec M, dependants map[string]time.Time) (runresult any, runerr error) {
			if isReflectNil(reflect.ValueOf(vm)) {
				return
			}

			if !isReflectNil(reflect.ValueOf(modrec)) {
				_, err = execModule(vm, modrec)
				return
			}
			var vmr RApi[R] = &rapi[R]{vm: vm}
			var prvprint, prvprintln = SetVMprints[R](vmr, w)
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
			if !isReflectNil(reflect.ValueOf(modrec)) {
				_, err = execModule(vm, modrec)
				return
			}
			if err = ExeDependants[R, P, M, O](prgmtmplts, compile, execDependant, vm, rootfi, fsstat, fsopen, dependants); err == nil {
				return execProgram(vm, prgm, rootfi, fsstat, fsopen, len(dependants) == 0)
			}
			return
		}, a...)
	}
	err = fmt.Errorf("nil Program Module handler")
	return
}

func Eval[R any, P any, M any, O any](prgmtmplts snkio.Map[string, Template[P, M]], compile func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, invokevm func() R, disposevm func(R), vmsetup FSTSetupVM[R, P, M, O], execDependant ExecDependantFunc[R, M], execModule ExecModuleFunc[R, M, O], execProgram ExecProgramFunc[R, P]) (result any, err error) {
	if prgmtmplts != nil && compile != nil {
		var evalcntnt EvalContentFunc
		if w != nil {
			evalcntnt = func(cntnt snkio.BufferWriter) (wrterr error) {
				_, wrterr = cntnt.WriteTo(w)
				return
			}
		}
		return EvalTemplate(prgmtmplts, compile, rootfi, ".html", fsstat, fsopen, evalcntnt, func(prgm P, modrec M, dependants map[string]time.Time) (runresult any, runerr error) {
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
			var vmr RApi[R] = &rapi[R]{vm: vm}
			vmr.Set("eval", func(a ...any) (result any, err error) {
				return EvalR(prgmtmplts, compile, rootfi, fsstat, fsopen, vm, w, execDependant, execModule, execProgram, a...)
			})
			var prvprint, prvprintln = SetVMprints[R](vmr, w)
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
				vmsetup(vmr)
			}

			if !isReflectNil(reflect.ValueOf(modrec)) {
				_, err = execModule(vm, modrec)
				return
			}
			if err = ExeDependants[R, P, M, O](prgmtmplts, compile, execDependant, vm, rootfi, fsstat, fsopen, dependants); err == nil {
				return execProgram(vm, prgm, rootfi, fsstat, fsopen, len(dependants) == 0)
			}
			return
		})
	}
	err = fmt.Errorf("nil Program Module handler")
	return
}

func Compile[R any, P any, M any, O any](h FSTHandler[R, P, M, O], cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error) {
	return h.CompileProgram(cde, rootfi, dependensies)
}

func Require[R any, P any, M any, O any](prgmtmplts snkio.Map[string, Template[P, M]], vm R, compile func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error), execModule func(vm R, mr M) (o O, err error), specifier string, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (o O, err error) {
	var mr M
	if mr, err = ResolveHostModule[R, P, M, O](prgmtmplts, compile, specifier, nil, rootfi, fsstat, fsopen); err != nil {
		return
	}
	return execModule(vm, mr)
}

func ResolveHostModule[R any, P any, M any, O any](prgmtmplts snkio.Map[string, Template[P, M]], compile func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error), specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem) (md M, mderr error) {
	if prgmtmplts != nil {
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
		_, mderr = EvalTemplate(prgmtmplts, compile, mdfi, ".js", fsstat, fsopen, nil, func(prgm P, modrec M, dependants map[string]time.Time) (runresult any, runerr error) {
			if md = modrec; !isReflectNil(reflect.ValueOf(md)) {
				if fndspecifierpath != nil {
					fndspecifierpath(mdfi.Path(), mdfi.ModTime())
				}
				return
			}
			runerr = fmt.Errorf("invalid module code %s", specifier)
			return
		})
		return
	}
	mderr = fmt.Errorf("nil Program Module handler")
	return
}

// ResolveHostModule implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) ResolveHostModule(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo) (md M, mderr error) {
	return ResolveHostModule[R, P, M, O](f.prgmtmplts, f.CompileProgram, specifier, fndspecifierpath, rootfi, f.fss, f.fss)
}

// SetVMArgs implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) SetVMArgs(vmr RApi[R], args ...VMItem) {
	SetVMArgs[R](vmr, args...)
}

// FormatQuery implements [FSTHandler].
func (f *FSTHANDLER[R, P, M, O]) FormatQuery(vm R, out io.Writer, name, driver, query string, a ...any) error {
	return FSTFormatSql(f.prgmtmplts, f.Eval, vm, f.fss, f.fss, out, name, driver, query)
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

type FSTSetupVM[R any, P any, M any, O any] func(vmr RApi[R])

type FSTDisposeVM[R any] func(vm R)

type FSTVMPrint func(...any) error
type FSTVMPrintln func(...any) error

type FSTEval[R any, P any, M any, O any] interface {
	Eval(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (result any, err error)
}

type FSTEvalFunc[R any, P any, M any, O any] func(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (result any, err error)

func (fstevlfn FSTEvalFunc[R, P, M, O]) Eval(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (result any, err error) {
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
	CompileProgram(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error)
}

type FSTCompileProgramFunc[R any, P any, M any, O any] func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error)

func (fstcmplfnc FSTCompileProgramFunc[R, P, M, O]) CompileProgram(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm P, modrec M, err error) {
	return fstcmplfnc(cde, rootfi, dependensies)
}

type ResolveHostModuleFunc[M any] func(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo) (md M, mderr error)

type FSTResolveHostModule[R any, P any, M any, O any] interface {
	ResolveHostModule(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo) (md M, mderr error)
}

type FSTResolveHostModuleFunc[R any, P any, M any, O any] func(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo) (md M, mderr error)

func (fstrslvmdfnc FSTResolveHostModuleFunc[R, P, M, O]) ResolveHostModule(specifier string, fndspecifierpath func(string, time.Time), rootfi snkfs.FileInfo) (md M, mderr error) {
	return fstrslvmdfnc(specifier, fndspecifierpath, rootfi)
}

type ExecDependantFunc[R any, M any] func(vm R, mr M) (err error)

type ExecModuleFunc[R any, M any, O any] func(vm R, mr M) (o O, err error)

type FSTExecProgram[R any, P any, M any, O any] interface {
	ExecProgram(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (result any, err error)
}

type ExecProgramFunc[R any, P any] func(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (result any, err error)

type FSTExecProgramFunc[R any, P any, M any, O any] func(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (result any, err error)

func (fsthpgrmfnc FSTExecProgramFunc[R, P, M, O]) ExecProgram(vm R, pgrm P, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (result any, err error) {
	return fsthpgrmfnc(vm, pgrm, rootfi, fsstat, fsopen, canrequire)
}

var emptyprint = func(...any) error {
	return nil
}

func SetVMArgs[R any](vmr RApi[R], args ...VMItem) {
	if vmr != nil {
		for ai := range args {
			vmr.Set(args[ai].Key(), args[ai].Value())
		}
	}
}

func SetVMEvalR[R any, P any, M any, O any](h FSTHandler[R, P, M, O], vm R, w io.Writer) {

}

func SetVMprints[R any](vmr RApi[R], w io.Writer) (prvprint func(...any) error, prvprintln func(...any) error) {
	if vmr != nil {
		if w == nil {
			vmr.Set("print", emptyprint)
			vmr.Set("println", emptyprint)
			return
		}

		var rprint, rprintln = vmr.Get("print"), vmr.Get("println")
		if rprint != nil {
			prvprint, _ = rprint.(func(...any) error)
		}
		if rprintln != nil {
			prvprintln, _ = rprintln.(func(...any) error)
		}
		vmr.Set("print", func(a ...any) (wrterr error) {
			_, wrterr = snkio.Fprint(w, a...)
			return
		})
		vmr.Set("println", func(a ...any) (wrterr error) {
			_, wrterr = snkio.Fprintln(w, a...)
			return
		})
	}
	return
}

type FTSSetupRequestVM[R any] func(vm R, vmr RApi[R], r *http.Request, vmsetup func(RApi[R], ...VMItem))

func ServeHTTP[R any, P any, M any, O any](h FSTHandler[R, P, M, O], w http.ResponseWriter, r *http.Request, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, setupRequestVM FTSSetupRequestVM[R]) {
	var vm R
	if _, err := h.Eval(rootfi, fsstat, fsopen, vm, w, func(vmr RApi[R]) {
		if setupRequestVM != nil {
			setupRequestVM(vm, vmr, r, SetVMArgs[R])
		}
	}); err != nil {
		snkio.Fprint(w, err.Error())
	}
}

type RApi[R any] interface {
	Set(string, any)
	Get(string) any
}

type rapi[R any] struct {
	vm      R
	getfunc func(name string) any
	setfunc func(name string, val any)
}

func (vr *rapi[R]) Set(key string, val any) {
	if vmval, set := reflect.ValueOf(vr.vm), vr.setfunc; !isReflectNil(vmval) {
	callset:
		if set != nil {
			set(key, val)
			return
		}
		if vmset := vmval.MethodByName("Set"); !isReflectNil(vmset) {
			if vmtype := vmset.Type(); vmtype.NumIn() == 2 {
				set = func(name string, val any) {
					vmset.Call([]reflect.Value{reflect.ValueOf(name), reflect.ValueOf(val)})
				}
				vr.setfunc = set
				goto callset
			}
		}
	}
}

func (vr *rapi[R]) Get(key string) (val any) {
	if vmval, get := reflect.ValueOf(vr.vm), vr.getfunc; !isReflectNil(vmval) {
	callget:
		if get != nil {
			val = get(key)
			return
		}
		if vmget := vmval.MethodByName("Get"); !isReflectNil(vmget) {
			if vmtype := vmget.Type(); vmtype.NumIn() == 1 && vmtype.NumOut() == 1 {
				get = func(name string) (val any) {
					rslt := vmget.Call([]reflect.Value{reflect.ValueOf(name)})
					if isReflectNil(rslt[0]) {
						return nil
					}
					val = rslt[0].Interface()
					return
				}
				vr.getfunc = get
				goto callget
			}
		}
	}
	return
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
	FormatQuery(vm R, out io.Writer, name, driver, query string, a ...any) error
}

type FSTFormatQueryFunc[R any] func(vm R, out io.Writer, name, driver, query string, a ...any) error

func (fstfrmtqryfnc FSTFormatQueryFunc[R]) FormatQuery(vm R, out io.Writer, name, driver, query string, a ...any) error {
	return fstfrmtqryfnc(vm, out, name, driver, query, a...)
}

func FSTFormatSql[R any, P any, M any, O any](prgmtmplts snkio.Map[string, Template[P, M]], eval func(rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, vm R, w io.Writer, vmsetup FSTSetupVM[R, P, M, O]) (result any, err error), vm R, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, out io.Writer, name, driver, query string, a ...any) (err error) {
	return snksqlfs.FSFormatQuery(out, name, driver, query, fsstat, fsopen, func(unmatchedpath string) {
		if prgmtmplts != nil && prgmtmplts.Exist(unmatchedpath) {
			go prgmtmplts.Delete(unmatchedpath)
		}
	}, func(out io.Writer, sqlfi fs.FileInfo, fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem) (preperr error) {
		_, preperr = eval(sqlfi, fsstat, fsopen, vm, out, nil)
		return
	})
}

type REval[R any] interface {
	EvalR(vm R, a ...any) (result any, err error)
}

type REvalFunc[R any] func(vm R, a ...any) (result any, err error)

func (revlfnc REvalFunc[R]) EvalR(vm R, a ...any) (result any, err error) {
	return revlfnc(vm, a...)
}
