package main

import (
	"fmt"
	"time"

	"github.com/lnksnk/snk/es"
)

type programmodrec struct {
	hostResolveImportedModule es.HostResolveImportedModuleFunc
	rqstdmods                 []string
	pgrm                      *es.Program
	exportedNames             []string
	exportedNamesCallbacks    []func([]string)
}

// Instantiate implements [es.CyclicModuleRecord].
func (p *programmodrec) Instantiate(rt *es.Runtime) (es.CyclicModuleInstance, error) {
	return nil, fmt.Errorf("should not be called")
}

// Evaluate implements [ModuleRecord].
func (p *programmodrec) Evaluate(r *es.Runtime) *es.Promise {
	prms, res, rej := r.NewPromise()
	var err error
	if rsvimp := p.hostResolveImportedModule; rsvimp != nil {
		var mrc es.ModuleRecord
		var mp *es.Promise
		for rsvm := range p.rqstdmods {
			if mrc, err = rsvimp(p, p.rqstdmods[rsvm]); err != nil {
				rej(err)
				return prms
			}
			mp = mrc.Evaluate(r)
			for mp.State() == es.PromiseStatePending {
				time.Sleep(time.Nanosecond * 1)
			}
			if mp.State() == es.PromiseStateRejected {
				res(mp.Result().(error))
				return prms
			}
			if mp.State() == es.PromiseStateFulfilled {
				var expnmes []string
				mrc.GetExportedNames(func(s []string) {
					expnmes = append(expnmes, s...)
				})
				var no = r.NamespaceObjectFor(mrc)
				for ei := range expnmes {
					r.Set(expnmes[ei], no.Get(expnmes[ei]))
				}
				continue
			}
		}
	}
	var result es.Value
	if result, err = r.RunProgram(p.pgrm); err != nil {
		rej(err)
		return prms
	}
	res(result)
	return prms
}

func (p *programmodrec) RequestedModules() []string { return p.rqstdmods }

func (p *programmodrec) InitializeEnvironment() error { return nil }

// GetExportedNames implements [ModuleRecord].
func (p *programmodrec) GetExportedNames(callback func([]string), resolveset ...es.ModuleRecord) bool {
	if p.exportedNames != nil {
		callback(p.exportedNames)
		return true
	}
	p.exportedNamesCallbacks = append(p.exportedNamesCallbacks, callback)
	return false
}

// Link implements [ModuleRecord].
func (p *programmodrec) Link() error {
	return nil
}

// ResolveExport implements [ModuleRecord].
func (p *programmodrec) ResolveExport(exportName string, _ ...es.ResolveSetElement) (*es.ResolvedBinding, bool) {
	return &es.ResolvedBinding{
		Module:      p,
		BindingName: exportName,
	}, false
}

type programcyclemodinstance struct {
	p                *programmodrec
	exports          *es.Object
	isEsModuleMarked bool
}
