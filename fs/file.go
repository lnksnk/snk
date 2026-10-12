package fs

import (
	"io"

	snkio "github.com/lnksnk/snk/io"
)

// A File provides access to a single file.
// The File interface is the minimum implementation required of the file.
type File interface {
	Stat() (FileInfo, error)
	Read([]byte) (int, error)
	Close() error
}

type SyncFile interface {
	Sync() error
}

type ReadAtFile interface {
	ReadAt(b []byte, off int64) (n int, err error)
}

type ReadAtFileFunc func(b []byte, off int64) (n int, err error)

func (rdatf ReadAtFileFunc) ReadAt(b []byte, off int64) (n int, err error) {
	return rdatf(b, off)
}

type SeekFile interface {
	Seek(offset int64, whence int) (ret int64, err error)
}

type ReadFromFile interface {
	ReadFrom(r io.Reader) (n int64, err error)
}

type WriteToFile interface {
	WriteTo(w io.Writer) (n int64, err error)
}

type WriteToFileFunc func(w io.Writer) (n int64, err error)

func (wrtf WriteToFileFunc) WriteTo(w io.Writer) (n int64, err error) {
	return wrtf(w)
}

type WriteFile interface {
	Write(b []byte) (n int, err error)
}

type WriteAtFile interface {
	WriteAt(b []byte, off int64) (n int, err error)
}

type TruncateFile interface {
	Truncate(size int64) error
}

type PrintFile interface {
	Print(...any) (n int64, err error)
}

type PrintlnFile interface {
	Print(...any) (n int64, err error)
}

// func (f *os.File) ReadDir(n int) ([]os.DirEntry, error)
// func (f *os.File) ReadFrom(r io.Reader) (n int64, err error)
//func (f *os.File) Readdir(n int) ([]os.FileInfo, error)
//func (f *os.File) Readdirnames(n int) (names []string, err error)
//func (f *os.File) SetDeadline(t time.Time) error
//func (f *os.File) SetReadDeadline(t time.Time) error
//func (f *os.File) SetWriteDeadline(t time.Time) error
//func (f *os.File) Stat() (os.FileInfo, error)
//func (f *os.File) SyscallConn() (syscall.RawConn, error)
//func (f *os.File) Truncate(size int64) error
//func (f *os.File) Write(b []byte) (n int, err error)
//func (f *os.File) WriteAt(b []byte, off int64) (n int, err error)

// Stat interface of [File] returning [FileInfo]
type StatFile interface {
	Stat() (FileInfo, error)
}

type StatFileFunc func() (FileInfo, error)

func (stf StatFileFunc) Stat() (FileInfo, error) {
	return stf()
}

// Open interface of [File] returnin a [File] or error
type OpenFile interface {
	Open(name string) (File, error)
}

type OpenFileFunc func(name string) (File, error)

func (of OpenFileFunc) Open(name string) (File, error) {
	return of(name)
}

func NewFile(fstat StatFile, frdr io.Reader, fskr io.Seeker, fcls io.Closer) File {
	var rdat, _ = frdr.(io.ReaderAt)
	var f *struct {
		StatFile
		io.Reader
		io.ReaderAt
		io.Closer
		io.Seeker
		io.WriterTo
	}
	var orgfcls = fcls
	if orgfcls != nil {
		fcls = snkio.CloseFunc(func() (err error) {
			f.Reader = nil
			f.StatFile = nil
			if orgfcls != nil {
				orgfcls.Close()
			}
			return
		})
	}

	f = &struct {
		StatFile
		io.Reader
		io.ReaderAt
		io.Closer
		io.Seeker
		io.WriterTo
	}{
		StatFile: fstat,
		Reader:   frdr,
		ReaderAt: ReadAtFileFunc(func(b []byte, off int64) (n int, err error) {
			if rdat != nil {
				n, err = rdat.ReadAt(b, off)
			}
			if fskr != nil && frdr != nil {
				if _, err = fskr.Seek(off, io.SeekStart); err != nil {
					return
				}
				n, err = frdr.Read(b)
			}
			return
		}),
		Closer: fcls,
		Seeker: fskr,
		WriterTo: WriteToFileFunc(func(w io.Writer) (n int64, err error) {
			if frdr != nil {
				n, err = snkio.Fprint(w, frdr)
			}
			return
		}),
	}
	return f
}
