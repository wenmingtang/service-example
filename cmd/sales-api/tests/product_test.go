package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.fenbishuo.com/fenbishuo/service-training/cmd/sales-api/internal/handlers"
	"gitlab.fenbishuo.com/fenbishuo/service-training/internal/tests"
)

func TestProducts(t *testing.T) {
	test := tests.New(t)
	defer test.Teardown()

	shutdown := make(chan os.Signal, 1)

	tests := ProductTests{
		app:        handlers.API(shutdown, test.DB, test.Log, test.Authenticator),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	t.Run("List", tests.List)
	t.Run("CreateRequiresFields", tests.CreateRequiresFields)
	t.Run("ProductCRUD", tests.ProductCRUD)
	t.Run("SalesList", tests.SalesList)
	t.Run("AddSale", tests.AddSale)
}

type ProductTests struct {
	app        http.Handler
	adminToken string
}

func (p *ProductTests) List(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/products", nil)
	resp := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+p.adminToken)

	p.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("getting: expected status code %v, got %v", http.StatusOK, resp.Code)
	}

	var list []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decoding: %s", err)
	}

	want := []map[string]interface{}{
		{
			"id":           "a2b0639f-2cc6-44b8-b97b-15d69dbb511e",
			"name":         "Comic Books",
			"cost":         float64(50),
			"quantity":     float64(42),
			"sold":         float64(7),
			"revenue":      float64(350),
			"user_id":      "00000000-0000-0000-0000-000000000000",
			"date_created": "2019-01-01T00:00:01.000001Z",
			"date_updated": "2019-01-01T00:00:01.000001Z",
		},
		{
			"id":           "72f8b983-3eb4-48db-9ed0-e45cc6bd716b",
			"name":         "McDonalds Toys",
			"cost":         float64(75),
			"quantity":     float64(120),
			"sold":         float64(3),
			"revenue":      float64(225),
			"user_id":      "00000000-0000-0000-0000-000000000000",
			"date_created": "2019-01-01T00:00:02.000001Z",
			"date_updated": "2019-01-01T00:00:02.000001Z",
		},
	}

	if diff := cmp.Diff(want, list); diff != "" {
		t.Fatalf("Response did not match expected. Diff: \n%s", diff)
	}
}

func (p *ProductTests) ProductCRUD(t *testing.T) {
	var created map[string]interface{}

	{
		// create
		body := strings.NewReader(`{"name":"product0","cost":55,"quantity":6}`)

		req := httptest.NewRequest("POST", "/v1/products", body)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+p.adminToken)

		p.app.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("posting: expected status code %v, got %v", http.StatusOK, resp.Code)
		}

		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decoding: %s", err)
		}

		if created["id"] == "" || created["id"] == nil {
			t.Fatalf("expected non-empty product id")
		}

		if created["date_created"] == "" || created["date_created"] == nil {
			t.Fatal("expected non-empty product date_created")
		}
		if created["date_updated"] == "" || created["date_updated"] == nil {
			t.Fatal("expected non-empty product date_updated")
		}

		want := map[string]interface{}{
			"id":           created["id"],
			"date_created": created["date_created"],
			"date_updated": created["date_updated"],
			"name":         "product0",
			"cost":         float64(55),
			"quantity":     float64(6),
			"sold":         float64(0),
			"revenue":      float64(0),
			"user_id":      tests.AdminID,
		}

		if diff := cmp.Diff(want, created); diff != "" {
			t.Fatalf("Response did not match expected. Diff:\n%s", diff)
		}
	}

	{
		// read
		url := fmt.Sprintf("/v1/products/%s", created["id"])
		req := httptest.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+p.adminToken)

		p.app.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("reading: expected status code %v, got %v", http.StatusOK, resp.Code)
		}

		var fetched map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf("decoding: %s", err)
		}

		if diff := cmp.Diff(created, fetched); diff != "" {
			t.Fatalf("Retrieved product should match create. Diff:\n%s", diff)
		}
	}
	{
		// update
		body := strings.NewReader(`{"name":"new name","cost":20,"quantity":10}`)
		url := fmt.Sprintf("/v1/products/%s", created["id"])
		req := httptest.NewRequest("PUT", url, body)
		resp := httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+p.adminToken)

		p.app.ServeHTTP(resp, req)

		if http.StatusNoContent != resp.Code {
			t.Fatalf("updating: expected status code %v, got %v", http.StatusNoContent, resp.Code)
		}

		req = httptest.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+p.adminToken)
		req.Header.Set("Content-Type", "application/json")
		resp = httptest.NewRecorder()

		p.app.ServeHTTP(resp, req)

		if http.StatusOK != resp.Code {
			t.Fatalf("getting: expected status code %v, got %v", http.StatusOK, resp.Code)
		}

		var updated map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf("decoding: %s", err)
		}

		want := map[string]interface{}{
			"id":           created["id"],
			"name":         "new name",
			"cost":         float64(20),
			"quantity":     float64(10),
			"sold":         float64(0),
			"revenue":      float64(0),
			"user_id":      tests.AdminID,
			"date_created": created["date_created"],
			"date_updated": updated["date_updated"],
		}

		if diff := cmp.Diff(want, updated); diff != "" {
			t.Fatalf("Retrieved product should match created. Diff:\n%s", diff)
		}
	}

	{
		// delete
		url := fmt.Sprintf("/v1/products/%s", created["id"])
		req := httptest.NewRequest("DELETE", url, nil)
		resp := httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+p.adminToken)

		p.app.ServeHTTP(resp, req)

		if http.StatusNoContent != resp.Code {
			t.Fatalf("deleting: expected status code %v, got %v", http.StatusNoContent, resp.Code)
		}

		req = httptest.NewRequest("GET", url, nil)
		resp = httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+p.adminToken)

		p.app.ServeHTTP(resp, req)

		if http.StatusNotFound != resp.Code {
			t.Fatalf("get product: expected status code %v, got %v", http.StatusNotFound, resp.Code)
		}
	}
}

func (p *ProductTests) AddSale(t *testing.T) {
	body := strings.NewReader(`{"quantity":3,"paid":5}`)
	productID := "a2b0639f-2cc6-44b8-b97b-15d69dbb511e"

	req := httptest.NewRequest("POST", "/v1/products/"+productID+"/sales", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+p.adminToken)

	p.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("add sale: expected status code %v, got %v", http.StatusCreated, resp.Code)
	}

	var created map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decoding: %s", err)
	}

	want := map[string]interface{}{
		"id":           created["id"],
		"product_id":   created["product_id"],
		"quantity":     float64(3),
		"paid":         float64(5),
		"date_created": created["date_created"],
	}

	if diff := cmp.Diff(want, created); diff != "" {
		t.Fatalf("Response did not match expected. Diff:\n%s", diff)
	}
}

func (p *ProductTests) SalesList(t *testing.T) {
	productID := "a2b0639f-2cc6-44b8-b97b-15d69dbb511e"
	req := httptest.NewRequest("GET", "/v1/products/"+productID+"/sales", nil)
	resp := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+p.adminToken)

	p.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("sales list: expected status code %v, got %v", http.StatusOK, resp.Code)
	}

	var list []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decoding: %s", err)
	}

	want := []map[string]interface{}{
		{
			"id":           "98b6d4b8-f04b-4c79-8c2e-a0aef46854b7",
			"product_id":   productID,
			"quantity":     float64(2),
			"paid":         float64(100),
			"date_created": "2019-01-01T00:00:03.000001Z",
		},
		{
			"id":           "85f6fb09-eb05-4874-ae39-82d1a30fe0d7",
			"product_id":   productID,
			"quantity":     float64(5),
			"paid":         float64(250),
			"date_created": "2019-01-01T00:00:04.000001Z",
		},
	}

	if diff := cmp.Diff(want, list); diff != "" {
		t.Fatalf("Response did not match expected. Diff:\n%s", diff)
	}
}

func (p *ProductTests) CreateRequiresFields(t *testing.T) {
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest("POST", "/v1/products", body)
	resp := httptest.NewRecorder()

	req.Header.Set("Authorization", "Bearer "+p.adminToken)

	p.app.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("getting: expected status code %v, got %v", http.StatusBadRequest, resp.Code)
	}
}
