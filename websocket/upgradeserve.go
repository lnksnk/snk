package websocket

import (
	"net/http"
	"strings"
)

func UpgradeAndServe(w http.ResponseWriter, r *http.Request, servews func(w http.ResponseWriter, r *http.Request) error) (served bool, err error) {
	if served = strings.Contains(r.Header.Get("Connection"), "Upgrade") && strings.Contains(r.Header.Get("Upgrade"), "websocket"); served {
		if servews != nil {
			err = servews(w, r)
		}
		return
	}
	return
}
