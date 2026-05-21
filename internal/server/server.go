package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Server *http.Server
}

func New(port string, r *chi.Mux) *Server {
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
