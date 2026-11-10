package net

import (
	"bufio"
	"net/http"
)

func Request(b *bufio.Reader) (r *http.Request, err error) {
	return http.ReadRequest(b)
}
