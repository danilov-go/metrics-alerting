// Package server реализует HTTP-сервер приложения.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"
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
	if err := serv.Server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}

// Stop останавливает HTTP-сервер.
func (serv *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serv.Server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
