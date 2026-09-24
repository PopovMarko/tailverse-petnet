package main

import (
	"fmt"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_middleware "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	loggerConfig := core_logger.NewLoggerConfigMust()
	logger, err := core_logger.NewLogger(loggerConfig)
	if err != nil {
		fmt.Errorf("Failed to configure logger: %w", err)
	}
	logger.Debug("Initialized logger")

	r := chi.NewRouter()
	r.Use(core_middleware.RequestID())
	r.Use(core_middleware.CORS())
	r.Use(core_middleware.Logger(logger))
	r.Use(core_middleware.Trace())
	// TODO: Panic middle
}
