package mid

import (
	"context"
	"github.com/mihailtudos/service3/business/sys/metrics"
	"github.com/mihailtudos/service3/foundation/web"
	"net/http"
)

// Metrics updates program counters.
func Metrics() web.Middleware {
	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// Add the metrics into the context for metric gathering.
			ctx = metrics.Set(ctx)

			// Call the next handler.
			err := next(ctx, w, r)

			// Handle updating the metrics that can be updated.
			n := metrics.AddRequests(ctx)
			if n%1000 == 0 {
				metrics.AddGoroutines(ctx)
			}

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}

	return m
}
