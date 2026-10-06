package io

import (
	"bufio"
	"context"
	"io"
	"math"
	"sync"
	"unicode/utf8"
)

type RunesBufferReader interface {
	io.WriterTo
	io.RuneReader
	io.Closer
	Size() int64
	String() string
}

type RunesBufferWriter interface {
	Write(...rune) error
	io.Closer
	io.ReaderFrom
	Size() int64
	Empty() bool
	String() string
	Reader() (r RunesBufferReader, err error)
	ReadFrom(r io.Reader) (n int64, err error)
	Print(a ...any) error
	Println(a ...any) error
	Clone() RunesBufferWriter
	Reset()
}

type runesbufferwriter struct {
	rnsbffr [][]rune
	rns     []rune
	rnsi    int
	lck     *sync.RWMutex
}

// Clone implements [RunesBufferWriter].
func (r *runesbufferwriter) Clone() RunesBufferWriter {
	if rnsbffr, rns, rnsi, lck := r.rnsbffr[:], r.rns[:], r.rnsi, r.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		return &runesbufferwriter{rnsbffr: append([][]rune{}, rnsbffr...), rns: append([]rune{}, rns...), rnsi: rnsi, lck: &sync.RWMutex{}}
	}
	return nil
}

// Close implements [RunesBufferWriter].
func (r *runesbufferwriter) Close() (err error) {
	if r == nil {
		return
	}
	if lck := r.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		r.rns = nil
		r.rnsbffr = nil
		r.lck = nil
	}
	return
}

// Empty implements [RunesBufferWriter].
func (r *runesbufferwriter) Empty() bool {
	return r == nil || r.Size() == 0
}

// Print implements [RunesBufferWriter].
func (r *runesbufferwriter) Print(a ...any) (err error) {
	if r == nil || len(a) == 0 {
		return
	}
	pi, pw := io.Pipe()
	ctx, ctxcncl := context.WithCancel(context.Background())
	defer func() { <-ctx.Done() }()
	go func() {
		defer ctxcncl()
		if _, pwerr := Fprint(pw, a...); pwerr != nil {
			pw.CloseWithError(pwerr)
			return
		}
		pw.Close()
	}()

	bfr := bufio.NewReaderSize(pi, 1)
	rr, sz := rune(0), 0
	for err == nil {
		if rr, sz, err = bfr.ReadRune(); sz > 0 {
			if err != nil {
				if err == io.EOF {
					if err = r.Write(rr); err != nil {
						return
					}
					err = nil
				}
				return
			}
			if err = r.Write(rr); err != nil {
				return
			}
			continue
		}
		if err == io.EOF {
			err = nil
		}
		return
	}
	if err == io.EOF {
		err = nil
	}
	return
}

// Println implements [RunesBufferWriter].
func (r *runesbufferwriter) Println(a ...any) (err error) {
	if r == nil {
		return
	}
	pi, pw := io.Pipe()
	ctx, ctxcncl := context.WithCancel(context.Background())
	defer func() { <-ctx.Done() }()
	go func() {
		defer ctxcncl()
		if len(a) == 0 {
			if _, pwerr := Fprintln(pw); pwerr != nil {
				pw.CloseWithError(pwerr)
				return
			}
			pw.Close()
			return
		}
		if _, pwerr := Fprintln(pw, a...); pwerr != nil {
			pw.CloseWithError(pwerr)
			return
		}
		pw.Close()
	}()

	bfr := bufio.NewReaderSize(pi, 1)
	rr, sz := rune(0), 0
	for err == nil {
		if rr, sz, err = bfr.ReadRune(); sz > 0 {
			if err != nil {
				if err == io.EOF {
					if err = r.Write(rr); err != nil {
						return
					}
					err = nil
				}
				return
			}
			if err = r.Write(rr); err != nil {
				return
			}
			continue
		}
		if err == io.EOF {
			err = nil
		}
		return
	}
	if err == io.EOF {
		err = nil
	}
	return
}

// ReadFrom implements [RunesBufferWriter].
func (r *runesbufferwriter) ReadFrom(rr io.Reader) (n int64, err error) {
	if r == nil {
		return
	}
	if lck := r.lck; lck != nil {
		lck.Lock()
		defer lck.Unlock()
		var rdr, _ = rr.(io.RuneReader)
		if rdr == nil {
			rdr = bufio.NewReader(rr)
		}
		for rn := range IterRunes(func(itrerr error) {
			err = itrerr
		}, rdr) {
			if r.rnsi++; r.rnsi <= math.MaxInt {
				r.rns = append(r.rns, rn)
				if r.rnsi == math.MaxInt {
					r.rnsbffr = append(r.rnsbffr, r.rns)
					r.rns = nil
					r.rnsi = 0
				}
				continue
			}
		}
	}
	return
}

type runesbufferreader struct {
	rnsbffr [][]rune
}

// Close implements [RunesBufferReader].
func (r *runesbufferreader) Close() (err error) {
	if r == nil {
		return
	}
	r.rnsbffr = nil
	return
}

// ReadRune implements [RunesBufferReader].
func (r *runesbufferreader) ReadRune() (rr rune, size int, err error) {
	if rnsbffr := r.rnsbffr; len(rnsbffr) > 0 {
		rns := rnsbffr[0]
		rr = rns[0]
		if rns = rns[1:]; len(rns) == 0 {
			r.rnsbffr = rnsbffr[1:]
		}
		return rr, utf8.RuneLen(rr), nil
	}
	return 0, 0, io.EOF
}

// Size implements [RunesBufferReader].
func (r *runesbufferreader) Size() (n int64) {
	if rnsbffr := r.rnsbffr[:]; len(rnsbffr) > 0 {
		for _, rns := range rnsbffr {
			for _, rr := range rns {
				n += int64(utf8.RuneLen(rr))
			}
		}
	}
	return
}

// String implements [RunesBufferReader].
func (r *runesbufferreader) String() (s string) {
	if rnsbffr := r.rnsbffr[:]; len(rnsbffr) > 0 {
		for _, rns := range rnsbffr {
			s += string(rns)
		}
	}
	return
}

// WriteTo implements [RunesBufferReader].
func (r *runesbufferreader) WriteTo(w io.Writer) (n int64, err error) {
	if rnsbffr := r.rnsbffr[:]; len(rnsbffr) > 0 {
		bn := int64(0)
		for _, rns := range rnsbffr {
			if bn, err = Fprint(w, rns); err == nil {
				n += bn
			}
		}
	}
	return
}

// Reader implements [RunesBufferWriter].
func (r *runesbufferwriter) Reader() (rdr RunesBufferReader, err error) {
	if lck := r.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		if rnsl := len(r.rnsbffr); r.rnsi > 0 || rnsl > 0 {
			var rnsbffr = [][]rune{}
			if rnsl > 0 {
				rnsbffr = append(rnsbffr, r.rnsbffr[:]...)
			}
			if r.rnsi > 0 {
				rnsbffr = append(rnsbffr, r.rns[:r.rnsi])
			}
			return &runesbufferreader{rnsbffr: rnsbffr}, nil
		}
	}
	return
}

// Reset implements [RunesBufferWriter].
func (r *runesbufferwriter) Reset() {
	if r == nil {
		return
	}
	if lck := r.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()

	}
}

// Size implements [RunesBufferWriter].
func (r *runesbufferwriter) Size() (sz int64) {
	if r == nil {
		return
	}
	if lck := r.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		for _, rns := range r.rnsbffr {
			for _, rn := range rns {
				sz += int64(utf8.RuneLen(rn))
			}
		}
		for _, rn := range r.rns[:r.rnsi] {
			sz += int64(utf8.RuneLen(rn))
		}
	}
	return
}

// String implements [RunesBufferWriter].
func (r *runesbufferwriter) String() (s string) {
	if r == nil {
		return
	}
	if lck := r.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		for _, rns := range r.rnsbffr {
			s += string(rns)
		}
		s += string(r.rns[:r.rnsi])
	}
	return
}

// Write implements [RunesBufferWriter].
func (r *runesbufferwriter) Write(rr ...rune) (err error) {
	if r == nil {
		return
	}
	if lck := r.lck; lck != nil {
		lck.Lock()
		defer lck.Unlock()
	}
	return
}

func WriteRunesBuffer() RunesBufferWriter {
	return &runesbufferwriter{rns: []rune{}, lck: &sync.RWMutex{}}
}
