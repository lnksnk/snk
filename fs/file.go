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
	var f *struct {
		StatFile
		io.Reader
		io.Closer
		io.Seeker
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
		io.Closer
		io.Seeker
	}{
		StatFile: fstat,
		Reader:   frdr,
		Closer:   fcls,
		Seeker:   fskr,
	}
	return f
}
