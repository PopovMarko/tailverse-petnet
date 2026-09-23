package main

import (
	// "context"
	"fmt"
	"net/http"
	"os"
	// "os/signal"
	// "syscall"
	"time"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	// Set time zone to UTC
	var TimeZone = time.UTC
	time.Local = TimeZone

	// Init logger config and logger
	loggerConfig := core_logger.NewLoggerConfigMust()
	logger, err := core_logger.NewLogger(loggerConfig)
	if err != nil {
		fmt.Printf("Failed to configure logger: %w", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Init context with Cancel by syscall signals
	// ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer cancel()

	logger.Debug("Application time zone", zap.Any("zone", TimeZone))
	logger.Debug("Local time: ", zap.Time("Now", time.Now()))

	router := chi.NewRouter()

	if err := http.ListenAndServe(":5050", router); err != nil {
		logger.Error("Failed to start server: %w", zap.Error(err))
	}
}
