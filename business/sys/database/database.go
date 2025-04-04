// Package database provides a database connection.
package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mihailtudos/service3/foundation/web"
	"go.uber.org/zap"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Set of errors for CRUD operations.
var (
	ErrNotFound             = errors.New("not found")
	ErrInvalidID            = errors.New("ID is not in its proper form")
	ErrForbidden            = errors.New("attempted action is not allowed")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslmode := "require"
	if cfg.DisableTLS {
		sslmode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslmode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {

	// If the user doesn't give us a deadline set 1 second.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	for attempts := 1; ; attempts++ {
		if err := db.Ping(); err == nil {
			break
		}

		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity.
	// Running this query forces a round trip through the database.
	const q = `SELECT TRUE`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing where field replacement is necessary.
func NamedExecContext(ctx context.Context, log *zap.SugaredLogger, db sqlx.DB, query string, data any) (err error) {
	q := queryString(query, data)
	log.Infow("database.NamedExecContext", "traceID", web.GetTraceID(ctx), "query", q)

	if _, err := db.NamedExecContext(ctx, query, data); err != nil {
		return err
	}

	return nil
}

// NamedQuerySlice is a helper function to execute a query that returns a
// collection of data to be unmarshalled into a slice.
func NamedQuerySlice(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, query string, data any, dest any) (err error) {
	q := queryString(query, data)
	log.Infow("database.NamedQuerySlice", "traceID", web.GetTraceID(ctx), "query", q)

	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Errorw("database.NamedQuerySlice.rows.Close", "error", err)
		}
	}()

	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		if err := rows.StructScan(v.Interface()); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}
	return nil
}

// NamedQueryStruct is a helper function to execute a query that returns a
// single data to be unmarshalled into a struct.
func NamedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, query string, data any, dest any) (err error) {
	q := queryString(query, data)
	log.Infow("database.NamedQueryStruct", "traceID", web.GetTraceID(ctx), "query", q)

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return ErrNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case []byte:
			value = fmt.Sprintf("'%s'", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
