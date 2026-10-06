package fs

import (
	"io"
	"io/fs"
	"sync"
	"time"

	snkio "github.com/lnksnk/snk/io"
)

type MemFile interface {
	Stat() (FileInfo, error)
	File() File
}

type memfile struct {
	lck *sync.RWMutex
	mfi FileInfo
	mbw snkio.BufferWriter
}

func (m *memfile) Close() {
	if m == nil {
		return
	}
	if lck := m.lck; lck != nil {
		lck.Lock()
		defer lck.Unlock()
		if mfb := m.mbw; mfb != nil {
			m.mbw = nil
			mfb.Close()
		}
		m.mfi = nil
	}
}

func (m *memfile) Stat() (fi FileInfo, err error) {
	if fi = m.mfi; fi != nil {
		return
	}
	err = fs.ErrClosed
	return
}

func (m *memfile) File() (f File) {
	if lck, mfi, mbw := m.lck, m.mfi, m.mbw; lck != nil && mbw != nil && mfi != nil {
		lck.RLock()
		defer lck.RUnlock()
		if bfr, _ := mbw.Reader(); bfr != nil {
			f = &memf{mfi: GenFileInfo(mfi.Path(), mfi.Base(), mfi.Size(), false, mfi.ModTime()), mbr: bfr}
			return
		}
	}
	return
}

func (m *memfile) Set(path, root string, a ...any) (err error) {
	if lck, mfw := m.lck, m.mbw; lck != nil {
		lck.Lock()
		defer lck.Unlock()
		var bfw snkio.BufferWriter
		if bfw, err = snkio.WriterBuffer(a...); err != nil {
			if bfw != nil {
				bfw.Close()
			}
			return
		}
		if mfw != nil {
			mfw.Close()
			mfw = bfw
			m.mbw = mfw
		}
		if mfw == nil {
			mfw = bfw
			m.mbw = mfw
		}
		m.mfi = GenFileInfo(path, root, mfw.Size(), false, time.Now())
	}
	return
}

func (m *memfile) Append(path, root string, a ...any) (err error) {
	if lck, mfw := m.lck, m.mbw; lck != nil {
		lck.Lock()
		defer lck.Unlock()
		var bfw snkio.BufferWriter
		if bfw, err = snkio.WriterBuffer(a...); err != nil {
			if bfw != nil {
				bfw.Close()
			}
			return
		}
		if mfw != nil {
			bfw.WriteTo(mfw)
		}
		if mfw == nil {
			mfw = bfw
			m.mbw = mfw
		}
		m.mfi = GenFileInfo(path, root, mfw.Size(), false, time.Now())
	}
	return
}

type memf struct {
	mfi FileInfo
	mbr snkio.BufferReader
}

// Close implements [File].
func (m *memf) Close() (err error) {
	mbr := m.mbr
	m.mbr = nil
	m.mfi = nil
	if mbr != nil {
		mbr.Close()
	}
	return
}

// Read implements [File].
func (m *memf) Read(p []byte) (n int, err error) {
	mbr := m.mbr
	if mbr != nil {
		if n, err = mbr.Read(p); err != nil {
			m.mbr = nil
			mbr.Close()
			m.Close()
		}
		return
	}
	return 0, io.EOF
}

// Stat implements [File].
func (m *memf) Stat() (fi FileInfo, err error) {
	if fi = m.mfi; fi != nil {
		return
	}
	err = &fs.PathError{Op: "Stat", Path: "", Err: fs.ErrClosed}
	return
}

func (m *memf) Seek(offset int64, whence int) (n int64, err error) {
	if mbr := m.mbr; mbr != nil {
		return mbr.Seek(offset, whence)
	}
	err = &fs.PathError{Op: "Stat", Path: "", Err: fs.ErrClosed}
	return
}
