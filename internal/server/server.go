package server

import (
	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
	"github.com/go-chi/chi/v5"
)

const defaultPort = "localhost:8080"

type Server struct {
	Server *http.Server
}

func New(port string, s *repository.MemStorage) *Server {
	if port == "" {
		port = defaultPort
	}
	r := chi.NewRouter()
	r.Post("/update/{mType}/{mName}/{mVal}", handler.PostMetricsHandler(s))
	r.Get("/value/{mType}/{mName}", handler.GetMetricHandler(s))
	r.Get("/", handler.ExposeMetricsHandler(s))
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}
	return &Server{
		Server: server,
	}
}

func (serv *Server) Run() error {
	return serv.Server.ListenAndServe()
}
