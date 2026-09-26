package main

import (
	"log"

	"github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/middleware"
	"github.com/PopovMarko/tailverse-petnet/internal/pet/transport/http"
	"github.com/go-chi/chi/v5"
)

func main() {
	loggerConfig := core_logger.NewLoggerConfigMust()
	logger, err := core_logger.NewLogger(loggerConfig)
	if err != nil {
		log.Fatal("Failed to configure logger: %s", err.Error())
	}
	logger.Debug("Initialized logger")

	r := chi.NewRouter()
	r.Use(core_http_middleware.RequestID())
	r.Use(core_http_middleware.CORS())
	r.Use(core_http_middleware.Logger(logger))
	r.Use(core_http_middleware.Trace())
	r.Use(core_http_middleware.Panic())

	var service pet_transport_http.PetService
	r.Mount("/pets", pet_transport_http.NewPetsRouter(pet_transport_http.NewPetHttpHandler(service)))
}
