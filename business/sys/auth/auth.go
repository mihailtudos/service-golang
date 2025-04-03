// Package auth provides authentication and authorization support.
package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// KeyLookup declares a method set of behaviour for looking up
// public and private keys for JWT authentication.
type KeyLookup interface {
	PublicKey(kid string) (*rsa.PublicKey, error)
	PrivateKey(kid string) (*rsa.PrivateKey, error)
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	activeKID string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	keyFunc   jwt.Keyfunc
	parser    *jwt.Parser
}

// New creates an Auth to support authentication/authorization.
func New(activeKID string, keyLookup KeyLookup) (*Auth, error) {

	// The activeKID is the key identifier that is used to sign the tokens.
	_, err := keyLookup.PrivateKey(activeKID)
	if err != nil {
		return nil, errors.New("active KID does not exist in store")
	}

	method := jwt.GetSigningMethod("RS256")
	if method == nil {
		return nil, errors.New("parsing signing method: %w")
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kidID, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		return keyLookup.PublicKey(kidID)
	}

	// Create the token parser to use. The algorithm used to sign the JWT must be
	// validated to avoid a critical vulnerability:
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	return &Auth{
		activeKID: activeKID,
		keyLookup: keyLookup,
		method:    method,
		keyFunc:   keyFunc,
		parser:    parser,
	}, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = a.activeKID

	privateKey, err := a.keyLookup.PrivateKey(a.activeKID)
	if err != nil {
		return "", errors.New("private key lookup failed")
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.New("signing token: %w")
	}

	return str, nil
}

func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}
