// Package testgr provides a handler for the test endpoints group.
package testgr

import (
	"context"
	"errors"
	"github.com/mihailtudos/service3/business/sys/validate"
	"github.com/mihailtudos/service3/foundation/web"
	"math/rand"
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
	if n := rand.Intn(100); n%2 == 0 {
		//panic("test panic")
		return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}

	data := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	statusCode := http.StatusOK

	return web.Respond(ctx, w, data, statusCode)
}
