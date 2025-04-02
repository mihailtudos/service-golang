package mid

import (
	"context"
	"github.com/mihailtudos/service3/foundation/web"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Logger(log *zap.SugaredLogger) web.Middleware {

	m := func(next web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := web.GetValues(ctx)
			if err != nil {
				return err // web.NewShutdownError("web value missing from context")
			}

			log.Infow("request start", "traceID", v.TraceID, "method", r.Method,
				"path", r.URL.Path, "remoteaddr", r.RemoteAddr)

			err = next(ctx, w, r)

			log.Infow("request completed", "traceID", v.TraceID, "method", r.Method,
				"path", r.URL.Path, "remoteaddr", r.RemoteAddr, "statusCode", v.StatusCode, "since", time.Since(v.Now))

			return err
		}

		return h
	}

	return m
}
