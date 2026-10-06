package io

import (
	"maps"
	"reflect"
	"sync"
)

type Map[T comparable, V any] interface {
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

type MapEvents[T comparable, V any] interface {
	Deleted(key T, value V)
	Stored(key T, previous V, value V)
}

type MapDeletedFunc[T comparable, V any] func(key T, value ...V)

func (mdf MapDeletedFunc[T, V]) Deleted(key T, value ...V) {
	mdf(key, value...)
}

type MapStoredFunc[T comparable, V any] func(key T, value V)

func (msf MapStoredFunc[T, V]) Stored(key T, value V) {
	msf(key, value)
}

type genericmap[T comparable, V any] struct {
	MapDeletedFunc[T, V]
	MapStoredFunc[T, V]
	//sncmp *sync.Map
	lck   *sync.RWMutex
	sncmp map[T]V
}

func (g *genericmap[T, V]) Lock() {
	g.lck.Lock()
}

func (g *genericmap[T, V]) Unlock() {
	g.lck.Unlock()
}

func (g *genericmap[T, V]) RLock() {
	g.lck.RLock()
}

func (g *genericmap[T, V]) RUnlock() {
	g.lck.RUnlock()
}

func (g *genericmap[T, V]) Clear() {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		g.Lock()
		g.sncmp = map[T]V{}
		g.Unlock()
		if evtdel != nil {
			for k, v := range sncmp {
				evtdel(k, v)
				delete(sncmp, k)
			}
			return
		}
		clear(sncmp)
	}
}

func (g *genericmap[T, V]) CompareAndDelete(key T, old V) (deleted bool) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		g.Lock()
		if val, ok := sncmp[key]; ok && reflect.DeepEqual(val, old) {
			deleted = true
			delete(sncmp, key)
			g.Unlock()
			if evtdel != nil {
				evtdel(key, old)
			}
			return
		}
		g.Unlock()
	}
	return
}

func (g *genericmap[T, V]) CompareAndSwap(key T, old V, new V) (swapped bool) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		g.Lock()
		if val, ok := sncmp[key]; ok && reflect.DeepEqual(val, old) {
			swapped = true
			sncmp[key] = new
			g.Unlock()
			if evtdel != nil {
				evtdel(key, old)
			}
			return
		}
		g.Unlock()
	}
	return
}

func (g *genericmap[T, V]) Delete(key T) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		g.Lock()
		if val, ok := sncmp[key]; ok {
			delete(sncmp, key)
			g.Unlock()
			if evtdel != nil {
				evtdel(key, val)
			}
			return
		}
		g.Unlock()
	}
}

func (g *genericmap[T, V]) Load(key T) (value V, ok bool) {
	if sncmp := g.sncmp; sncmp != nil {
		g.RLock()
		value, ok = sncmp[key]
		g.RUnlock()
	}
	return
}

func (g *genericmap[T, V]) LoadAndDelete(key T) (value V, loaded bool) {
	if sncmp, evtdel := g.sncmp, g.MapDeletedFunc; sncmp != nil {
		g.Lock()
		if value, loaded = sncmp[key]; loaded {
			delete(sncmp, key)
		}
		g.Unlock()
		if evtdel != nil {
			if loaded {
				evtdel(key, value)
				return
			}
			evtdel(key)
		}
	}
	return
}

func (g *genericmap[T, V]) LoadOrStore(key T, value V) (actual V, loaded bool) {
	if sncmp, evtstr := g.sncmp, g.MapStoredFunc; sncmp != nil {
		g.RLock()
		actual, loaded = sncmp[key]
		if !loaded {
			g.RUnlock()
			g.Lock()
			sncmp[key] = value
			actual = value
			g.Unlock()
			if evtstr != nil {
				evtstr(key, value)
			}
			return
		}
		g.RUnlock()
	}
	return
}

func (g *genericmap[T, V]) Exist(key T) (exist bool) {
	if sncmp := g.sncmp; len(sncmp) > 0 {
		g.RLock()
		defer g.RUnlock()
		for k := range sncmp {
			if exist = reflect.DeepEqual(k, key); exist {
				return exist
			}
		}
		return
	}
	return
}

func (g *genericmap[T, V]) Range(f func(key T, value V) bool) {
	if sncmp := g.sncmp; len(sncmp) > 0 {
		g.RLock()
		var sncmpcpy map[T]V
		sncmpcpy = maps.Clone(sncmp)
		g.RUnlock()
		for k, v := range sncmpcpy {
			if !f(k, v) {
				clear(sncmpcpy)
				return
			}
		}
		clear(sncmpcpy)
		return
	}
}

/*func (g *genericmap[T, V]) Store(key T, value V) {
	if sncmp, evtstr := g.sncmp, g.MapStoredFunc; sncmp != nil {
		sncmp.Store(key, value)
		if evtstr != nil {
			evtstr(key, value)
		}
	}
}*/

func (g *genericmap[T, V]) Store(key T, value V) {
	if sncmp, evtstr := g.sncmp, g.MapStoredFunc; sncmp != nil {
		g.Lock()
		sncmp[key] = value
		g.Unlock()
		if evtstr != nil {
			evtstr(key, value)
		}
	}
}

/*func (g *genericmap[T, V]) Swap(key T, value V) (previous V, loaded bool) {
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
}*/

func (g *genericmap[T, V]) Swap(key T, value V) (previous V, loaded bool) {
	if sncmp, evtstr, evtdel := g.sncmp, g.MapStoredFunc, g.MapDeletedFunc; sncmp != nil {
		g.Lock()
		previous, loaded = sncmp[key]
		sncmp[key] = value
		g.Unlock()
		if evtstr != nil {
			evtstr(key, value)
		}
		if loaded && evtdel != nil {
			evtdel(key, previous)
		}
	}
	return
}

func NewMap[T comparable, V any](a ...any) Map[T, V] {
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
	return &genericmap[T, V]{ /*sncmp: &sync.Map{},*/ sncmp: map[T]V{}, lck: &sync.RWMutex{}, MapDeletedFunc: evtdel, MapStoredFunc: evtstr}
}
