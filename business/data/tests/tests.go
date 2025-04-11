// Package tests contains supporting code for running tests.
package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/service3/business/data/schema"
	"github.com/mihailtudos/service3/business/data/store/user"
	"github.com/mihailtudos/service3/business/sys/auth"
	"github.com/mihailtudos/service3/business/sys/database"
	"github.com/mihailtudos/service3/foundation/docker"
	"github.com/mihailtudos/service3/foundation/keystore"
	"github.com/mihailtudos/service3/foundation/logger"
	"go.uber.org/zap"
	"io"
	"os"
	"testing"
	"time"
)

// Success and failure messages.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// DBContainer provides a container for testing.
type DBContainer struct {
	Image string
	Port  string
	Args  []string
}

// NewUnit creates a test database inside a container. It creates the
// required database schema in an empty database. It returns
// the database to use as well as a function to call at the end of the tests.
func NewUnit(t *testing.T, dbc DBContainer) (*zap.SugaredLogger, *sqlx.DB, func()) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	c := docker.StartContainer(t, dbc.Image, dbc.Port, dbc.Args...)
	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	})

	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}

	t.Log("waiting for database to be ready...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("migrating database: %v", err)
	}

	if err := schema.Seed(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("seeding database: %v", err)
	}

	log, err := logger.New("TEST")
	if err != nil {
		t.Fatalf("logger error: %s", err)
	}

	teardown := func() {
		t.Helper()
		db.Close()
		docker.StopContainer(t, c.ID)

		log.Sync()
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = old
		fmt.Println("****************************** LOGS *******************************")
		fmt.Println(buf.String())
		fmt.Println("****************************** LOGS *******************************")
	}

	return log, db, teardown
}

// Test owns state for running tests.
type Test struct {
	Log      *zap.SugaredLogger
	DB       *sqlx.DB
	Auth     *auth.Auth
	Teardown func()

	t *testing.T
}

// NewIntegration creates a database, seeds it, and sets up authentication.
func NewIntegration(t *testing.T, dbc DBContainer) *Test {
	log, db, teardown := NewUnit(t, dbc)

	// Create RSA keys needed to enable authentication in our services.
	keyID := "4754d86b-7a6d-4df5-9c65-224741361492"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating private key: %v", err)
	}

	auth, err := auth.New(keyID, keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
	if err != nil {
		t.Fatalf("constructing auth: %v", err)
	}

	test := &Test{
		Log:      log,
		DB:       db,
		Auth:     auth,
		Teardown: teardown,
		t:        t,
	}

	return test
}

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}

// Token generates new auth token for the specified user.
func (test *Test) Token(email, password string) string {
	test.t.Log("generating token for test...")

	store := user.NewStore(test.DB, test.Log)
	claims, err := store.Authenticate(context.Background(), time.Now(), email, password)
	if err != nil {
		test.t.Fatalf("authenticating: %v", err)
	}
	token, err := test.Auth.GenerateToken(claims)
	if err != nil {
		test.t.Fatalf("generating token: %v", err)
	}

	return token
}
