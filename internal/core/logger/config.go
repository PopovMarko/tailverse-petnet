package core_logger

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type LoggerConfig struct {
	Level  string `envconfig:"LEVEL" default:"DEBUG"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

func NewLoggerConfig() (LoggerConfig, error) {
	var config LoggerConfig
	if err := envconfig.Process("LOGGER", &config); err != nil {
		return LoggerConfig{}, fmt.Errorf("process logger config: %w", err)
	}
	return config, nil
}

func NewLoggerConfigMust() LoggerConfig {
	config, err := NewLoggerConfig()
	if err != nil {
		panic(fmt.Errorf("get logger config: %w", err))
	}
	return config
}
