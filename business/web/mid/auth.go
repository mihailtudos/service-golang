package mid

import (
	"context"
	"errors"
	"fmt"
	"github.com/mihailtudos/service3/business/sys/auth"
	"github.com/mihailtudos/service3/business/sys/validate"
	"github.com/mihailtudos/service3/foundation/web"
	"net/http"
	"strings"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {
	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Expecting Authorization: Bearer <token>
			authStr := r.Header.Get("Authorization")

			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("invalid authorization header format: bearer <token>")
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			// Add claims to the context so they can be retrieved later.
			ctx = auth.SetClaims(ctx, claims)

			// Call the next handler.
			return next(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(roles ...string) web.Middleware {
	h := func(next web.Handler) web.Handler {
		m := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value return failure.
			claims, err := auth.GetClaims(ctx)
			if err != nil {
				return validate.NewRequestError(
					fmt.Errorf("you are not authorized for that action, no claims"),
					http.StatusForbidden)
			}

			if !claims.Authorize(roles...) {
				return validate.NewRequestError(
					fmt.Errorf("you are not authorized for that action claims[%v] roles[%v", claims.Roles, roles),
					http.StatusForbidden,
				)
			}

			return next(ctx, w, r)
		}

		return m
	}

	return h
}
