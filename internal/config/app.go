package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type App struct {
	Environment string `default:"development"`
	Port        string `default:"8080"`
	Database    *Database
}

func appConfigFromEnvironment() *App {
	var cfg App

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal().Err(err).Msg("error processing environment variables")
	}

	return &cfg
}

func NewAppConfig() *App {
	cfg := appConfigFromEnvironment()

	cfg.Database = NewDatabaseConfig(cfg.Environment)

	return cfg
}
