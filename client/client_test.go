package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/YaoZengzeng/kr/types"
)

func TestClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check parameters.
		r.ParseForm()
		_, ok := r.Form["address"]
		if !ok {
			http.Error(w, fmt.Sprintf("failed to parse address of service"), http.StatusBadRequest)
			return
		}

		paramPort, ok := r.Form["port"]
		if !ok {
			http.Error(w, fmt.Sprintf("failed to parse port of service"), http.StatusBadRequest)
			return
		}
		_, err := strconv.Atoi(paramPort[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert port number to int"), http.StatusBadRequest)
			return
		}

		_, ok = r.Form["endpoint"]
		if !ok {
			http.Error(w, fmt.Sprintf("failed to parse endpoint of service"), http.StatusBadRequest)
			return
		}
	}))
	defer ts.Close()

	opts := []Option{
		WithRegistry(ts.URL),
		WithHeartbeat(100 * time.Millisecond),
	}

	client, err := New(opts...)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}

	service := &types.Service{
		Address:  "localhost",
		Port:     8080,
		Endpoint: "/webhook",
	}

	// Register same service multiple times to test idempotency.
	for i := 0; i < 3; i++ {
		if err = client.Register(service); err != nil {
			t.Fatalf("register service failed: %v", err)
		}
	}

	if len(client.services) != 1 {
		t.Fatalf("the number of registered services is %d, should be 1", len(client.services))
	}

	// Deregister same service multiple times to test idempotency.
	for i := 0; i < 3; i++ {
		if err := client.Deregister(service); err != nil {
			t.Fatalf("deregister service failed: %v", err)
		}
	}

	if len(client.services) != 0 {
		t.Fatalf("the number of registered service is %d, should be 0 after deregisteration", len(client.services))
	}
}
