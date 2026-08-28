// Package server реализует HTTP-сервер приложения.
package server

import (
	"net/http"
)

type log interface {
	Infow(msg string, keysAndValues ...any)
}

// Server определяет конфигурацию HTTP-сервера.
type Server struct {
	Server *http.Server
	Logger log
}

// New создает новый экземпляр Server.
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

// Run запускает HTTP-сервер.
func (serv *Server) Run() error {
	serv.Logger.Infow("Running server", "address", serv.Server.Addr)
	return serv.Server.ListenAndServe()
}
