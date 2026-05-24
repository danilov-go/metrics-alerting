package server

import (
	"net/http"

	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"go.uber.org/zap"
)

type Server struct {
	Server *http.Server
}

func New(port string, r http.Handler) *Server {
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}
	return &Server{
		Server: server,
	}
}

func (serv *Server) Run() error {
	logger.Log.Info("Running server", zap.String("address", serv.Server.Addr))
	return serv.Server.ListenAndServe()
}
