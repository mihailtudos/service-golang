// Package testgr provides a handler for the test endpoints group.
package testgr

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
}

// Test handles the test endpoint.
func (h *Handlers) Test(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	json.NewEncoder(w).Encode(data)

	statusCode := http.StatusOK
	h.Log.Infow("test", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)
}
