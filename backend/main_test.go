package main

import (
	"bre_new_backend/config"
	"bre_new_backend/models"
	"bre_new_backend/services"
	"testing"
)

func TestRunUpdateTask(t *testing.T) {
	config.InitConfig()
	config.InitDB()
	db := config.DB
	if err := db.AutoMigrate(&models.BatchLog{}, &models.NewsItem{}, &models.Analysis{}); err != nil {
		t.Fatalf("Failed to migrate sqlite db: %v", err)
	}

	services.RunUpdateTask()
}
