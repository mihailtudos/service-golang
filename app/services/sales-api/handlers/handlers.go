// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"expvar"
	"github.com/jmoiron/sqlx"

	"net/http"
	"net/http/pprof"
	"os"

	"github.com/mihailtudos/service3/app/services/sales-api/handlers/debug/checkgr"
	v1TestGrp "github.com/mihailtudos/service3/app/services/sales-api/handlers/v1/testgrp"
	v1UserGrp "github.com/mihailtudos/service3/app/services/sales-api/handlers/v1/usergrp"
	userCore "github.com/mihailtudos/service3/business/core/user"
	"github.com/mihailtudos/service3/business/sys/auth"
	"github.com/mihailtudos/service3/business/web/mid"
	"github.com/mihailtudos/service3/foundation/web"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by the handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

// APIMux constrcuts an http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {
	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		// Panic and recover from panics need to be at the top of the chain
		mid.Panics(),
	)

	// Load the routes for the different versions of the API.
	v1(app, cfg)

	return app
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
func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	cgh := checkgr.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}

	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}

// v1 binds all version 1 routes.
func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"

	tgh := v1TestGrp.Handlers{
		Log: cfg.Log,
	}

	app.Handle(http.MethodGet, version, "/test", tgh.Test)
	app.Handle(http.MethodGet, version, "/testauth", tgh.Test, mid.Authenticate(cfg.Auth), mid.Authorize("ADMIN"))

	ugh := v1UserGrp.Handlers{User: userCore.NewCore(cfg.Log, cfg.DB), Auth: cfg.Auth}

	app.Handle(http.MethodGet, version, "/users/token", ugh.Token)
	app.Handle(http.MethodGet, version, "/users/:page/:rows", ugh.Query, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, version, "/users/:id", ugh.QueryByID, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPost, version, "/users", ugh.Create, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPut, version, "/users/:id", ugh.Update, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, version, "/users/:id", ugh.Delete, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
}
