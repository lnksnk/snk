package serversent

import (
	"net/http"
	"strings"
)

func UpgradeAndServe(w http.ResponseWriter, r *http.Request, servesse func(w http.ResponseWriter, r *http.Request) error) (served bool, err error) {
	accepts := r.Header.Get("Accept")
	if served = accepts != "" && strings.Contains(accepts, "text/event-stream"); served {
		if servesse != nil {
			err = servesse(w, r)
			return
		}
	}
	return
}
