package config

import (
	"fmt"
	"os"

	"github.com/Moranilt/http-utils/clients/database"
	"github.com/spf13/viper"
)

const (
	ENV_PRODUCTION = "PRODUCTION"
	ENV_PORT       = "PORT"

	ENV_DB_NAME     = "DB_NAME"
	ENV_DB_HOST     = "DB_HOST"
	ENV_DB_USER     = "DB_USER"
	ENV_DB_PASSWORD = "DB_PASSWORD"
	ENV_DB_SSL_MODE = "DB_SSL_MODE"

	ENV_TRACER_URL  = "TRACER_URL"
	ENV_TRACER_NAME = "TRACER_NAME"
)

var envVariables []string = []string{
	ENV_PORT,
	ENV_DB_NAME,
	ENV_DB_HOST,
	ENV_DB_USER,
	ENV_DB_PASSWORD,
	ENV_DB_SSL_MODE,
	ENV_TRACER_URL,
	ENV_TRACER_NAME,
}

type TracerConfig struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
}

type Config struct {
	Tracer     *TracerConfig
	DB         *database.Credentials
	Port       string
	Production bool
}

func Read() (*Config, error) {
	var envCfg Config
	viper.AutomaticEnv()
	isProduction := viper.GetBool(ENV_PRODUCTION)

	result := make(map[string]string, len(envVariables))
	for _, name := range envVariables {
		value := os.Getenv(name)
		if value == "" {
			return nil, fmt.Errorf("env %q is empty", name)
		}
		result[name] = value
	}

	envCfg = Config{
		DB: &database.Credentials{
			Username: result[ENV_DB_USER],
			Password: result[ENV_DB_PASSWORD],
			DBName:   result[ENV_DB_NAME],
			Host:     result[ENV_DB_HOST],
			SSLMode:  result[ENV_DB_SSL_MODE],
		},
		Tracer: &TracerConfig{
			URL:  result[ENV_TRACER_URL],
			Name: result[ENV_TRACER_NAME],
		},
		Port:       result[ENV_PORT],
		Production: isProduction,
	}

	return &envCfg, nil
}
