package io

import (
	"io"
	"unicode"
	"unicode/utf8"
)

// RunesToUTF8 convert rs []rune to []byte of raw utf8
func RunesToUTF8(rs ...rune) (bs []byte) {
	if rnl := len(rs); rnl > 0 {
		size := 0
		rnsze := make([]int, rnl)
		for rn := range rs {
			sze := utf8.RuneLen(rs[rn])
			rnsze[rn] = sze
			size += sze
			if rn == rnl-1 {
				bs = make([]byte, size)
				count := 0
				for rn := range rnl {
					rnsze[rn] = utf8.EncodeRune(bs[count:], rs[rn])
					count += rnsze[rn]
				}
				return
			}
		}
	}

	return bs
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func IsSpace(r rune) bool {
	return asciiSpace[r] == 1 || unicode.IsSpace(r)
}

func IsTxt(r rune) bool {
	return txtpar[r] == 1
}

var txtpar = [256]uint8{'\'': 1, '"': 1, '`': 1}

type CloseFunc func() error

func (clsefnc CloseFunc) Close() error {
	return clsefnc()
}

type ReadFunc func([]byte) (n int, err error)

func (rdfnc ReadFunc) Read(p []byte) (int, error) {
	return rdfnc(p)
}

type ReaderFunc func() (io.Reader, error)

func (rdrfnc ReaderFunc) Read() (io.Reader, error) {
	return rdrfnc()
}

func ReadUntilBytes(orgr io.Reader, reseteof bool, eof ...byte) io.Reader {
	ei := 0
	eprv := byte(0)
	var rdbt = make([]byte, 1)
	var rerr error
	var rn = 0
	var rdbtsiter = func(f func(byte) bool) {
		if eofl := len(eof); eofl > 0 {
			for rerr == nil {
				if rn, rerr = orgr.Read(rdbt); rn > 0 {
					if ei > 0 && eof[ei-1] == eprv && eof[ei] != rdbt[0] {
						for _, rb := range eof[:ei] {
							if !f(rb) {
								return
							}
						}
						ei = 0
						eprv = 0
					}
					if eof[ei] == rdbt[0] {
						if ei++; ei == eofl {
							if rerr == nil {
								rerr = io.EOF
							}
							if reseteof {
								ei = 0
								eprv = 0
							}
							return
						}
						eprv = rdbt[0]
						continue
					}
					if ei > 0 {
						for _, rb := range eof[:ei] {
							if !f(rb) {
								return
							}
						}
						ei = 0
						eprv = rdbt[0]
						if !f(rdbt[0]) {
							return
						}
						continue
					}
					eprv = rdbt[0]
					if !f(rdbt[0]) {
						return
					}
					continue
				}
			}
		}
	}
	return ReadFunc(func(b []byte) (n int, err error) {
		if btl := len(b); btl > 0 {
			for rb := range rdbtsiter {
				b[n] = rb
				if n++; n == btl {
					return
				}
			}
			if rerr != nil {
				err = rerr
			}
			return
		}
		return 0, io.EOF
	})
}
