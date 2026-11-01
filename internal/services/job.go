package services

import (
	"context"
	"math"

	"github.com/mvgoliveira/FetchJob-API/internal/models"
	"github.com/mvgoliveira/FetchJob-API/internal/repositories"
)

type JobsService interface {
	FindJobs(filter models.JobFilter) (models.PaginatedJobsResponse, error)
	SearchLinkedInJobs(ctx context.Context, params models.LinkedInSearchParams) ([]models.Job, error)
}

type jobsService struct {
	jobsRepository     repositories.DatabaseRepository
	linkedInRepository repositories.LinkedInRepository
}

func (s *jobsService) FindJobs(filter models.JobFilter) (models.PaginatedJobsResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}

	jobs, total, err := s.jobsRepository.GetJobs(filter)
	if err != nil {
		return models.PaginatedJobsResponse{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	return models.PaginatedJobsResponse{
		Data:       jobs,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *jobsService) SearchLinkedInJobs(ctx context.Context, params models.LinkedInSearchParams) ([]models.Job, error) {
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 25 {
		params.Limit = 25
	}

	if params.Start < 0 {
		params.Start = 0
	}

	return s.linkedInRepository.SearchJobs(ctx, params)
}

func NewJobsService(
	jobsRepository repositories.DatabaseRepository,
	linkedInRepository repositories.LinkedInRepository,
) JobsService {
	return &jobsService{
		jobsRepository:     jobsRepository,
		linkedInRepository: linkedInRepository,
	}
}