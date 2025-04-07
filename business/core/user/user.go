// Package user provides and example of a core business API.
// Right now these calls are just wrapping db/sql calls. But at some point,
// you will want auditing, logging, and metrics which is where you would place
// that logic rather than in the handlers.
package user

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/service3/business/data/store/user"
	"github.com/mihailtudos/service3/business/sys/auth"
	"go.uber.org/zap"
	"time"
)

type Core struct {
	user user.Store
	log  *zap.SugaredLogger
}

func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		log:  log,
		user: user.NewStore(db, log),
	}
}

func (c Core) Create(ctx context.Context, nu user.NewUser, now time.Time) (user.User, error) {

	u, err := c.user.Create(ctx, nu, now)
	if err != nil {
		return user.User{}, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (c Core) Update(ctx context.Context, claims auth.Claims, userID string, uu user.UpdateUser, now time.Time) error {

	err := c.user.Update(ctx, claims, userID, uu, now)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (c Core) Delete(ctx context.Context, claims auth.Claims, userID string) error {

	err := c.user.Delete(ctx, claims, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]user.User, error) {
	users, err := c.user.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (c Core) QueryByID(ctx context.Context, claims auth.Claims, userID string) (user.User, error) {
	u, err := c.user.QueryByID(ctx, claims, userID)
	if err != nil {
		return user.User{}, fmt.Errorf("query user: %w", err)
	}
	return u, nil
}

// QueryByEmail gets the specified user from the database by email address.
func (c Core) QueryByEmail(ctx context.Context, claims auth.Claims, email string) (user.User, error) {
	u, err := c.user.QueryByEmail(ctx, claims, email)
	if err != nil {
		return user.User{}, fmt.Errorf("query user: %w", err)
	}
	return u, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (c Core) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	claims, err := c.user.Authenticate(ctx, now, email, password)
	if err != nil {
		return auth.Claims{}, fmt.Errorf("authenticate: %w", err)
	}

	return claims, nil
}
