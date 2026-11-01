package repositories

import (
	"strings"

	customerrors "github.com/mvgoliveira/FetchJob-API/internal/errors"
	"github.com/mvgoliveira/FetchJob-API/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type DatabaseRepository interface {
	GetJobs(filter models.JobFilter) ([]models.Job, int64, error)
}

type DatabaseRepositoryImpl struct {
	db *gorm.DB
}

func (r *DatabaseRepositoryImpl) GetJobs(filter models.JobFilter) ([]models.Job, int64, error) {
	var jobs []models.Job
	var total int64

	query := r.db.Model(&models.Job{})

	if search := strings.TrimSpace(filter.Search); search != "" {
		likeSearch := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR company ILIKE ? OR short_description ILIKE ?",
			likeSearch, likeSearch, likeSearch,
		)
	}

	if company := strings.TrimSpace(filter.Company); company != "" {
		query = query.Where("company ILIKE ?", "%"+company+"%")
	}

	if location := strings.TrimSpace(filter.Location); location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	if workModel := strings.TrimSpace(filter.WorkModel); workModel != "" {
		query = query.Where("LOWER(work_model) = LOWER(?)", workModel)
	}

	if filter.IsRemote != nil {
		query = query.Where("is_remote = ?", *filter.IsRemote)
	}

	if seniority := strings.TrimSpace(filter.Seniority); seniority != "" {
		query = query.Where("LOWER(seniority) = LOWER(?)", seniority)
	}

	if tag := strings.TrimSpace(filter.Tag); tag != "" {
		query = query.Where(
			"EXISTS (SELECT 1 FROM jsonb_array_elements_text(tags) AS t WHERE t ILIKE ?)",
			tag,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		log.Error().Str("method", "DatabaseRepository.GetJobs.Count").Err(err).Send()
		return nil, 0, customerrors.ErrWhileQueryingOnDatabase
	}

	offset := (filter.Page - 1) * filter.Limit

	if err := query.Order("created_at DESC").Limit(filter.Limit).Offset(offset).Find(&jobs).Error; err != nil {
		log.Error().Str("method", "DatabaseRepository.GetJobs.Find").Err(err).Send()
		return nil, 0, customerrors.ErrWhileQueryingOnDatabase
	}

	return jobs, total, nil
}

func NewDatabaseRepository(db *gorm.DB) DatabaseRepository {
	return &DatabaseRepositoryImpl{
		db: db,
	}
}

