package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/mihailtudos/service3/business/data/schema"
	"github.com/mihailtudos/service3/business/sys/database"
	"log"
	"os"
	"time"
)

func main() {
	err := migrate()
	if err != nil {
		log.Fatalf("error generating schema: %v", err)
	}

	err = seed()
	if err != nil {
		log.Fatalf("error seeding: %v", err)
	}

}

func seed() error {
	db, err := database.Open(database.Config{
		Host:         "localhost",
		Name:         "postgres",
		User:         "postgres",
		Password:     "password",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	})

	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("seed database successfully")

	return nil
}

func migrate() error {
	db, err := database.Open(database.Config{
		Host:         "localhost",
		Name:         "postgres",
		User:         "postgres",
		Password:     "password",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	})

	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrate database successfully")

	return nil
}

// GenKey creates an x509 private/public key for auth tokens.
func GenKey() error {

	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	fmt.Println("private and public key files generated")
	return nil
}
