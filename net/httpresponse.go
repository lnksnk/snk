package net

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type responsewriter struct {
	conn       net.Conn
	wtrbf      *bufio.Writer
	headers    http.Header
	statusCode int
	headerSent bool
	closeConn  bool
}

// Close implements an anonymous interface.
func (w *responsewriter) Close() error {
	wtrbf := w.wtrbf
	w.wtrbf, w.conn = nil, nil
	if wtrbf != nil && wtrbf.Buffered() > 0 {
		return wtrbf.Flush()
	}
	return nil
}

// Closed implements an anonymous interface.
func (w *responsewriter) Closed() bool {
	return w.closeConn || w.conn == nil || w.wtrbf == nil || (w.headers != nil && w.headers.Get("Connection") == "close")
}

// Flush implements [http.Flusher].
func (w *responsewriter) Flush() {
	if wtrbf, conn, headerSent := w.wtrbf, w.conn, w.headerSent; wtrbf != nil && conn != nil {
		if !headerSent {
			w.WriteHeader(http.StatusOK)
		}
		if wtrbf.Buffered() > 0 {
			if err := wtrbf.Flush(); err != nil {
				w.conn = nil
				w.wtrbf = nil
			}
			wtrbf.Reset(conn)
		}
	}
}

// Header implements [http.ResponseWriter].
func (w *responsewriter) Header() http.Header {
	return w.headers
}

// Write implements [http.ResponseWriter].
func (w *responsewriter) Write(data []byte) (int, error) {
	if wtrbf := w.wtrbf; wtrbf != nil {
		// If WriteHeader hasn't been explicitly called, implicit 200 OK is sent first
		if !w.headerSent {
			w.WriteHeader(http.StatusOK)
		}
		return wtrbf.Write(data)
	}
	return 0, net.ErrClosed
}

// WriteHeader implements [http.ResponseWriter].
func (w *responsewriter) WriteHeader(statusCode int) {
	if w.headerSent {
		return
	}
	if wtrbf, conn := w.wtrbf, w.conn; wtrbf != nil && conn != nil {
		w.statusCode = statusCode
		w.headerSent = true

		// 1. Write the Status Line (e.g., HTTP/1.1 200 OK)
		statusText := http.StatusText(w.statusCode)
		fmt.Fprintf(wtrbf, "HTTP/1.1 %d %s\r\n", w.statusCode, statusText)

		// 2. Write all mapped headers
		for key, values := range w.headers {
			for _, value := range values {
				fmt.Fprintf(wtrbf, "%s: %s\r\n", key, value)
			}
		}

		// 3. Write the blank line ending header metadata block
		fmt.Fprint(wtrbf, "\r\n")
		if err := wtrbf.Flush(); err != nil {
			w.wtrbf = nil
			w.conn = nil
			return
		}
		wtrbf.Reset(conn)
	}
}

func ResponseWriter(conn net.Conn) (w interface {
	http.ResponseWriter
	http.Flusher
	Closed() bool
	Close() error
}) {
	resp := &responsewriter{conn: conn, headers: make(http.Header), wtrbf: bufio.NewWriter(conn), statusCode: http.StatusOK}
	w = resp
	return
}
