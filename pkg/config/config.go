package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Specification structured configuration variables.
type Specification struct {
	Debug       bool   `envconfig:"DEBUG" default:"false"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	Environment string `envconfig:"ENVIRONMENT" default:"staging"`

	Database struct {
		DSN            string `envconfig:"PGSQL_DSN"`
		MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"file://resources/migrations"`
	}
	HTTP struct {
		BaseURL string `envconfig:"HTTP_BASE_URL"`
		Port    int    `envconfig:"HTTP_PORT" default:"8000"`
	}
	QUOTE struct {
		StocksPath    string `envconfig:"STOCKS_PATH" default:"resources/import/stocks"`
		DividendsPath string `envconfig:"DIVIDENDS_PATH" default:"resources/import/dividends"`
	}
	BANK struct {
		TransferPath string `envconfig:"STOCKS_PATH" default:"resources/import/transfers"`
	}
}

// LoadEnv load config variables into Specification.
func LoadEnv() (*Specification, error) {
	var config Specification
	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}
