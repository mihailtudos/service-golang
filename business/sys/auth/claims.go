package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

// Authorize returns true if the claims has at least one of the provided roles.
func (c Claims) Authorize(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}

// ctxKey represents the type of value for the context key.
type ctxKey int

// key is how request values or stored/retrieved.
const key ctxKey = 1

// SetClaims stores the claims in the context.
func SetClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, key, c)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) (Claims, error) {
	v, ok := ctx.Value(key).(Claims)
	if !ok {
		return Claims{}, errors.New("claims value missing from request context")
	}
	return v, nil
}
