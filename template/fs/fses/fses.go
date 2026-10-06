package fses

import (
	"time"

	"github.com/lnksnk/snk/es"
	"github.com/lnksnk/snk/es/ast"
	"github.com/lnksnk/snk/es/parser"
	snkfs "github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
	fst "github.com/lnksnk/snk/template/fs"
)

func ESHandler() fst.FSTHandler[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object] {
	return fst.FSTPooledSyncHandler(func() (r *es.Runtime) {
		r = es.New()
		return r
	}, func(r *es.Runtime) *es.Runtime {
		r.ClearInterrupt()
		var gblobj = r.GlobalObject()
		if gblobj != nil {
			if keys := gblobj.Keys(); gblobj.Get("require") == nil {
				for ki := range keys {
					gblobj.Set(keys[ki], nil)
				}
				return r
			}
			r = es.New()
		}
		return r
	}, func(f *fst.FSTHANDLER[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object]) {
		f.FSTCompileProgram = fst.FSTCompileProgramFunc[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object](func(cde snkio.BufferWriter, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, dependensies map[string]time.Time) (prgm *es.Program, modrec es.CyclicModuleRecord, err error) {
			var prg *ast.Program

			if prg, err = es.Parse("", cde.String(), parser.IsModule); err != nil {
				return
			}
			if prgm, err = es.CompileAST(prg, false); err == nil {
				var rqstmods []string
				for _, imprt := range prg.ImportEntries {
					if _, err = f.ResolveHostModule(string(imprt.FromClause.ModuleSpecifier), func(deppath string, depmodtime time.Time) {
						dependensies[deppath] = depmodtime
					}, rootfi, fsstat, fsopen); err != nil {
						return
					}
					rqstmods = append(rqstmods, string(imprt.FromClause.ModuleSpecifier))
				}

				expl := len(prg.ExportEntries)
				if expl > 0 {
					if modrec, err = es.ModuleFromAST(prg, func(referencingScriptOrModule interface{}, specifier string) (md es.ModuleRecord, mderr error) {
						return f.ResolveHostModule(specifier, nil, rootfi, fsstat, fsopen)
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
		f.FSTExecDependant = fst.FSTExecDependantFunc[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object](execDependant)
		f.FSTExecModule = fst.FSTExecModuleFunc[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object](execModule)
		f.FSTExecProgram = fst.FSTExecProgramFunc[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object](func(vm *es.Runtime, pgrm *es.Program, rootfi snkfs.FileInfo, fsstat snkfs.StatFileSystem, fsopen snkfs.OpenFileSystem, canrequire bool) (err error) {
			if vm != nil && pgrm != nil {
				if canrequire {
					vm.Set("require", func(specifier string) (o *es.Object, err error) {
						return fst.Require(f, vm, specifier, rootfi, fsstat, fsopen)
					})
				}
				_, err = vm.RunProgram(pgrm)
			}
			return
		})
		f.FSTSetVMArgs = fst.FSTSetVMArgsFunc[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object](func(h fst.FSTHandler[*es.Runtime, *es.Program, es.CyclicModuleRecord, *es.Object], vm *es.Runtime, args ...fst.VMItem) {
			if vm != nil {
				for ai := range args {
					vm.Set(args[ai].Key(), args[ai].Value())
				}
			}
		})
	})

}

func execDependant(vm *es.Runtime, mr es.CyclicModuleRecord) (err error) {
	if vm != nil && mr != nil {
		promise := mr.Evaluate(vm)
		for promise.State() == es.PromiseStatePending {
			time.Sleep(time.Nanosecond * 1)
		}
		if promise.State() == es.PromiseStateFulfilled {
			var o = vm.NamespaceObjectFor(mr)
			mr.GetExportedNames(func(s []string) {
				for si := range s {
					vm.Set(s[si], o.Get(s[si]))
				}
			})

			return err
		}
		if promise.State() == es.PromiseStateRejected {
			err = promise.Result().Export().(error)
		}
	}
	return
}

func execModule(vm *es.Runtime, mr es.CyclicModuleRecord) (o *es.Object, err error) {
	if vm != nil && mr != nil {
		promise := mr.Evaluate(vm)
		for promise.State() == es.PromiseStatePending {
			time.Sleep(time.Nanosecond * 1)
		}
		if promise.State() == es.PromiseStateFulfilled {
			o = vm.NamespaceObjectFor(mr)
			return o, err
		}
		if promise.State() == es.PromiseStateRejected {
			err = promise.Result().Export().(error)
		}
	}
	return
}
