package net

import "net"

type Listener interface {
	net.Listener
}

type listener struct {
	ln net.Listener
}

// Accept implements [Listener].
func (l *listener) Accept() (net.Conn, error) {
	if ln := l.ln; ln != nil {
		return ln.Accept()
	}
	return nil, net.ErrClosed
}

// Addr implements [Listener].
func (l *listener) Addr() net.Addr {
	if ln := l.ln; ln != nil {
		return ln.Addr()
	}
	return nil
}

// Close implements [Listener].
func (l *listener) Close() (err error) {
	ln := l.ln
	l.ln = nil
	if ln != nil {
		err = ln.Close()
	}
	return
}

func New(network, addr string) (Listener, error) {
	var ln, lnerr = net.Listen(network, addr)
	if lnerr != nil {
		return nil, lnerr
	}
	return &listener{ln: ln}, nil
}
