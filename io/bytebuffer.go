package io

import (
	"io"
	"strings"
	"sync"
)

// BufferWriter presents Writer a Print interface for cached slice of [][BytesWriter]
type BufferWriter interface {
	io.Writer
	WriteRunes(...rune) error
	io.Closer
	io.ReaderFrom
	Size() int64
	Empty() bool
	String() string
	Reader() (r BufferReader, err error)
	ReadFrom(r io.Reader) (n int64, err error)
	Print(a ...any) error
	Println(a ...any) error
	Clone() BufferWriter
	Reset()
	Contains(a ...any) bool
	HasSuffix(a ...any) bool
	HasPrefix(a ...any) bool
	io.WriterTo
}

// WriteBuffer returns an implemented instance of [BufferWriter]
func WriterBuffer(a ...any) (bfwtr BufferWriter, err error) {
	if len(a) == 0 {
		return &bufferwriter{btsw: WriterBytes(), lck: &sync.RWMutex{}}, err
	}
	bfwtr = &bufferwriter{btsw: WriterBytes(), lck: &sync.RWMutex{}}
	if err = bfwtr.Print(a...); err != nil {
		bfwtr.Close()
		return
	}
	return
}

type bufferwriter struct {
	btswtrs []BytesWriter
	btsw    BytesWriter
	lck     *sync.RWMutex
}

// HasPrefix implements [BufferWriter].
func (b *bufferwriter) HasPrefix(a ...any) (prefix bool) {
	if btswtrs, btsw, lck := b.btswtrs[:], b.btsw, b.lck; len(a) > 0 && !b.Empty() && lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		bfts, _ := WriterBuffer()
		bfts.Print(a...)
		var cntnbts = []byte(bfts.String())
		bfts.Close()
		var cnti, cntl = 0, len(cntnbts)
		if cntl == 0 {
			return
		}

		var eof = false

		var chkfrd = func(p ...byte) {
			for bi := range p {
				if cntnbts[cnti] == p[bi] {
					if prefix = cnti+1 == cntl; prefix {
						return
					}
					cnti++
					continue
				}
				eof = true
				return
			}
		}

		if btswl := len(btswtrs); btswl > 0 {
			for btsi := range btswtrs {
				if chkfrd(btswtrs[btsi].Bytes()...); eof || prefix {
					return
				}
			}
		}

		var frstbts []byte
		if btsw != nil {
			frstbts = btsw.Bytes()
		}
		chkfrd(frstbts...)
	}
	return
}

// HasSuffix implements [BufferWriter].
func (b *bufferwriter) HasSuffix(a ...any) (suffix bool) {
	if btswtrs, btsw, lck := b.btswtrs[:], b.btsw, b.lck; len(a) > 0 && !b.Empty() && lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		bfts, _ := WriterBuffer(a...)
		if bfts != nil {
			return
		}
		var cntnbts = []byte(bfts.String())
		bfts.Close()
		var cnti, cntl = 0, len(cntnbts)
		if cntl == 0 {
			return
		}
		cnti = cntl - 1
		var eof = false

		var chkfrd = func(p ...byte) {
			for bi := range p {
				if cntnbts[cnti] == p[len(p)-(bi+1)] {
					if suffix = cnti == 0; suffix {
						return
					}
					if cnti > 0 {
						cnti--
					}
					continue
				}
				eof = true
				return
			}
		}

		var frstbts []byte
		if btsw != nil {
			frstbts = btsw.Bytes()
		}
		if chkfrd(frstbts...); eof || suffix {
			return
		}

		if btswl := len(btswtrs); btswl > 0 {
			for btsi := range btswtrs {
				if chkfrd(btswtrs[btswl-(btsi+1)].Bytes()...); eof || suffix {
					return
				}
			}
		}

	}
	return
}

// Contains implements [BufferWriter].
func (b *bufferwriter) Contains(a ...any) (contain bool) {
	if btswtrs, btsw, lck := b.btswtrs[:], b.btsw, b.lck; len(a) > 0 && !b.Empty() && lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		bfts, _ := WriterBuffer(a...)
		if bfts == nil {
			return
		}
		var cntnbts = []byte(bfts.String())
		bfts.Close()
		var cnti, cntl = 0, len(cntnbts)
		if cntl == 0 {
			return
		}
		var cnttpi = cntl - 1

		var chkfrd = func(fwrd, rvrs bool, pp ...[]byte) bool {
			if len(pp) == 0 {
				return false
			}
			var p = pp[0]
			var tp = pp[0]
			if len(pp) > 1 {
				tp = pp[1]
			}
			if fwrd && rvrs {
				for bi := range p {
					var tbi = len(tp) - (bi + 1)
					if cntnbts[cnti] == p[bi] {
						if cnti+1 == cntl {
							return true
						}
						cnti++
						if cntnbts[cnttpi] == tp[tbi] {
							if cnttpi == 0 {
								return true
							}
							if cnttpi > 0 {
								cnttpi--
							}
							continue
						}
						cnttpi = cntl - 1
						continue
					}
					cnti = 0
					if cntnbts[cnttpi] == tp[tbi] {
						if cnttpi == 0 {
							return true
						}
						if cnttpi > 0 {
							cnttpi--
						}
						continue
					}
					cnttpi = cntl - 1
				}
				return false
			}
			if fwrd && !rvrs {
				for bi := range p {
					if cntnbts[cnti] == p[bi] {
						if cnti+1 == cntl {
							return true
						}
						cnti++
						continue
					}
					cnti = 0
				}
				return false
			}
			if !fwrd && rvrs {
				for bi := range p {
					var tbi = len(tp) - (bi + 1)
					if cntnbts[cnttpi] == tp[tbi] {
						if cnttpi == 0 {
							return true
						}
						if cnttpi > 0 {
							cnttpi--
						}
						continue
					}
					cnttpi = cntl - 1
				}
				return false
			}
			return false
		}

		var frstbts []byte
		if btsw != nil {
			frstbts = btsw.Bytes()
		}
		if len(frstbts) > 0 {
			if contain = chkfrd(len(btswtrs) == 0, true, frstbts); contain {
				return
			}
		}
		if btswl := len(btswtrs); btswl > 0 {
			for btsi := range btswtrs {
				if btsi < len(btswtrs)-1 {
					if contain = chkfrd(true, true, btswtrs[btsi].Bytes(), btswtrs[btswl-(btsi+1)].Bytes()); contain {
						return
					}
					continue
				}
				contain = chkfrd(true, false, btswtrs[btsi].Bytes())
				return
			}
		}
	}
	return
}

// WriteTo implements [BufferWriter].
func (b *bufferwriter) WriteTo(w io.Writer) (n int64, err error) {
	if btn, btswtrs, btsw, lck := int64(0), b.btswtrs[:], b.btsw, b.lck; lck != nil {
		lck.RLock()
		defer lck.RUnlock()
		btswtrs = append(btswtrs, btsw)
		for _, btsw = range btswtrs {
			if btn, err = btsw.WriteTo(w); btn > 0 {
				n += btn
			}
			if err != nil {
				return
			}
		}
	}
	return
}

// Reset implements [BufferWriter].
func (b *bufferwriter) Reset() {
	if b == nil {
		return
	}
	lck, btsw, btswtrs := b.lck, b.btsw, b.btswtrs
	lck.Lock()
	defer lck.Unlock()
	b.btsw = WriterBytes()
	b.btswtrs = nil
	if btsw != nil {
		btsw.Close()
	}
	for _, btsw = range btswtrs {
		btsw.Close()
	}
}

// String implements [BufferWriter].
func (b *bufferwriter) String() (s string) {
	btsw, btswtrs := b.btsw, b.btswtrs[:]
	for _, bw := range btswtrs {
		s += bw.String()
	}
	if btsw != nil {
		s += btsw.String()
	}
	return
}

// Clone implements [BufferWriter].
func (b *bufferwriter) Clone() BufferWriter {
	btswtrs, btsw, lck := b.btswtrs[:], b.btsw, b.lck
	var clnbts []BytesWriter
	var clnbtsw BytesWriter
	func() {
		lck.RLock()
		defer lck.RUnlock()
		clnbts = make([]BytesWriter, len(btswtrs))
		for orgn, orgbts := range btswtrs {
			clnbts[orgn] = orgbts.Clone()
		}
		if btsw != nil {
			clnbtsw = btsw.Clone()
		}
	}()
	return &bufferwriter{btswtrs: clnbts, btsw: clnbtsw, lck: &sync.RWMutex{}}
}

// Println implements [BufferWriter].
func (b *bufferwriter) Println(a ...any) (err error) {
	if b == nil {
		return
	}
	_, err = Fprintln(b, a...)
	return
}

// Print implements [BufferWriter].
func (b *bufferwriter) Print(a ...any) (err error) {
	if b == nil {
		return
	}
	_, err = Fprint(b, a...)
	return
}

// Close implements [BufferWriter].
func (b *bufferwriter) Close() (err error) {
	if b == nil {
		return
	}
	btsw := b.btsw
	b.btsw = nil
	btswtrs := b.btswtrs
	b.btswtrs = nil
	if btsw != nil {
		btsw.Close()
	}
	for _, btsw = range btswtrs {
		btsw.Close()
	}
	return
}

// ReadFrom implements [BufferWriter].
func (b *bufferwriter) ReadFrom(r io.Reader) (n int64, err error) {
	if b == nil {
		return
	}
	if btsw, lck := b.btsw, b.lck; btsw != nil {
		var rn = int64(0)
		for {
			func() {
				lck.RLock()
				defer lck.RUnlock()
				if rn, err = btsw.ReadFrom(r); err == io.EOF {
					b.btswtrs = append(b.btswtrs, btsw)
					btsw = WriterBytes()
					b.btsw = btsw

					n += rn
				}
			}()
			if rn > 0 {
				n += rn
				continue
			}
			return
		}
	}
	return
}

// Reader implements [BufferWriter].
func (b *bufferwriter) Reader() (r BufferReader, err error) {
	if b == nil {
		return
	}
	var btsrdrs []BytesReader
	if btws, btw, lck, sz := b.btswtrs, b.btsw, b.lck, int64(0); btw != nil {
		func() {
			lck.RLock()
			defer lck.RUnlock()
			var br BytesReader
			for _, bw := range btws {
				if br, err = bw.Reader(); err != nil {
					for _, br = range btsrdrs {
						br.Close()
					}
					return
				}
				if br != nil {
					sz += br.Size()
					btsrdrs = append(btsrdrs, br)
				}
			}
			if br, err = btw.Reader(); err != nil {
				for _, br = range btsrdrs {
					br.Close()
				}
				return
			}
			if br != nil {
				sz += br.Size()
				btsrdrs = append(btsrdrs, br)
			}
		}()
		if len(btsrdrs) > 0 {
			r = &bufferreader{btsr: btsrdrs[0], btsrdrs: btsrdrs[1:], sz: sz, crntoffset: 0}
			return
		}

	}
	return
}

// Size implements [BufferWriter].
func (b *bufferwriter) Size() (sz int64) {
	if b != nil {
		bstw, btswtrs := b.btsw, b.btswtrs[:]
		if bstw != nil {
			sz = bstw.Size()
		}
		for _, bstw = range btswtrs {
			sz += bstw.Size()
		}
	}
	return
}

// Empty implements [BufferWriter].
func (b *bufferwriter) Empty() bool {
	if b != nil {
		bstw, btswtrs := b.btsw, b.btswtrs[:]
		if bstw != nil {
			if bstw.Size() > 0 {
				return false
			}
		}
		for _, bstw = range btswtrs {
			if bstw.Size() == 0 {
				continue
			}
			return false
		}
	}
	return true
}

// Write implements [BufferWriter].
func (b *bufferwriter) Write(p []byte) (n int, err error) {
	if b == nil {
		return
	}
	if btsw := b.btsw; btsw != nil {
		rn := 0
		for {
			rn, p, err = maxwrite(btsw, p)
			n += rn
			if rn > 0 && err == io.EOF {
				err = nil
				b.btswtrs = append(b.btswtrs, btsw)
				btsw = WriterBytes()
				b.btsw = btsw
				continue
			}
			return
		}
	}
	return
}

// WriteRunes implements [BufferWriter].
func (b *bufferwriter) WriteRunes(rns ...rune) (err error) {
	if b == nil {
		return
	}
	_, err = strings.NewReader(string(rns)).WriteTo(b)
	return
}

type BufferReader interface {
	io.Reader
	io.Seeker
	io.WriterTo
	io.RuneReader
	io.Closer
	Size() int64
	String() string
}

type bufferreader struct {
	btsrdrs []BytesReader
	btsr    BytesReader
	//bfr     *bufio.Reader
	crntoffset int64
	sz         int64
}

// Seek implements [BufferReader].
func (b *bufferreader) Seek(offset int64, whence int) (r int64, err error) {
	if cof, br, sz := b.crntoffset, b.btsr, b.Size(); br != nil && sz > 0 {
		if whence == io.SeekEnd {
			if cof = sz - offset; cof < 0 {
				cof = 0
			}
			return seakreadbfr(b, cof, sz)
		}
		if whence == io.SeekCurrent || whence == io.SeekStart {
			if cof+offset < sz {
				cof += offset
				return seakreadbfr(b, cof, sz)
			}
		}
		return
	}
	return
}

func seakreadbfr(b *bufferreader, cof int64, sz int64) (r int64, err error) {
	var p, n = make([]byte, 1), 0
	for err == nil {
		if n, err = b.Read(p); n > 0 {
			if cof -= int64(n); cof == 0 {
				return
			}
		}
	}
	return
}

// String implements [BufferReader].
func (b *bufferreader) String() (s string) {
	if b == nil {
		return
	}
	btsrdrs, btsr := b.btsrdrs, b.btsr
	for _, btr := range btsrdrs {
		s += btr.String()
	}
	if btsr != nil {
		s += btsr.String()
	}
	return
}

// Close implements [BufferReader].
func (b *bufferreader) Close() (err error) {
	if b == nil {
		return
	}
	b.sz = 0
	btsrdrs := b.btsrdrs
	b.btsrdrs = nil
	btsr := b.btsr
	b.btsr = nil
	if btsr != nil {
		btsr.Close()
	}
	for _, btsr = range btsrdrs {
		btsr.Close()

	}
	btsrdrs = nil
	return
}

// Read implements [BufferReader].
func (b *bufferreader) Read(p []byte) (n int, err error) {
	if b == nil {
		return 0, io.EOF
	}
rdagn:
	if br, btsrdrs := b.btsr, b.btsrdrs; br != nil {
		defer func() {
			if b.sz, b.crntoffset = b.Size(), b.crntoffset+int64(n); b.crntoffset > b.sz {
				b.crntoffset = b.sz
			}
		}()
		var rn = 0
		for {
			if rn, err = br.Read(p[n:]); rn > 0 {
				n += rn
			}
			if rn == 0 || err == io.EOF {
				br.Close()
				b.btsr = nil
				if len(btsrdrs) > 0 {
					b.btsr = btsrdrs[0]
					btsrdrs = btsrdrs[1:]
					b.btsrdrs = btsrdrs
					if n < len(p) {
						goto rdagn
					}
				}
				return
			}
			if rn > 0 && n < len(p) {
				continue
			}
			return
		}
	}
	return 0, io.EOF
}

// ReadRune implements [BufferReader].
func (b *bufferreader) ReadRune() (r rune, size int, err error) {
	if b == nil {
		return
	}
rdagn:
	if btsr, btsrdrs := b.btsr, b.btsrdrs; btsr != nil {
		r, size, err = btsr.ReadRune()
		if size == 0 || err == io.EOF {
			btsr.Close()
			b.btsr = nil
			if len(btsrdrs) > 0 {
				b.btsr = btsrdrs[0]
				btsrdrs = btsrdrs[1:]
				b.btsrdrs = btsrdrs
				if size == 0 {
					goto rdagn
				}
			}
			return
		}
		return
	}
	return 0, 0, io.EOF
}

// Size implements [BufferReader].
func (b *bufferreader) Size() int64 {
	return b.sz
}

// WriteTo implements [BufferReader].
func (b *bufferreader) WriteTo(w io.Writer) (n int64, err error) {
	if b == nil {
		return
	}
rdagn:
	if btsr, btsrdrs := b.btsr, b.btsrdrs; btsr != nil {
		var rn int64
		for {
			if rn, err = btsr.WriteTo(w); rn > 0 {
				n += rn
			}
			if rn == 0 || err == io.EOF {
				btsr.Close()
				b.btsr = nil
				if len(btsrdrs) > 0 {
					b.btsr = btsrdrs[0]
					btsrdrs = btsrdrs[1:]
					b.btsrdrs = btsrdrs
					if rn == 0 {
						goto rdagn
					}
				}
				return
			}
		}
	}
	return
}
