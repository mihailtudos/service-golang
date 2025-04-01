// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

// Handler handles a http request within our little mini framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers.
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handles a set of routes for the application
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is detected
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
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

		// Call the wrapped handler functions
		if err := handler(r.Context(), w, r); err != nil {

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

	a.ContextMux.Handle(method, finalPath, h)
}
