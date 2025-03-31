// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/mihailtudos/service3/app/services/sales-api/handlers/debug/checkgr"
	"github.com/mihailtudos/service3/app/services/sales-api/handlers/v1/testgr"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by the handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

func APIMux(cfg APIMuxConfig) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	tgh := testgr.Handlers{
		Log: cfg.Log,
	}

	mux.Handle(http.MethodGet, "/v1/test", tgh.Test)

	return mux
}

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars/", expvar.Handler())

	return mux
}

// DebugMux registers
func DebugMux(build string, log *zap.SugaredLogger) http.Handler {
	mux := DebugStandardLibraryMux()

	cgh := checkgr.Handlers{
		Build: build,
		Log:   log,
	}

	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}
