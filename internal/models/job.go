package models

import (
	"time"
)

type WorkModel string

const (
	WorkModelRemote WorkModel = "remote"
	WorkModelHybrid WorkModel = "hybrid"
	WorkModelOnsite WorkModel = "onsite"
)

type Job struct {
	ID               uint      `json:"id" gorm:"primaryKey;column:id"`
	Name             string    `json:"name" gorm:"column:name;not null"`
	Company          string    `json:"company" gorm:"column:company;not null"`
	Url              string    `json:"url" gorm:"column:url;not null;uniqueIndex"`
	SalaryRange      *string   `json:"salary_range,omitempty" gorm:"column:salary_range"`
	ShortDescription string    `json:"short_description" gorm:"column:short_description;not null"`
	Location         string    `json:"location" gorm:"column:location"`
	WorkModel        WorkModel `json:"work_model" gorm:"column:work_model"`
	IsRemote         bool      `json:"is_remote" gorm:"column:is_remote;default:false"`
	Seniority        *string   `json:"seniority,omitempty" gorm:"column:seniority"`
	Tags             []string  `json:"tags" gorm:"column:tags;type:jsonb;serializer:json"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Job) TableName() string {
	return "jobs"
}

type JobFilter struct {
	Search    string `form:"search"`
	Company   string `form:"company"`
	Location  string `form:"location"`
	WorkModel string `form:"work_model"`
	IsRemote  *bool  `form:"is_remote"`
	Seniority string `form:"seniority"`
	Tag       string `form:"tag"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
}

type PaginatedJobsResponse struct {
	Data       []Job `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

type LinkedInSearchParams struct {
	Keywords  string `form:"keywords" binding:"required"`
	Location  string `form:"location"`
	WorkModel string `form:"work_model"`
	IsRemote  *bool  `form:"is_remote"`
	Limit     int    `form:"limit,default=10"`
	Start     int    `form:"start,default=0"`
}


