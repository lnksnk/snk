package serve

import "io"

type Handler interface {
	Serve(path string, w io.Writer, r io.Reader) (err error)
	Print(a ...any) (err error)
	Println(a ...any)
	Flush() (err error)
	ReadRune() (r rune, size int, err error)
	WriteTo(w io.Writer) (n int64, err error)
	ReadFrom(r io.Reader) (n int64, err error)
	ReadRunes() func(func(rune, error) bool)
}

func Serve(a ...any) (hndlr Handler, err error) {
	return &Server{}, nil
}

type Server struct {
}

// Flush implements [Handler].
func (s *Server) Flush() (err error) {
	panic("unimplemented")
}

// Print implements [Handler].
func (s *Server) Print(a ...any) (err error) {
	panic("unimplemented")
}

// Println implements [Handler].
func (s *Server) Println(a ...any) {
	panic("unimplemented")
}

// ReadFrom implements [Handler].
func (s *Server) ReadFrom(r io.Reader) (n int64, err error) {
	panic("unimplemented")
}

// ReadRune implements [Handler].
func (s *Server) ReadRune() (r rune, size int, err error) {
	panic("unimplemented")
}

// ReadRunes implements [Handler].
func (s *Server) ReadRunes() func(func(rune, error) bool) {
	panic("unimplemented")
}

// WriteTo implements [Handler].
func (s *Server) WriteTo(w io.Writer) (n int64, err error) {
	panic("unimplemented")
}

// Serve implements [Handler].
func (s *Server) Serve(path string, w io.Writer, r io.Reader) (err error) {
	panic("unimplemented")
}
