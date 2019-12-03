package kubernetes

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/YaoZengzeng/kr/types"
)

func TestRegistery(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	registry, err := newRegistry(clientset, 60 * time.Second)
	if err != nil {
		t.Fatalf("create new registry failed: %v", err)
	}

	service := &types.Service{
		Address:  "localhost",
		Port:     8080,
		Endpoint: "/webhook",
	}

	err = registry.Register(service)
	if err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	// Need to wait for the local cache to be populated.
	time.Sleep(100 * time.Millisecond)

	services, err := registry.ListServices()
	if err != nil {
		t.Fatalf("list services failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("the number of listed services is %d, should get 1", len(services))
	}

	if !reflect.DeepEqual(service, services[0]) {
		t.Fatalf("the content of service changed after register")
	}
}

func TestServiceExpire(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Make service expire quickly.
	registry, err := newRegistry(clientset, 3 * time.Second)
	if err != nil {
		t.Fatalf("create new registry failed: %v", err)
	}

	service := &types.Service{
		Address: 	"localhost",
		Port:		8080,
		Endpoint:	"/webhook",
	}

	err = registry.Register(service)
	if err != nil {
		t.Fatalf("register service failed: %v", err)
	}

	// Need to wait for the local cache to be populated.
	time.Sleep(100 * time.Millisecond)

	services, err := registry.ListServices()
	if err != nil {
		t.Fatalf("list services failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("the number of listed services is %d, should get 1", len(services))
	}

	time.Sleep(5 * time.Second)

	services, err = registry.ListServices()
	if err != nil {
		t.Fatalf("list services failed: %v", err)
	}

	if len(services) != 0 {
		t.Fatalf("the number of listed services is %d, should get 0, because it should be expired", len(services))
	}

	err = registry.Register(service)
	if err != nil {
		t.Fatalf("register service again failed: %v", err)
	}

	// Also need to wait for the local cache to be populated.
	time.Sleep(100 * time.Millisecond)

	services, err = registry.ListServices()
	if err != nil {
		t.Fatalf("list services failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("the number of listed services is %d, should get 1", len(services))
	}

	if !reflect.DeepEqual(service, services[0]) {
		t.Fatalf("the content of service changed after register")
	}
}
