package server

import (
	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
)

type Server struct {
	Server *http.Server
}

func New(s *repository.MemStorage) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.MetricsHandler(s))
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	return &Server{
		Server: server,
	}
}

func (serv *Server) Run() error {
	return serv.Server.ListenAndServe()
}
