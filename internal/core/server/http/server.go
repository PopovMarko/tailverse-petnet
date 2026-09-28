package core_server_http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"go.uber.org/zap"
)

type HttpServer struct {
	server *http.Server
	config HttpServerConfig
	logger *core_logger.Logger
}

func NewHttpServer(config HttpServerConfig, handler http.Handler, logger *core_logger.Logger) *HttpServer {
	return &HttpServer{
		server: &http.Server{
			Addr:         net.JoinHostPort(config.Address, strconv.Itoa(config.Port)),
			Handler:      handler,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		config: config,
		logger: logger,
	}
}

// Run serves until ctx is cancelled, then shuts the server down gracefully.
func (s *HttpServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("http server started", zap.String("address", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
	}

	s.logger.Info("http server shutting down", zap.Duration("timeout", s.config.ShutdownTimeout))
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}

// RegisterOnShutdown runs f when the server starts shutting down; used to close hijacked WebSocket connections.
func (s *HttpServer) RegisterOnShutdown(f func()) {
	s.server.RegisterOnShutdown(f)
}
