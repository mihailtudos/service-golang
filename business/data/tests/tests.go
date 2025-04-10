// Package tests contains supporting code for running tests.
package tests

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/service3/business/data/schema"
	"github.com/mihailtudos/service3/business/sys/database"
	"github.com/mihailtudos/service3/foundation/docker"
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
