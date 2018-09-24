package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestStatsHandler(t *testing.T) {

	req, _ := http.NewRequest("GET", "/stats", nil)

	resp := httptest.NewRecorder()
	handler := http.HandlerFunc(StatsHandler)

	handler.ServeHTTP(resp, req)

	if status := resp.Code; status != http.StatusOK {
		t.Errorf("status code= %v not %v", status, http.StatusOK)
	}

	expected := `{"average":0,"total":0}`
	if resp.Body.String() != expected {
		t.Errorf("handler returned %v expected %v", resp.Body.String(), expected)
	}
}

func TestGetPasswordHandlerFailure(t *testing.T) {

	req, _ := http.NewRequest("GET", "/hash/0", nil)

	resp := httptest.NewRecorder()
	handler := http.HandlerFunc(GetPasswordHandler)

	handler.ServeHTTP(resp, req)

	if status := resp.Code; status != http.StatusNotFound {
		t.Errorf("status code= %v not %v", status, http.StatusNotFound)
	}
}
func TestCreatePasswordHandler(t *testing.T) {

	params := url.Values{}
	params.Set("password", "angryMonkey")
	req, _ := http.NewRequest("POST", "/hash", strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	jobs := make(chan map[int64]string, 1)
	resp := httptest.NewRecorder()
	handler := http.HandlerFunc(CreatePasswordHandler(jobs))

	handler.ServeHTTP(resp, req)

	if status := resp.Code; status != http.StatusOK {
		t.Errorf("status code= %v not %v", status, http.StatusOK)
	}

	expected := "1"
	if resp.Body.String() != expected {
		t.Errorf("handler returned %v expected %v", resp.Body.String(), expected)
	}
}
