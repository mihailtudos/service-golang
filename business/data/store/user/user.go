// Package user provides user related CRUD functionality.
package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mihailtudos/service3/business/sys/auth"
	"github.com/mihailtudos/service3/business/sys/database"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mihailtudos/service3/business/sys/validate"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	db  *sqlx.DB
	log *zap.SugaredLogger
}

func NewStore(db *sqlx.DB, log *zap.SugaredLogger) Store {
	return Store{
		db:  db,
		log: log,
	}
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, nu NewUser, now time.Time) (User, error) {
	if err := validate.Check(nu); err != nil {
		return User{}, fmt.Errorf("validating data: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generating password hash: %w", err)
	}

	usr := User{
		ID:           validate.GenerateID(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now,
		DateUpdated:  now,
	}

	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
	VALUES
	    (:user_id, :name, :email, :password_hash, :roles, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, usr); err != nil {
		return User{}, fmt.Errorf("inserting user: %w", err)
	}

	return usr, nil
}

func (s Store) Update(ctx context.Context, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	if err := validate.CheckID(userID); err != nil {
		return database.ErrInvalidID
	}

	if err := validate.Check(uu); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	usr, err := s.QueryByID(ctx, claims, userID)
	if err != nil {
		return fmt.Errorf("updating user userID[%s]: %w", userID, err)
	}

	if uu.Name != nil {
		usr.Name = *uu.Name
	}

	if uu.Email != nil {
		usr.Email = *uu.Email
	}

	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}

	if uu.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating password hash: %w", err)
		}
		usr.PasswordHash = hash
	}

	usr.DateUpdated = now

	const q = `
	UPDATE 
		users
	SET 
	    "name" = :name,
	    "email" = :email,
	    "roles" = :roles,
	    "password_hash" = :password_hash,
	    "date_updated" = :date_updated
	WHERE 
	    user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, usr); err != nil {
		return fmt.Errorf("updating user userID[%s]: %w", userID, err)
	}

	return nil
}

func (s Store) Delete(ctx context.Context, claims auth.Claims, userID string) error {
	if err := validate.CheckID(userID); err != nil {
		return database.ErrInvalidID
	}

	// If you are not an admin and looking to delete someone else then you.
	if !claims.Authorize(auth.RoleAdmin) && claims.Subject != userID {
		return database.ErrForbidden
	}

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
		DELETE FROM
			users
		WHERE
		    user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting user userID[%s]: %w", userID, err)
	}

	return nil
}

func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {
	data := struct {
		OffSet      int `db:"offset"`
		RowsPerPage int `db:"row_per_page"`
	}{
		OffSet:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		users
	ORDER BY
	    user_id
	OFFSET :offset ROWS FETCH NEXT :row_per_page ROWS ONLY `

	var users []User
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &users); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, database.ErrNotFound
		}

		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, claims auth.Claims, userID string) (User, error) {
	if err := validate.CheckID(userID); err != nil {
		return User{}, database.ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorize(auth.RoleAdmin) && claims.Subject != userID {
		return User{}, database.ErrForbidden
	}

	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	SELECT
		*
	FROM
	    users
	WHERE
	    user_id = :user_id`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return User{}, database.ErrNotFound
		}
		return User{}, fmt.Errorf("selecting user userID[%s]: %w", userID, err)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email address.
func (s Store) QueryByEmail(ctx context.Context, claims auth.Claims, email string) (User, error) {
	if err := validate.Email(email); err != nil {
		return User{}, database.ErrInvalidID
	}

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return User{}, database.ErrNotFound
		}
		return User{}, fmt.Errorf("selecting user email[%s]: %w", email, err)
	}

	if !claims.Authorize(auth.RoleAdmin) && claims.Subject != usr.ID {
		return User{}, database.ErrForbidden
	}

	return usr, nil
}

// Authenticate find a user by their email and verifies their password.
// On success, it returns a Claims representing this user.
// The claims can be used to generate a token for future authentication.
func (s Store) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	if err := validate.Email(email); err != nil {
		return auth.Claims{}, database.ErrInvalidID
	}

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `
		SELECT
			*
		FROM
		    users
		WHERE
		    email = :email`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return auth.Claims{}, database.ErrNotFound
		}
		return auth.Claims{}, fmt.Errorf("selecting user email[%s]: %w", email, err)
	}

	// Compare the provided password with the saved hash. Use the subtle.ConstantTimeCompare() function
	// to avoid a timing attack.
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, database.ErrAuthenticationFailed
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   usr.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Roles: usr.Roles,
	}

	return claims, nil
}
