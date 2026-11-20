package websocket

import (
	"context"
	"net/http"
	"strings"
)

type ServeWSContextFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

func UpgradeAndServeContext(ctx context.Context, w http.ResponseWriter, r *http.Request, servews ServeWSContextFunc) (served bool, err error) {
	if served = strings.Contains(r.Header.Get("Connection"), "Upgrade") && strings.Contains(r.Header.Get("Upgrade"), "websocket"); served {
		if servews != nil {
			err = servews(ctx, w, r)
		}
		return
	}
	return
}
