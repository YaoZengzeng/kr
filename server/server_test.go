package server

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func TestServer(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(s.HandleRegister))
	defer ts.Close()

	res, err := http.PostForm(ts.URL, url.Values{"address": {"localhost"}, "port": {"8080"}, "endpoint": {"/webhook"}})
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("register service failed")
	}
}
