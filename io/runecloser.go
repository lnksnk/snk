package io

import "io"

type RuneReadCloser interface {
	io.RuneReader
	io.Closer
}

type RuneReadCloseHandler interface {
	ReadRune() (r rune, size int, err error)
	Close() error
}
