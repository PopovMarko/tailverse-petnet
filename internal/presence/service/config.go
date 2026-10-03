package presence_service

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type PresenceConfig struct {
	TTL time.Duration `envconfig:"TTL" default:"2h"`
}

func NewPresenceConfigMust() PresenceConfig {
	var config PresenceConfig
	if err := envconfig.Process("PRESENCE", &config); err != nil {
		panic(fmt.Errorf("process presence config: %w", err))
	}
	return config
}
