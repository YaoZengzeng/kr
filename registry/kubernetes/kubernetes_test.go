package kubernetes

import (
	"context"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/YaoZengzeng/kr/types"
)

func TestRegistery(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	registry, err := newRegistry(clientset, 60*time.Second, 10*time.Minute)
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
	registry, err := newRegistry(clientset, 3*time.Second, 10*time.Minute)
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

func TestKeepRegisterSerivce(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Make service expire quickly.
	registry, err := newRegistry(clientset, 3*time.Second, 10*time.Minute)
	if err != nil {
		t.Fatalf("create new registry failed: %v", err)
	}

	service := &types.Service{
		Address:  "localhost",
		Port:     8080,
		Endpoint: "/webhook",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		// Keep register service every second.
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := registry.Register(service)
				if err != nil {
					t.Fatalf("register service failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait a second to let cache populated.
	time.Sleep(1 * time.Second)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
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
		case <-ctx.Done():
			return
		}
	}
}

func TestRegisterModifiedService(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Make service expire quickly.
	registry, err := newRegistry(clientset, 3*time.Second, 10*time.Minute)
	if err != nil {
		t.Fatalf("create new registry failed: %v", err)
	}

	services := []*types.Service{
		{
			Address:  "localhost",
			Port:     8080,
			Endpoint: "/webhook",
		},
		{
			Address: "localhost",
			// Only Port has been modified.
			Port:     8081,
			Endpoint: "/webhook",
		},
	}

	for _, service := range services {
		err = registry.Register(service)
		if err != nil {
			t.Fatalf("register service failed: %v", err)
		}
	}

	// Need to wait for the local cache to be populated.
	time.Sleep(100 * time.Millisecond)

	registeredServices, err := registry.ListServices()
	if err != nil {
		t.Fatalf("list services failed: %v", err)
	}

	if len(registeredServices) != 2 {
		t.Fatalf("the number of listed services is %d, should get 2", len(registeredServices))
	}

	for _, s := range services {
		ok := false
		for _, rs := range registeredServices {
			if reflect.DeepEqual(s, rs) {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("Could not find service %v in registry", s)
		}
	}
}

func TestServiceCleanup(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// Make service expire and cleanup quickly.
	registry, err := newRegistry(clientset, 3*time.Second, 5*time.Second)
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

	// Wait service to cleanup.
	time.Sleep(8 * time.Second)

	endpoints, err := registry.lister.List(labels.SelectorFromSet(labels.Set{labelKey: labelValue}))
	if err != nil {
		t.Fatalf("list endpoints directly failed: %v", err)
	}

	if len(endpoints) != 0 {
		t.Fatalf("the number of underlying endpoints is %d, should get 0, because it get expired and cleanup", len(endpoints))
	}
}
