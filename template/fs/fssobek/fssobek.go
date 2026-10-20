package fssobek

import (
	"time"

	snkfs "github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
	fst "github.com/lnksnk/snk/template/fs"

	"github.com/grafana/sobek"
	"github.com/grafana/sobek/ast"
	"github.com/grafana/sobek/parser"
)

func SobekHandler(fss snkfs.FileSystem) fst.FSTHandler[*sobek.Runtime, *sobek.Program, sobek.CyclicModuleRecord, *sobek.Object] {
	return fst.FSTPooledSyncHandler(func() (r *sobek.Runtime) {
		r = sobek.New()
		return r
	}, func(r *sobek.Runtime) *sobek.Runtime {
		r.ClearInterrupt()
		var gblobj = r.GlobalObject()
		if gblobj != nil {
			if keys := gblobj.Keys(); gblobj.Get("require") == nil {
				for ki := range keys {
					gblobj.Set(keys[ki], nil)
				}
				return r
			}
			r = sobek.New()
		}
		return r
	}, func(f *fst.FSTHANDLER[*sobek.Runtime, *sobek.Program, sobek.CyclicModuleRecord, *sobek.Object]) {
		f.FSTCompileProgram = fst.FSTCompileProgramFunc[*sobek.Runtime, *sobek.Program, sobek.CyclicModuleRecord, *sobek.Object](func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, dependensies map[string]time.Time) (prgm *sobek.Program, modrec sobek.CyclicModuleRecord, err error) {
			var prg *ast.Program
			if prg, err = sobek.Parse("", cde.String(), parser.IsModule); err != nil {
				return
			}
			if prgm, err = sobek.CompileAST(prg, false); err == nil {
				var rqstmods []string
				for _, imprt := range prg.ImportEntries {
					if _, err = f.ResolveHostModule(string(imprt.FromClause.ModuleSpecifier), func(deppath string, depmodtime time.Time) {
						dependensies[deppath] = depmodtime
					}, rootfi); err != nil {
						return
					}
					rqstmods = append(rqstmods, string(imprt.FromClause.ModuleSpecifier))
				}

				expl := len(prg.ExportEntries)
				if expl > 0 {
					if modrec, err = sobek.ModuleFromAST(prg, func(referencingScriptOrModule interface{}, specifier string) (md sobek.ModuleRecord, mderr error) {
						return f.ResolveHostModule(specifier, nil, rootfi)
					}); err != nil {
						return
					}
					if err = modrec.Link(); err == nil {
						if err = modrec.InitializeEnvironment(); err == nil {
							return nil, modrec, err
						}
						return
					}
					return
				}
				return
			}
			return
		})
		f.FSTExecProgram = fst.FSTExecProgramFunc[*sobek.Runtime, *sobek.Program, sobek.CyclicModuleRecord, *sobek.Object](func(vm *sobek.Runtime, pgrm *sobek.Program, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (err error) {
			if vm != nil && pgrm != nil {
				if canrequire {
					vm.Set("require", func(specifier string) (o *sobek.Object, err error) {
						return fst.Require(f.ProgramModules(), vm, f.CompileProgram, f.ExecModule, specifier, rootfi, fsstat, fsopen)
					})
				}
				_, err = vm.RunProgram(pgrm)
			}
			return
		})
	}, fss, fst.ExecModuleFunc[*sobek.Runtime, sobek.CyclicModuleRecord, *sobek.Object](execModule), fst.ExecDependantFunc[*sobek.Runtime, sobek.CyclicModuleRecord](execDependant))

}

func execDependant(vm *sobek.Runtime, mr sobek.CyclicModuleRecord) (err error) {
	if vm != nil && mr != nil {
		promise := mr.Evaluate(vm)
		for promise.State() == sobek.PromiseStatePending {
			time.Sleep(time.Nanosecond * 1)
		}
		if promise.State() == sobek.PromiseStateFulfilled {
			var o = vm.NamespaceObjectFor(mr)
			mr.GetExportedNames(func(s []string) {
				for si := range s {
					vm.Set(s[si], o.Get(s[si]))
				}
			})

			return err
		}
		if promise.State() == sobek.PromiseStateRejected {
			err = promise.Result().Export().(error)
		}
	}
	return
}

func execModule(vm *sobek.Runtime, mr sobek.CyclicModuleRecord) (o *sobek.Object, err error) {
	if vm != nil && mr != nil {
		promise := mr.Evaluate(vm)
		for promise.State() == sobek.PromiseStatePending {
			time.Sleep(time.Nanosecond * 1)
		}
		if promise.State() == sobek.PromiseStateFulfilled {
			o = vm.NamespaceObjectFor(mr)
			return o, err
		}
		if promise.State() == sobek.PromiseStateRejected {
			err = promise.Result().Export().(error)
		}
	}
	return
}
