package io

import (
	"reflect"
	"sync"
)

type SyncMap[T comparable, V any] interface {
	// Clear deletes all the entries, resulting in an empty Map.
	Clear()
	// CompareAndDelete deletes the entry for key if its value is equal to old.
	// The old value must be of a comparable type.
	//
	// If there is no current value for key in the map, CompareAndDelete
	// returns false (even if the old value is the nil interface value).
	CompareAndDelete(key T, old V) (deleted bool)
	// CompareAndSwap swaps the old and new values for key
	// if the value stored in the map is equal to old.
	// The old value must be of a comparable type.
	CompareAndSwap(key T, old V, new V) (swapped bool)
	// Delete deletes the value for a key.
	// If the key is not in the map, Delete does nothing.
	Delete(key T)
	// Exist returns true if a key was found in the map
	Exist(key T) (exist bool)
	// Load returns the value stored in the map for a key, or nil if no
	// value is present.
	// The ok result indicates whether value was found in the map.
	Load(key T) (value V, ok bool)
	// LoadAndDelete deletes the value for a key, returning the previous value if any.
	// The loaded result reports whether the key was present.
	LoadAndDelete(key T) (value V, loaded bool)
	// LoadOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored.
	LoadOrStore(key T, value V) (actual V, loaded bool)
	// Range calls f sequentially for each key and value present in the map.
	// If f returns false, range stops the iteration.
	//
	// Range does not necessarily correspond to any consistent snapshot of the Map's
	// contents: no key will be visited more than once, but if the value for any key
	// is stored or deleted concurrently (including by f), Range may reflect any
	// mapping for that key from any point during the Range call. Range does not
	// block other methods on the receiver; even f itself may call any method on m.
	//
	// Range may be O(N) with the number of elements in the map even if f returns
	// false after a constant number of calls.
	Range(f func(key T, value V) bool)
	// Store sets the value for a key.
	Store(key T, value V)
	// Swap swaps the value for a key and returns the previous value if any.
	// The loaded result reports whether the key was present.
	Swap(key T, value V) (previous V, loaded bool)
}

type genericsyncmap[T comparable, V any] struct {
	MapDeletedFunc[T, V]
	MapStoredFunc[T, V]
	sncmp *sync.Map
}

func (g *genericsyncmap[T, V]) Clear() {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		if evtdel != nil {
			var mp = map[T]V{}
			sncmp.Range(func(key, value any) bool {
				mp[key.(T)] = value.(V)
				return true
			})
			sncmp.Clear()
			for k, v := range mp {
				evtdel(k, v)
				delete(mp, k)
			}
			return
		}
		sncmp.Clear()
	}
}

func (g *genericsyncmap[T, V]) CompareAndDelete(key T, old V) (deleted bool) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		if deleted = sncmp.CompareAndDelete(key, old); deleted {
			if evtdel != nil {
				evtdel(key, old)
			}
		}
	}
	return
}

func (g *genericsyncmap[T, V]) CompareAndSwap(key T, old V, new V) (swapped bool) {
	if sncmp, evtdel, evtstr := g.sncmp, g.MapDeletedFunc, g.MapStoredFunc; sncmp != nil {
		swapped = sncmp.CompareAndSwap(key, old, new)
		if swapped {
			if evtdel != nil {
				evtdel(key, old)
			}
			if evtstr != nil {
				evtstr(key, new)
			}
		}
	}
	return
}

func (g *genericsyncmap[T, V]) Delete(key T) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		var prva, loaded = sncmp.LoadAndDelete(key)
		if evtdel != nil {
			if loaded {
				var prv, _ = prva.(V)
				evtdel(key, prv)
				return
			}
			evtdel(key)
		}
	}
}

func (g *genericsyncmap[T, V]) Load(key T) (value V, ok bool) {
	if sncmp := g.sncmp; sncmp != nil {
		var v any
		v, ok = sncmp.Load(key)
		value, _ = v.(V)
	}
	return
}

func (g *genericsyncmap[T, V]) LoadAndDelete(key T) (value V, loaded bool) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		var v any
		v, loaded = sncmp.LoadAndDelete(key)
		value, _ = v.(V)
		if !loaded && evtdel != nil {
			evtdel(key)
		}
	}
	return
}

func (g *genericsyncmap[T, V]) LoadOrStore(key T, value V) (actual V, loaded bool) {
	if sncmp, evtstr := g.sncmp, g.MapStoredFunc; sncmp != nil {
		var acv any
		acv, loaded = sncmp.LoadOrStore(key, value)
		actual, _ = acv.(V)
		if !loaded && evtstr != nil {
			evtstr(key, value)
		}
	}
	return
}

func (g *genericsyncmap[T, V]) Exist(key T) (exist bool) {
	if sncmp := g.sncmp; sncmp != nil {
		if sncmp := g.sncmp; sncmp != nil {
			sncmp.Range(func(k, value any) bool {
				exist = reflect.DeepEqual(k, key)
				return !exist
			})
		}
	}
	return
}

func (g *genericsyncmap[T, V]) Range(f func(key T, value V) bool) {
	if sncmp := g.sncmp; sncmp != nil {
		if sncmp := g.sncmp; sncmp != nil {
			sncmp.Range(func(key, value any) bool {
				return f(key.(T), value.(V))
			})
		}
	}
}

func (g *genericsyncmap[T, V]) Store(key T, value V) {
	if sncmp, evtstr := g.sncmp, g.MapStoredFunc; sncmp != nil {
		sncmp.Store(key, value)
		if evtstr != nil {
			evtstr(key, value)
		}
	}
}

func (g *genericsyncmap[T, V]) Swap(key T, value V) (previous V, loaded bool) {
	if sncmp, evtstr, evtdel := g.sncmp, g.MapStoredFunc, g.MapDeletedFunc; sncmp != nil {
		var prva any
		prva, loaded = sncmp.Swap(key, value)
		previous, _ = prva.(V)
		if evtstr != nil {
			evtstr(key, value)
		}
		if loaded && evtdel != nil {
			evtdel(key, previous)
		}
	}
	return
}

func NewSyncMap[T comparable, V any](a ...any) Map[T, V] {
	var evtdel MapDeletedFunc[T, V]
	var evtstr MapStoredFunc[T, V]

	for ai := range a {
		if evtdeld, evtdelk := a[ai].(MapDeletedFunc[T, V]); evtdelk {
			if evtdel == nil {
				evtdel = evtdeld
			}
			continue
		}
		if evtstrd, evtsrk := a[ai].(MapStoredFunc[T, V]); evtsrk {
			if evtstr == nil {
				evtstr = evtstrd
			}
			continue
		}
	}
	var orgevtdel = evtdel
	if orgevtdel != nil {
		evtdel = func(key T, value ...V) {
			go orgevtdel(key, value...)
		}
	}
	if orgevtdel == nil {
		evtdel = func(key T, value ...V) {}
	}
	var orgevtstr = evtstr
	if orgevtstr != nil {
		evtstr = func(key T, value V) {
			go orgevtstr(key, value)
		}
	}
	if orgevtstr == nil {
		evtstr = func(key T, value V) {}
	}
	return &genericsyncmap[T, V]{sncmp: &sync.Map{}, MapDeletedFunc: evtdel, MapStoredFunc: evtstr}
}
