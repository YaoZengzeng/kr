package memory

import (
	"testing"
	"reflect"

	"github.com/YaoZengzeng/kr/types"
)

func TestRegistery(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("create new registry failed: %v", err)
	}

	service := &types.Service{
		Address:	"localhost",
		Port:		8080,
		Endpoint:	"/webhook",
	}

	err = registry.Register(service)
	if err != nil {
		t.Fatalf("register service failed: %v", err)
	}

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
