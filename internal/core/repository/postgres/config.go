package core_postgres

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type PostgresConfig struct {
	Host     string        `envconfig:"HOST" default:"localhost"`
	Port     int           `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER" required:"true"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	DB       string        `envconfig:"DB" required:"true"`
	SSLMode  string        `envconfig:"SSLMODE" default:"disable"`
	Timeout  time.Duration `envconfig:"TIMEOUT" default:"30s"`
}

func NewPostgresConfig() (PostgresConfig, error) {
	var config PostgresConfig
	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return PostgresConfig{}, fmt.Errorf("process postgres config: %w", err)
	}
	return config, nil
}

func NewPostgresConfigMust() PostgresConfig {
	config, err := NewPostgresConfig()
	if err != nil {
		panic(fmt.Errorf("get postgres config: %w", err))
	}
	return config
}

func (c PostgresConfig) DSN() string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:     c.DB,
		RawQuery: url.Values{"sslmode": []string{c.SSLMode}}.Encode(),
	}
	return dsn.String()
}
