package io

import (
	"reflect"
	"sync"
)

type Pool[V any] interface {
	New() V
	Put(V)
	Get() V
}

type sncpool[V any] struct {
	pl     *sync.Pool
	clnup  PoolCleanupFunc[V]
	ovrflw chan V
	new    PoolNewFunc[V]
}

func (sp *sncpool[V]) Put(v V) {
	if pl, clnup := sp.pl, sp.clnup; pl != nil && clnup != nil {
		v = clnup(v)
		pl.Put(v)
	}
}

func (sp *sncpool[V]) New() (v V) {
	if pl, clnup, new, ovrflw := sp.pl, sp.clnup, sp.new, sp.ovrflw; pl != nil && clnup != nil && new != nil {
		select {
		case v = <-ovrflw:
			return
		default:
		}

		v, _ = pl.Get().(V)
		var val = reflect.ValueOf(v)
		if !(!val.IsZero() && !val.IsNil()) || val.Interface() == nil {
			clnup(v)
			v = pl.New().(V)
			return
		}
	}
	return
}

func (sp *sncpool[V]) Get() (v V) {
	if pl, clnup, ovrflw := sp.pl, sp.clnup, sp.ovrflw; pl != nil && clnup != nil {
		select {
		case v = <-ovrflw:
			return
		default:
		}
		v, _ = pl.Get().(V)
		var val = reflect.ValueOf(v)
		if !(!val.IsZero() && !val.IsNil()) || val.Interface() == nil {
			clnup(v)
			v = pl.New().(V)
			return
		}
	}
	return
}

type PoolNewFunc[V any] func() V

func (plnwfnc PoolNewFunc[V]) New() V {
	return plnwfnc()
}

type PoolCleanupFunc[V any] func(V) V

func (plclnfnc PoolCleanupFunc[V]) Cleanup(clnme V) V {
	return plclnfnc(clnme)
}

func NewPool[V any](new PoolNewFunc[V], cleanup PoolCleanupFunc[V]) Pool[V] {
	if cleanup == nil {
		cleanup = func(v V) V {
			return v
		}
	}
	return &sncpool[V]{
		pl: &sync.Pool{New: func() any {
			return new()
		}},
		clnup: cleanup,
		new:   new, ovrflw: make(chan V),
	}
}
