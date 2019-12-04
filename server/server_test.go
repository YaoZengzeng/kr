package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/YaoZengzeng/kr/registry/memory"
)

func TestServer(t *testing.T) {
	registry, err := memory.NewRegistry()
	if err != nil {
		return nil, fmt.Errorf("create new registry failed: %v", err)
	}

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
