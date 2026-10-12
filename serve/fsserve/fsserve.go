package fsserve

import (
	"io"
	"os"

	"github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
)

type FSServer struct {
	bfw          snkio.BufferWriter
	UTF8Exts     map[string]bool
	utf8exts     []string
	pathnotfound func(string)
	fsstat       fs.StatFileSystem
	fsopen       fs.OpenFileSystem
	fslist       fs.ListFileSystem
	fsset        fs.SetFileSystem
	fsappend     fs.AppendFileSystem
}

// Flush implements an anonymous interface.
func (f *FSServer) Flush() error {
	panic("unimplemented")
}

// Print implements an anonymous interface.
func (f *FSServer) Print(a ...any) error {
	panic("unimplemented")
}

// Println implements an anonymous interface.
func (f *FSServer) Println(a ...any) error {
	panic("unimplemented")
}

// ReadRunes implements an anonymous interface.
func (f *FSServer) ReadRunes() func(rune, error) bool {
	panic("unimplemented")
}

// Serve implements an anonymous interface.
func (f *FSServer) Serve(w io.Writer, r io.Reader) {
	panic("unimplemented")
}

// ServeStd implements an anonymous interface.
func (f *FSServer) ServeStd(stdout *os.File, stdin *os.File, stderr *os.File) {

}

func ServeFS(fsstat fs.StatFileSystem, fsopen fs.OpenFileSystem, pathnotfound func(string), a ...any) (hndlr interface {
	Print(a ...any) error
	Println(a ...any) error
	Flush() error
	ReadRunes() func(rune, error) bool
	Serve(w io.Writer, r io.Reader)
	ServeStd(stdout *os.File, stdin *os.File, stderr *os.File)
}, err error) {
	return &FSServer{}, nil
}
