package tests

import (
	"encoding/json"
	"github.com/mihailtudos/service3/app/services/sales-api/handlers"
	"github.com/mihailtudos/service3/business/data/tests"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// TestUsers is the entry point for testing the user API endpoints.
func TestUsers(t *testing.T) {
	test := tests.NewIntegration(
		t,
		tests.DBContainer{
			Image: "postgres:17-alpine",
			Port:  "5432",
			Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
		},
	)

	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ts := UserTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		adminToken: test.Token("admin@example.com", "gophers"),
		userToken:  test.Token("user@example.com", "gophers"),
	}

	t.Run("getToken200", ts.getToken200)
}

func (ut *UserTests) getToken200(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("admin@example.com", "gophers")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to issue tokens to known users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with a valid credentials.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a HTTP 200 status code : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a HTTP 200 status code.", tests.Success, testID)

			var got struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to decode the response : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to decode the response.", tests.Success, testID)
		}
	}
}
