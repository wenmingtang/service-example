package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	var u struct {
		Name string `validate:"required"`
	}

	body := strings.NewReader(`{}`)

	r := httptest.NewRequest("POST", "/", body)
	err := Decode(r, &u)

	if err == nil {
		t.Errorf("decode with missing argments should return an error but returned nil")
	}

	t.Log(err)
}
