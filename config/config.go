package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/viper"
)

const (
	ENV_PRODUCTION = "PRODUCTION"
	ENV_PORT       = "PORT"

	ENV_VAULT_TOKEN         = "VAULT_TOKEN"
	ENV_VAULT_HOST          = "VAULT_HOST"
	ENV_VAULT_MOUNT_PATH    = "VAULT_MOUNT_PATH"
	ENV_VAULT_DB_CREDS_PATH = "VAULT_DB_CREDS_PATH"

	ENV_TRACER_URL  = "TRACER_URL"
	ENV_TRACER_NAME = "TRACER_NAME"
)

var envVariables []string = []string{
	ENV_PORT,
	ENV_VAULT_TOKEN,
	ENV_VAULT_HOST,
	ENV_VAULT_MOUNT_PATH,
	ENV_VAULT_DB_CREDS_PATH,
	ENV_TRACER_URL,
	ENV_TRACER_NAME,
}

type VaultEnv struct {
	MountPath   string `mapstructure:"VAULT_MOUNT_PATH"`
	DbCredsPath string `mapstructure:"VAULT_DB_CREDS_PATH"`
	Token       string `mapstructure:"VAULT_TOKEN"`
	Host        string `mapstructure:"VAULT_HOST"`
}

type TracerConfig struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
}

type Config struct {
	Vault      *VaultEnv
	Tracer     *TracerConfig
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

	vaultHostUrl, err := url.Parse(result[ENV_VAULT_HOST])
	if err != nil {
		return nil, fmt.Errorf("%q has invalid value: %w", ENV_VAULT_HOST, err)
	}

	envCfg = Config{
		Vault: &VaultEnv{
			Token:       result[ENV_VAULT_TOKEN],
			MountPath:   result[ENV_VAULT_MOUNT_PATH],
			DbCredsPath: result[ENV_VAULT_DB_CREDS_PATH],
			Host:        vaultHostUrl.String(),
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
