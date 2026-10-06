package template

import (
	"bufio"
	"io"

	snkio "github.com/lnksnk/snk/io"
)

type RuneReaders interface {
	ReadRune() (r rune, size int, err error)
	Print(a ...any) (err error)
	Reset() error
}

type runereaders struct {
	rdrs  []io.RuneReader
	clsrs map[io.RuneReader]io.Closer
}

// Reset implements [RuneReaders].
func (rr *runereaders) Reset() (err error) {
	rdrs := rr.rdrs
	rr.rdrs = nil
	for ri := range rdrs {
		if clsr, _ := rdrs[ri].(io.Closer); clsr != nil {
			clsr.Close()
		}
	}
	return
}

// Print implements [RuneReaders].
func (rr *runereaders) Print(a ...any) (err error) {
	if clsrs := rr.clsrs; rr != nil {
		var rdrs []io.RuneReader
		var rnr io.RuneReader
		var crnta []any
		var cptrany = func() {
			if len(crnta) > 0 {
				bfw, bfwerr := snkio.WriterBuffer(crnta...)
				if err = bfwerr; err != nil {
					return
				}
				bfr, _ := bfw.Reader()
				bfw.Close()
				rdrs = append(rdrs, bfr)
			}
		}
		for ai := range a {
			if r, rk := a[ai].(io.Reader); rk {
				cptrany()
				if rnr, _ = r.(io.RuneReader); rnr == nil {
					rnr = bufio.NewReaderSize(r, 1)
				}
				if clsr, _ := r.(io.Closer); clsr != nil {
					if clsrs == nil {
						clsrs = map[io.RuneReader]io.Closer{}
						rr.clsrs = clsrs
					}
					clsrs[rnr] = clsr
				}
				rdrs = append(rdrs, rnr)
				continue
			}
			if rnr, _ := a[ai].(io.RuneReader); rnr != nil {
				cptrany()
				if clsr, _ := rnr.(io.Closer); clsr != nil {
					if clsrs == nil {
						clsrs = map[io.RuneReader]io.Closer{}
						rr.clsrs = clsrs
					}
					clsrs[rnr] = clsr
				}
				rdrs = append(rdrs, rnr)
				continue
			}
			crnta = append(crnta, a[ai])
		}
		cptrany()
		rr.rdrs = append(rdrs, rr.rdrs...)
	}
	return
}

// ReadRune implements [RuneReaders].
func (rr *runereaders) ReadRune() (r rune, size int, err error) {
	if rdrs, clsrs := rr.rdrs, rr.clsrs; len(rdrs) > 0 {
	rdnxt:
		if r, size, err = rdrs[0].ReadRune(); err != nil || size == 0 {
			if size == 0 && err == nil {
				err = io.EOF
			}
			rr.rdrs = rr.rdrs[1:]
			if size > 0 && err == io.EOF && len(rr.rdrs) > 0 {
				err = nil
			}
			if clsr, _ := rdrs[0].(io.Closer); clsr != nil {
				clsr.Close()
				if clsrs != nil {
					delete(clsrs, rdrs[0])
				}
			}
			if err == io.EOF && size == 0 && len(rr.rdrs) > 0 {
				rdrs = rr.rdrs
				goto rdnxt
			}
			if size > 0 && err == io.EOF {
				err = nil
			}
			if err != nil && err != io.EOF {
				rr.Reset()
				return
			}
		}
		return
	}
	return 0, 0, io.EOF
}

func Readers(a ...any) (rr RuneReaders, err error) {

	rr = &runereaders{}
	if err = rr.Print(a...); err != nil {
		rr.Reset()
		rr = nil
	}
	return
}
