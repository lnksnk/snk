package net

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func ServeHttp(ln net.Listener, handler http.Handler) (err error) {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleServe(conn, handler)
	}
}

func handleServe(conn net.Conn, appHandler http.Handler) {
	defer conn.Close()

	// Keep-Alive configuration parameters
	const keepAliveTimeout = 5 * time.Second
	// 1. Set a read deadline for the incoming request header.
	// If the client idles longer than this, the read fails and the loop ends safely.
	conn.SetReadDeadline(time.Now().Add(keepAliveTimeout))

	var bufrw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	for {

		req, err := Request(bufrw.Reader)
		if err != nil {
			// If it's a timeout or EOF (client gracefully left), break out cleanly
			if err == io.EOF || isTimeout(err) {
				return
			}
			log.Printf("Read error: %v", err)
			return
		}

		// 2. Clear deadlines while executing the app logic
		conn.SetReadDeadline(time.Time{})

		// 3. Initialize custom ResponseWriter using our buffered writer stack
		w := ResponseWriter(conn, bufrw.Writer)

		// Determine if the client wants to terminate the persistence loop immediately
		clientWantsClose := strings.ToLower(req.Header.Get("Connection")) == "close" || req.Close

		// 4. Default policy: assume Keep-Alive unless explicitly denied
		if clientWantsClose {
			w.Header().Set("Connection", "close")
		} else {
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Keep-Alive", fmt.Sprintf("timeout=%d", int(keepAliveTimeout.Seconds())))
		}

		// Execute core application logic
		appHandler.ServeHTTP(w, req)

		w.Flush()

		// 6. Terminate connection immediately if requested by client or business logic
		if clientWantsClose || w.Closed() {
			w.Close()
			return
		}
		w.Close()
		// Drain any unread bytes from the request body to avoid corrupting the next stream packet
		if req.Body != nil {
			io.Copy(io.Discard, req.Body)
			req.Body.Close()
		}
		conn.SetReadDeadline(time.Now().Add(1 * time.Nanosecond))
	}
}

// Helper to determine if an error is an active network deadline timeout
func isTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}
