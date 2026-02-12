package models

import (
	"time"

	"gorm.io/gorm"
)

type BatchType string

const (
	BatchMorning BatchType = "morning"
	BatchNoon    BatchType = "noon"
	BatchEvening BatchType = "evening"
)

type BatchLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Type      BatchType      `json:"type"` // morning, noon, evening
	Date      string         `json:"date"` // YYYY-MM-DD
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type NewsItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BatchID   uint           `json:"batch_id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Url       string         `json:"url"`
	Source    string         `json:"source"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type AnalysisType string

const (
	Analysis3Day AnalysisType = "3_day"
	Analysis7Day AnalysisType = "7_day"
)

type Analysis struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BatchID   uint           `json:"batch_id"`
	Type      AnalysisType   `json:"type"` // 3_day, 7_day
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
