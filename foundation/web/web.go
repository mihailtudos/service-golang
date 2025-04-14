// Package web contains a small web framework extension.
package web

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Handler handles a http request within our little mini framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers.
type App struct {
	mux      *httptreemux.ContextMux
	otmux    http.Handler
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handles a set of routes for the application
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/response.
	//
	// This is configured to use the W3C TraceContext standard to set the remote IP
	// and anderson headers in every trace.

	mux := httptreemux.NewContextMux()
	otmux := otelhttp.NewHandler(mux, "request")

	return &App{
		mux:      mux,
		otmux:    otmux,
		shutdown: shutdown,
		mw:       mw,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is detected
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This is set up in the NewApp() function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server
func (a *App) Handle(method, group, path string, handler Handler, mw ...Middleware) {
	// First wrap handler specific middleware around the given handler - local mw.
	handler = wrapMiddleware(mw, handler)

	// Second wrap the given handler middleware around the new handler - application level mw.
	handler = wrapMiddleware(a.mw, handler)

	// function that executes for each request
	h := func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		// Capture the parent request span from the request context.
		span := trace.SpanFromContext(ctx)

		// Set the context with the required values to process the request.
		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Now:     time.Now(),
		}

		ctx = context.WithValue(ctx, key, &v)

		// Call the wrapped handler functions
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			// Log error - handle it
			// ERROR HANDLING
			return
		}

		// POST CODE PROCESSING
		// login ended
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.mux.Handle(method, finalPath, h)
}
