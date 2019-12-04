package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/YaoZengzeng/kr/client"
	"github.com/YaoZengzeng/kr/types"
)

const (
	registry = "http://127.0.0.1:10812/register"
)

func main() {
	opts := []client.Option{
		client.WithRegistry(registry),
		// Heartbeat to registry every 3s.
		client.WithHeartbeat(3 * time.Second),
	}
	c, err := client.New(opts...)
	if err != nil {
		log.Printf("create client failed: %v\n", err)
		os.Exit(1)
	}

	// Register ourself to registry, so server could dispatch message to us.
	if err := c.Register(&types.Service{
		Address:  "localhost",
		Port:     10813,
		Endpoint: "/message",
	}); err != nil {
		log.Printf("register service failed: %v\n", err)
		os.Exit(1)
	}

	log.Printf("register service succeeded\n")

	http.HandleFunc("/message", handleMessage)

	if err := http.ListenAndServe(":10813", nil)
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read the body of request"), http.StatusBadRequest)
		return
	}

	fmt.Printf("The message read from client:\n%s\n", string(data))
}
