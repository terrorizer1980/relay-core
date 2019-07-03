package server

import (
	"net/http"
	"path"

	utilapi "github.com/puppetlabs/horsehead/httputil/api"
	"github.com/puppetlabs/nebula-tasks/pkg/data/secrets"
)

type secretsHandler struct {
	sec secrets.Store
}

func (h *secretsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var tn, key string

	// TODO clean this path validation logic up
	tn, r.URL.Path = shiftPath(r.URL.Path)
	if tn == "" || tn == "/" {
		http.NotFound(w, r)

		return
	}

	key, r.URL.Path = shiftPath(r.URL.Path)
	if key == "" || key == "/" {
		http.NotFound(w, r)

		return
	}

	sec, err := h.sec.Get(r.Context(), path.Join(tn, key))
	if err != nil {
		utilapi.WriteError(r.Context(), w, err)

		return
	}

	utilapi.WriteObjectOK(r.Context(), w, sec)
}
