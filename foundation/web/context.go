package web

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// key is how request values or stored/retrieved.
const key ctxKey = 1

// Values represents states for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) (*Values, error) {
	values, ok := ctx.Value(key).(*Values)
	if !ok {
		return nil, fmt.Errorf("web value missing from context")
	}
	return values, nil
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	values, err := GetValues(ctx)
	if err != nil {
		return "00000000-0000-0000-0000-000000000000"
	}
	return values.TraceID
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) error {
	values, err := GetValues(ctx)
	if err != nil {
		return errors.New("web value missing from context")
	}
	values.StatusCode = statusCode
	return nil
}
