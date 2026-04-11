package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Env  string `envconfig:"APP_ENV" default:"development"`
	Port string `envconfig:"APP_PORT" default:"8080"`
}

type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD"`
	Name     string `envconfig:"DB_NAME" default:"banking"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
}

type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     string `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD"`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

func ProvideConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg.App); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg.Redis); err != nil {
		return nil, err
	}
	return &cfg, nil
}
