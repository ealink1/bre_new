package config

import (
	"fmt"
	"log"
	"os"

	"bre_new_backend/models"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var AppConfig Config

type Config struct {
	Mysql struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"mysql"`
	AI struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"ai"`
	System struct {
		Port string `yaml:"port"`
	}
}

func InitConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config.yaml: %v", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Fatalf("Error parsing config.yaml: %v", err)
	}
}

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		AppConfig.Mysql.User,
		AppConfig.Mysql.Password,
		AppConfig.Mysql.Host,
		AppConfig.Mysql.Port,
		AppConfig.Mysql.Database,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto Migrate
	err = DB.AutoMigrate(&models.BatchLog{}, &models.NewsItem{}, &models.Analysis{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
