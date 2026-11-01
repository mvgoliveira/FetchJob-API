package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customerrors "github.com/mvgoliveira/FetchJob-API/internal/errors"
	"github.com/mvgoliveira/FetchJob-API/internal/models"
	"github.com/mvgoliveira/FetchJob-API/internal/services"
)

type JobsController struct {
	service services.JobsService
}

func NewJobsController(service services.JobsService) *JobsController {
	return &JobsController{
		service: service,
	}
}

func (c *JobsController) FindJobs(ctx *gin.Context) {
	var filter models.JobFilter

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.ErrInvalidQueryParameters.Error(),
		})
		return
	}

	response, err := c.service.FindJobs(filter)
	if err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *JobsController) SearchLinkedInJobs(ctx *gin.Context) {
	var params models.LinkedInSearchParams

	if err := ctx.ShouldBindQuery(&params); err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.ErrInvalidQueryParameters.Error(),
		})
		return
	}

	jobs, err := c.service.SearchLinkedInJobs(ctx.Request.Context(), params)
	if err != nil {
		_ = ctx.Error(err)
		ctx.JSON(http.StatusBadGateway, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  jobs,
		"total": len(jobs),
	})
}