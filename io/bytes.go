package io

import (
	"bufio"
	"bytes"
	"io"
	"math"
)

type BytesReader interface {
	io.Reader
	io.WriterTo
	io.RuneReader
	io.Closer
	Reset() error
	Size() int64
	String() string
}

// BytesWriter presents Writer a Print interface for cached slice of []byte
type BytesWriter interface {
	io.Writer
	io.WriterTo
	WriteRunes(...rune) error
	io.Closer
	FlushWrite(w io.Writer, p []byte) (n int, err error)
	FlushWriteRunes(w io.Writer, pr []rune) error
	FlushWriteTo(io.Writer) (n int, err error)
	Size() int64
	Reader() (r BytesReader, err error)
	ReadFrom(r io.Reader) (n int64, err error)
	Clone() BytesWriter
	String() string
	Bytes() []byte
}

func ReaderBytes(bts []byte) BytesReader {
	return &bytesReader{br: bytes.NewReader(bts)}
}

// WriteBytes returns an implemented instance of [BytesWriter]
func WriterBytes() BytesWriter {
	return &bytesWrite{}
}

type bytesWrite struct {
	bts  []byte
	btsl int
}

// Bytes implements [BytesWriter].
func (b *bytesWrite) Bytes() []byte {
	return b.bts[:]
}

// WriteTo implements [BytesWriter].
func (b *bytesWrite) WriteTo(w io.Writer) (n int64, err error) {
	if btsn, btsi, bts, btsl := 0, 0, b.bts[:], b.btsl; btsl > 0 && len(bts) == btsl {
		for err == nil && btsi < btsl {
			if btsn, err = w.Write(bts[btsi:]); btsn > 0 {
				btsi += btsn
				n += int64(btsn)
				continue
			}
			return
		}
	}
	return
}

// String implements [BytesWriter].
func (b *bytesWrite) String() (s string) {
	if bts := b.bts[:]; len(bts) > 0 {
		return string(bts)
	}
	return
}

// Clone implements [BytesWriter].
func (b *bytesWrite) Clone() BytesWriter {
	bts, btsl := b.bts[:], b.btsl
	return &bytesWrite{bts: bts, btsl: btsl}
}

// ReadFrom implements [BytesWriter].
func (b *bytesWrite) ReadFrom(r io.Reader) (n int64, err error) {
	if b == nil || r == nil {
		return
	}
	bfr := bufio.NewWriter(WriteFunc(func(p []byte) (n int, err error) {
		return b.Write(p)
	}))
	n, err = bfr.ReadFrom(r)
	bfr.Flush()
	return
}

// Reader implements [BytesWriter].
func (b *bytesWrite) Reader() (r BytesReader, err error) {
	if b == nil {
		return
	}
	r = ReaderBytes(b.bts[:])
	return
}

type bytesReader struct {
	br *bytes.Reader
}

// String implements [BytesReader].
func (b *bytesReader) String() (s string) {
	if br := b.br; br != nil {
		for r := range IterRunes(nil, br) {
			s += string(r)
		}
	}
	return
}

// WriteTo implements [BytesReader].
func (b *bytesReader) WriteTo(w io.Writer) (n int64, err error) {
	if b == nil {
		return
	}
	if br := b.br; br != nil {
		return br.WriteTo(w)
	}
	return
}

// Reset implements [BytesReader].
func (b *bytesReader) Reset() (err error) {
	if br := b.br; br != nil {
		_, err = br.Seek(0, io.SeekStart)
	}
	return io.EOF
}

// Close implements [BytesReader].
func (b bytesReader) Close() (err error) {
	br := b.br
	b.br = nil
	if br != nil {
		br.Reset(nil)
	}
	return
}

// Read implements [BytesReader].
func (b bytesReader) Read(p []byte) (n int, err error) {
	if br := b.br; br != nil {
		if n, err = br.Read(p); err != nil {
			b.Close()
		}
		return
	}
	return 0, io.EOF
}

// ReadRune implements [BytesReader].
func (b bytesReader) ReadRune() (r rune, size int, err error) {
	if br := b.br; br != nil {
		if r, size, err = br.ReadRune(); err != nil {
			b.Close()
		}
		return
	}
	return 0, 0, io.EOF
}

// Size implements [BytesReader].
func (b bytesReader) Size() int64 {
	if br := b.br; br != nil {
		return br.Size()
	}
	return 0
}

// FlushWrite implements [BytesWriter].
func (b *bytesWrite) FlushWrite(w io.Writer, p []byte) (n int, err error) {
	if b == nil {
		return
	}
	if n, p, err = maxwrite(b, p); err == nil && len(p) > 0 {
		bts := b.bts
		b.bts = p[:]
		b.btsl = len(p[:])
		return w.Write(bts)
	}
	return 0, err
}

func maxwrite(w io.Writer, p []byte) (n int, bts []byte, err error) {
	if c := 1; w != nil {
		for c > 0 && n < len(p) && err == nil {
			c, err = w.Write(p[n:])
			n += c
		}
		if err != nil {
			bts = p[n:]
		}
	}
	return
}

// FlushWriteRunes implements [BytesWriter].
func (b *bytesWrite) FlushWriteRunes(w io.Writer, pr []rune) (err error) {
	_, err = b.FlushWrite(w, []byte(string(pr)))
	return
}

// FlushWriteTo implements [BytesWriter].
func (b *bytesWrite) FlushWriteTo(w io.Writer) (n int, err error) {
	if b == nil {
		return
	}
	bts := b.bts
	b.btsl = 0
	if len(bts) > 0 {
		if n, err = w.Write(bts); n > 0 {
			b.bts = bts[n:]
			b.btsl = len(bts[n:])
		}
		return
	}
	return
}

// Close implements [BytesWriter].
func (b *bytesWrite) Close() (err error) {
	if b == nil {
		return
	}
	b.bts = nil
	b.btsl = 0
	return
}

// Size implements [BytesWriter].
func (b *bytesWrite) Size() int64 {
	return int64(len(b.bts))
}

// Write implements [BytesWriter].
func (b *bytesWrite) Write(p []byte) (n int, err error) {
	if b == nil {
		return
	}
	if bl, pl := b.btsl, len(p); pl > 0 {
		if bl == 0 {
			b.bts = append([]byte{}, p...)
			b.btsl = len(p)
			return b.btsl, nil
		}
		if mdl := (bl + pl) % math.MaxInt; mdl > 0 && (bl+pl) > math.MaxInt {
			b.bts = append(b.bts, p[:pl-mdl]...)
			b.btsl += (pl - mdl)
			return pl - mdl, io.EOF
		}
		b.bts = append(b.bts, p...)
		b.btsl += len(p)
		if b.btsl == math.MaxInt {
			return pl, io.EOF
		}
		return pl, nil
	}
	return
}

// WriteRunes implements [BytesWriter].
func (b *bytesWrite) WriteRunes(rns ...rune) (err error) {
	_, err = b.Write([]byte(string(rns)))
	return
}
