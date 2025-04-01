// Package testgr provides a handler for the test endpoints group.
package testgr

import (
	"context"
	"github.com/mihailtudos/service3/foundation/web"
	"net/http"

	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
}

// Test handles the test endpoint.
func (h *Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	data := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	statusCode := http.StatusOK
	h.Log.Infow("test", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

	return web.Respond(ctx, w, data, statusCode)
}
