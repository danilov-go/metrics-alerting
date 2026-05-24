package server

import (
	"net/http"
)

type log interface {
	Info(args ...any)
}

type Server struct {
	Server *http.Server
	Logger log
}

func New(port string, l log, r http.Handler) *Server {
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}
	return &Server{
		Server: server,
		Logger: l,
	}
}

func (serv *Server) Run() error {
	serv.Logger.Info("Running server", "address", serv.Server.Addr)
	return serv.Server.ListenAndServe()
}
