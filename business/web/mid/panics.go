package mid

import (
	"context"
	"fmt"
	"github.com/mihailtudos/service3/business/sys/metrics"
	"github.com/mihailtudos/service3/foundation/web"
	"net/http"
	"runtime/debug"
)

// Panics recovers from panics and converts the panic to an error
// so it is easier to report to Metrics and handle the errors.
func Panics() web.Middleware {

	// This is the middleware function to recover from a panic.
	m := func(next web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {

					// Grab the stack trace.
					trace := debug.Stack()
					err = fmt.Errorf("PANIC: [%v] TRACE: [%s]", rec, string(trace))

					// Update the metrics stored in the context.
					metrics.AddPanics(ctx)
				}
			}()

			// Call the next handler and set its return value in the err variable.
			return next(ctx, w, r)
		}
		return h
	}

	return m
}
