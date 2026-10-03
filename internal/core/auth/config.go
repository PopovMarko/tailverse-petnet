package core_auth

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type AuthConfig struct {
	JWTSecret  string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTTL  time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"REFRESH_TTL" default:"720h"`
}

func NewAuthConfig() (AuthConfig, error) {
	var config AuthConfig
	if err := envconfig.Process("AUTH", &config); err != nil {
		return AuthConfig{}, fmt.Errorf("process auth config: %w", err)
	}
	if len(config.JWTSecret) < 32 {
		return AuthConfig{}, fmt.Errorf("AUTH_JWT_SECRET must be at least 32 characters")
	}
	return config, nil
}

func NewAuthConfigMust() AuthConfig {
	config, err := NewAuthConfig()
	if err != nil {
		panic(fmt.Errorf("get auth config: %w", err))
	}
	return config
}
