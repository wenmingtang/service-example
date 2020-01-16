package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gitlab.fenbishuo.com/fenbishuo/service-training/cmd/sales-api/internal/handlers"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/tests"
)

func TestUsers(t *testing.T) {
	test := tests.New(t)
	defer test.Teardown()

	shutdown := make(chan os.Signal, 1)

	ut := UserTests{app: handlers.API(shutdown, test.DB, test.Log, test.Authenticator)}

	t.Run("TokenRequireAuth", ut.TokenRequireAuth)
	t.Run("TokenDenyUnknown", ut.TokenDenyUnknown)
	t.Run("TokenDenyBadPassword", ut.TokenDenyBadPassword)
	t.Run("TokenSuccess", ut.TokenSuccess)
}

type UserTests struct {
	app http.Handler
}

func (ut *UserTests) TokenRequireAuth(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/users/token", nil)
	resp := httptest.NewRecorder()

	ut.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("getting: expected status code %v, got %v", http.StatusUnauthorized, resp.Code)
	}
}

func (ut *UserTests) TokenDenyUnknown(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/users/token", nil)
	resp := httptest.NewRecorder()

	req.SetBasicAuth("unknown@example.com", "gophers")

	ut.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("getting: expected status code %v, got %v", http.StatusUnauthorized, resp.Code)
	}
}

func (ut *UserTests) TokenDenyBadPassword(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/users/token", nil)
	resp := httptest.NewRecorder()

	req.SetBasicAuth("admin@example.com", "GOPHERS")

	ut.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("getting: expected status code %v, got %v", http.StatusUnauthorized, resp.Code)
	}
}

func (ut *UserTests) TokenSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/users/token", nil)
	resp := httptest.NewRecorder()

	req.SetBasicAuth("admin@example.com", "gophers")

	ut.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code %v, got %v", resp.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decoding: %s", err)
	}

	if len(got) != 1 {
		t.Error("unexpected values in token response")
	}

	if got["token"] == "" {
		t.Fatal("token was not in response")
	}
}
