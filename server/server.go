package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/YaoZengzeng/kr/registry"
	"github.com/YaoZengzeng/kr/types"
)

type Server struct {
	Registry registry.Registry
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	paramAddress, ok := r.Form["address"]
	if !ok {
		http.Error(w, fmt.Sprintf("failed to parse address of service"), http.StatusBadRequest)
		return
	}

	paramPort, ok := r.Form["port"]
	if !ok {
		http.Error(w, fmt.Sprintf("failed to parse port of service"), http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(paramPort[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to convert port number to int"), http.StatusBadRequest)
		return
	}

	paramEndpoint, ok := r.Form["endpoint"]
	if !ok {
		http.Error(w, fmt.Sprintf("failed to parse endpoint of service"), http.StatusBadRequest)
		return
	}

	err = s.Registry.Register(&types.Service{
		Address:  paramAddress[0],
		Port:     port,
		Endpoint: paramEndpoint[0],
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to register service"), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func New(registry registry.Registry) (*Server, error) {
	return &Server{
		Registry: registry,
	}, nil
}

func (s *Server) Run() error {
	http.HandleFunc("/register", s.HandleRegister)

	log.Printf("start serving request...")

	return http.ListenAndServe(":10812", nil)
}
