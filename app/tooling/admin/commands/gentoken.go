// Package commands contains the commands for the admin tool.
package commands

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GenToken generates a JWT for the specified user.
func GenToken() error {
	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   uuid.UUID{}.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Roles: []string{"USER"},
	}

	// This will generate a JWT with the claims embedded in them. The database
	// with need to be configured with the information found in the public key
	// file to validate these claims. Dgraph does not support key rotate at
	// this time.
	kid := "456F21BD-1296-449A-9C2E-85A92092E966"

	_, err := GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return nil
}

func GenerateToken(kid string, claims jwt.Claims) (string, error) {
	method := jwt.GetSigningMethod("RS256")
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = kid

	pathName := path.Join("zarf/keys", kid) + ".pem"
	file, err := os.Open(pathName)
	if err != nil {
		return "", fmt.Errorf("opening key file: %w", err)
	}
	defer file.Close()
	// limit PEM file size to 1 megabyte. This should be reasonable for
	// almost any PEM file and prevents shenanigans like linking the file
	// to /dev/random or something like that.
	pemFile, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return "", fmt.Errorf("reading auth private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemFile)
	if err != nil {
		return "", fmt.Errorf("parsing auth private key: %w", err)
	}

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", tokenStr)

	// ==========================================================================================

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return "", fmt.Errorf("encoding to public file: %w", err)
	}

	// ==========================================================================================

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	var parsedClaims struct {
		jwt.RegisteredClaims
		Roles []string
	}

	fmt.Println("=================================")

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		fmt.Printf("kid: %s\n", kid)
		return &privateKey.PublicKey, nil
	}

	parsedToken, err := parser.ParseWithClaims(tokenStr, &parsedClaims, keyFunc)
	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}
	if !parsedToken.Valid {
		return "", errors.New("invalid token")
	}

	fmt.Printf("token is valid\n")
	return tokenStr, nil
}
