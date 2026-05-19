package server

import (
	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
)

const defaultPort = "localhost:8080"

type Server struct {
	Server *http.Server
}

func New(port string, s *repository.MemStorage) *Server {
	if port == "" {
		port = defaultPort
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.MetricsHandler(s))
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	return &Server{
		Server: server,
	}
}

func (serv *Server) Run() error {
	return serv.Server.ListenAndServe()
}
