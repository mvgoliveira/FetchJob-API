package main

import (
	"net/http"

	"github.com/mvgoliveira/FetchJob-API/internal/config"
	"github.com/mvgoliveira/FetchJob-API/internal/controllers"
	"github.com/mvgoliveira/FetchJob-API/internal/handlers"
	"github.com/mvgoliveira/FetchJob-API/internal/infrastructure"
	"github.com/mvgoliveira/FetchJob-API/internal/repositories"
	"github.com/mvgoliveira/FetchJob-API/internal/services"
	"github.com/rs/zerolog/log"
)

func main() {
	appConfig := config.NewAppConfig()

	database := infrastructure.NewDatabaseClient(appConfig.Database.ConnectionString)
	databaseRepository := repositories.NewDatabaseRepository(database)
	linkedInRepository := repositories.NewLinkedInRepository()

	jobsService := services.NewJobsService(databaseRepository, linkedInRepository)
	jobsController := controllers.NewJobsController(jobsService)


	router := handlers.NewRouter(jobsController)

	log.Info().Str("port", appConfig.Port).Msg("starting the application server")

	if err := http.ListenAndServe(":"+appConfig.Port, router); err != nil {
		log.Panic().Err(err).Send()
	}
}
