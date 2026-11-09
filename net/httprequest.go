package net

import (
	"bufio"
	"net"
	"net/http"
)

func Request(conn net.Conn) (r *http.Request, err error) {
	return http.ReadRequest(bufio.NewReader(conn))
}
