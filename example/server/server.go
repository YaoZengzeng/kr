package main

import (
	"log"
	"os"
	"time"
	"fmt"
	"net/http"
	"strings"

	"github.com/YaoZengzeng/kr/registry"
	"github.com/YaoZengzeng/kr/registry/kubernetes"
	"github.com/YaoZengzeng/kr/server"
)

func main() {
	// Set environment here, just for convenience of testing.
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.247.0.1")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")

	registry, err := kubernetes.NewRegistry()
	if err != nil {
		log.Printf("create registry based on kubernetes failed: %v\n", err)
		os.Exit(1)
	}

	// Send messages to all services regularly.
	go dispatcher(registry)

	server, err := server.New(registry)
	if err != nil {
		log.Printf("create registry server failed: %v\n", err)
		os.Exit(1)
	}

	err = server.Run()
	if err != nil {
		log.Printf("run registry server failed: %v\n", err)
		os.Exit(1)
	}
}

func dispatcher(registry registry.Registry) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		services, err := registry.ListServices()
		if err != nil {
			log.Printf("list services failed in dispatcher(): %v\n", err)
		}

		message := fmt.Sprintf("message in timestamp %v", time.Now())

		for _, service := range services {
			url := fmt.Sprintf("http://%s:%d%s", service.Address, service.Port, service.Endpoint)
			go func(url string){
				resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(message))
				if err != nil {
					log.Printf("dispatch message to %v failed: %v\n", url, err)
				}
				if resp != nil && resp.StatusCode != http.StatusOK {
					log.Printf("the status code of dispatching message is %v\n", resp.StatusCode)
				}
			}(url)
		}
	}
}
