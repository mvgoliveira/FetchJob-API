package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mvgoliveira/FetchJob-API/internal/controllers"
	"github.com/mvgoliveira/FetchJob-API/internal/middlewares"
)

func NewRouter(jobsController *controllers.JobsController) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middlewares.AccessLog(), gin.Recovery())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "SERVING",
		})
	})

	api := router.Group("/api/v1")

	alerts := api.Group("/job")
	alerts.Use(middlewares.RequesterID())
	alerts.GET("/find", jobsController.FindJobs)
	alerts.GET("/linkedin", jobsController.SearchLinkedInJobs)

	return router
}

