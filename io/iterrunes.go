package io

import (
	"io"
	"math"
	"unicode/utf8"
)

var emptyrnsitr func(func(rune) bool) = func(f func(rune) bool) {

}

func IterRunes(evterr func(error), rnr ...io.RuneReader) func(func(rune) bool) {
	if len(rnr) == 0 {
		return emptyrnsitr
	}
	var chdrns []rune
	var chdl int
	var r, sz, rerr = rune(0), 1, error(nil)
	return func(f func(rune) bool) {
		var rnsl = len(rnr)
		var rdnne = func() (r rune, size int, err error) {
			if rnsl > 0 {
			again:
				if r, size, err = rnr[0].ReadRune(); size > 0 {
					if err != nil && err == io.EOF {
						if rnsl > 0 {
							rnr = rnr[1:]
							rnsl--
						}
					}
					return
				}
				if size == 0 && (err == nil || err == io.EOF) {
					if rnsl > 0 {
						rnr = rnr[1:]
						if rnsl--; rnsl > 0 {
							goto again
						}
					}
				}
			}
			return 0, 0, io.EOF
		}

	itrccchd:
		if chdl > 0 {
			if rerr != nil && rerr != io.EOF && evterr != nil {
				chdrns = nil
				evterr(rerr)
			}
			for chdl > 0 {
				r = chdrns[0]
				chdrns = chdrns[1:]
				sz = utf8.RuneLen(r)
				if chdl--; chdl >= 0 && f(r) {
					continue
				}
				return
			}
			chdrns = nil
		}
		if rerr == io.EOF {
			return
		}
		for rerr == nil {
			r, sz, rerr = rdnne()
			if sz > 0 {
				chdrns = append(chdrns, r)
				if chdl++; chdl == math.MaxInt || rerr == io.EOF {
					goto itrccchd
				}
				if rerr == nil {
					continue
				}
			}
			if sz == 0 || rerr == io.EOF {
				if chdl > 0 {
					goto itrccchd
				}
			}
			if rerr != nil && rerr != io.EOF && evterr != nil {
				evterr(rerr)
			}
			return
		}
	}
}
