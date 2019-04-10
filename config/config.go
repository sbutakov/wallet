package config

import (
	"github.com/kelseyhightower/envconfig"
	"gitlab.atuko.ru/lib/errors"

	"github.com/sbutakov/wallet/pkg/account"
	"github.com/sbutakov/wallet/pkg/postgres"
)

// Config service configuration
type Config struct {
	Service struct {
		ListenAddress string
	}

	Account  account.Config
	Postgres postgres.Config
}

// LoadConfigFromEnv load configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{}
	if err := envconfig.Process("service", &config.Service); err != nil {
		return nil, errors.Wrap(err, "error on parse config")
	}

	if err := envconfig.Process("account", &config.Account); err != nil {
		return nil, errors.Wrap(err, "error on parse config")
	}

	if err := envconfig.Process("postgres", &config.Postgres); err != nil {
		return nil, errors.Wrap(err, "error on parse config")
	}
	return config, nil
}
