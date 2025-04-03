// Package keystore implements the auth.Keystore interface. This implementation
// uses a local file to store the keys.
package keystore

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
)

type KeyStore struct {
	mu   sync.RWMutex
	keys map[string]*rsa.PrivateKey
}

// NewKeyStore creates an empty - zero value - KeyStore.
func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]*rsa.PrivateKey),
	}
}

// NewMap creates a new KeyStore from a map.
func NewMap(store map[string]*rsa.PrivateKey) *KeyStore {
	return &KeyStore{
		keys: store,
	}
}

// NewFS constructs a KeyStore based on a set of PEM files rooted inside
// a directory. The name of each PEM file is used as the key identifier.
// Example: keystore.NewFS(os.DirFS("./zarf/keys/"))
// Example: zarf/keys/95A369C9-068E-4932-90FC-4D46ADFC0FB3.pem
func NewFS(fsys fs.FS) (*KeyStore, error) {
	ks := KeyStore{
		keys: make(map[string]*rsa.PrivateKey),
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walkdir error: %w", err)
		}

		if dirEntry.IsDir() {
			return nil
		}

		if path.Ext(fileName) != ".pem" {
			return nil
		}

		file, err := fsys.Open(fileName)
		if err != nil {
			return fmt.Errorf("opening key file: %w", err)
		}

		defer file.Close()

		// limit PEM file size to 1 megabyte. This should be reasonable for
		// almost any PEM file and prevents shenanigans like linking the file
		// to /dev/random or something like that.
		pemFile, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("reading auth private key: %w", err)
		}

		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemFile)
		if err != nil {
			return fmt.Errorf("parsing auth private key: %w", err)
		}

		// The key identifier is the file name without the extension.
		KID := strings.TrimSuffix(dirEntry.Name(), ".pem")

		ks.keys[KID] = privateKey
		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return &ks, nil
}

// Add adds a private key to the keystore.
func (ks *KeyStore) Add(key *rsa.PrivateKey, kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.keys[kid] = key
}

// Remove removes a private key from the keystore.
func (ks *KeyStore) Remove(kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	delete(ks.keys, kid)
}

// PrivateKey returns a private key from the keystore.
func (ks *KeyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	privateKey, ok := ks.keys[kid]
	if !ok {
		return nil, errors.New("key not found")
	}

	return privateKey, nil
}

// PublicKey returns a public key from the keystore.
func (ks *KeyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	privateKey, err := ks.PrivateKey(kid)
	if err != nil {
		return nil, err
	}

	return &privateKey.PublicKey, nil
}
