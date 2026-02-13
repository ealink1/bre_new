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

// TrafficStat 用于记录网站访问统计
type TrafficStat struct {
	IP        string    `gorm:"primaryKey;size:50" json:"ip"` // IP 地址，特殊值 "TOTAL_PV" 表示总访问量
	Count     int64     `json:"count"`                        // 访问次数
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SiteCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Sort      int            `json:"sort"`
	Sites     []SiteItem     `gorm:"foreignKey:CategoryID" json:"sites"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type SiteItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CategoryID  uint           `json:"category_id"`
	Name        string         `json:"name"`
	Url         string         `json:"url"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	Sort        int            `json:"sort"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type AdminUser struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string         `gorm:"size:255" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type AdminSession struct {
	Token     string    `gorm:"primaryKey;size:128" json:"token"`
	UserID    uint      `gorm:"index" json:"user_id"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
