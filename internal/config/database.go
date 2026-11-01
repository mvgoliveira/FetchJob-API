package config

import (
	"github.com/kelseyhightower/envconfig"
	customerrors "github.com/mvgoliveira/FetchJob-API/internal/errors"
	"github.com/rs/zerolog/log"
)


type Database struct {
	ConnectionString string `json:"connection_string" split_words:"true"`
}

var databaseConfigStrategyMap = map[string]func() *Database{
	"development": databaseConfigFromEnvironment,
	"production":  databaseConfigFromEnvironment,
}

func databaseConfigFromEnvironment() *Database {
	var cfg Database

	if err := envconfig.Process("DATABASE", &cfg); err != nil {
		log.Fatal().Str("context", "database").Err(err).Msg(customerrors.ErrProcessingEnvironmentIntoConfiguration.Error())
	}

	return &cfg
}

func NewDatabaseConfig(environment string) *Database {
	strategy := databaseConfigStrategyMap[environment]
	return strategy()
}
