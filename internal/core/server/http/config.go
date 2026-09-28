package core_server_http

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type HttpServerConfig struct {
	Address         string        `envconfig:"ADDRESS" default:"127.0.0.1"`
	Port            int           `envconfig:"PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"15s"`
	IdleTimeout     time.Duration `envconfig:"IDLE_TIMEOUT" default:"60s"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s"`
}

func NewHttpServerConfigMust() HttpServerConfig {
	var config HttpServerConfig
	if err := envconfig.Process("HTTP", &config); err != nil {
		panic(fmt.Errorf("process http server config: %w", err))
	}
	return config
}
