package user_test

import (
	"context"
	"errors"
	"github.com/mihailtudos/service3/business/sys/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/mihailtudos/service3/business/data/store/user"
	"github.com/mihailtudos/service3/business/data/tests"
	"github.com/mihailtudos/service3/business/sys/auth"

	"testing"
	"time"
)

var dbc = tests.DBContainer{
	Image: "postgres:17-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
}

func TestUser(t *testing.T) {
	log, db, teardown := tests.NewUnit(t, dbc)
	t.Cleanup(teardown)

	store := user.NewStore(db, log)
	t.Log("Given the need to work with User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()
			now := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)

			nu := user.NewUser{
				Name:            "John Doe",
				Email:           "johndoe@example.com",
				Roles:           []string{auth.RoleUser},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			usr, err := store.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a user : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a user.", tests.Success, testID)

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   usr.ID,
					Issuer:    "service project",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
				Roles: []string{auth.RoleAdmin},
			}

			saved, err := store.QueryByID(ctx, claims, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve a user by id : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve a user by id.", tests.Success, testID)

			if diff := cmp.Diff(usr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff: %s", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same user.", tests.Success, testID)

			upd := user.UpdateUser{
				Name:  tests.StringPointer("Johnny Doe"),
				Email: tests.StringPointer("johnnydoe@example.com"),
			}

			claims = auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   usr.ID,
					Issuer:    "service project",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
				Roles: []string{auth.RoleAdmin},
			}

			if err := store.Update(ctx, claims, usr.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update a user : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update a user.", tests.Success, testID)

			saved, err = store.QueryByEmail(ctx, claims, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve a user by id : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve a user by id.", tests.Success, testID)

			if saved.Email != *upd.Email {
				t.Errorf("\t%s\tTest %d:\tShould be able to update a user. Diff: %s", tests.Failed, testID, err)
				t.Logf("\t\tTest %d:\tGot = %v", testID, saved.Email)
				t.Logf("\t\tTest %d:\tExp = %v", testID, upd.Email)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to user Email.", tests.Success, testID)
			}

			if saved.Name != *upd.Name {
				t.Errorf("\t%s\tTest %d:\tShould be able to update a user. Diff: %s", tests.Failed, testID, err)
				t.Logf("\t\tTest %d:\tGot = %v", testID, saved.Name)
				t.Logf("\t\tTest %d:\tExp = %v", testID, upd.Email)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to user Name.", tests.Success, testID)
			}

			// ========================== DELETE USER ==========================
			if err := store.Delete(ctx, claims, usr.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete a user : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete a user.", tests.Success, testID)

			_, err = store.QueryByID(ctx, claims, usr.ID)
			if !errors.Is(err, database.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve a user by id : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted user id.", tests.Success, testID)
		}
	}

}
