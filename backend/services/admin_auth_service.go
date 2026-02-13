package services

import (
	"bre_new_backend/models"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SetupAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func EnsureAdminUser(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	var count int64
	if err := db.Model(&models.AdminUser{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || password == "" {
		return nil
	}

	_, err := CreateAdminUser(db, username, password)
	return err
}

func CreateAdminUser(db *gorm.DB, username, password string) (*models.AdminUser, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	if username == "" || password == "" {
		return nil, errors.New("username or password empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.AdminUser{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func ValidateAdminLogin(db *gorm.DB, username, password string) (*models.AdminUser, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	var user models.AdminUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}

func CreateAdminSession(db *gorm.DB, userID uint, ttl time.Duration) (*models.AdminSession, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	session := models.AdminSession{
		Token:     base64.RawURLEncoding.EncodeToString(tokenBytes),
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := db.Create(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func DeleteAdminSession(db *gorm.DB, token string) error {
	if db == nil {
		return errors.New("db is nil")
	}
	return db.Where("token = ?", token).Delete(&models.AdminSession{}).Error
}

func SetupAdminUser(db *gorm.DB, req SetupAdminRequest) (*models.AdminUser, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	var count int64
	if err := db.Model(&models.AdminUser{}).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("already initialized")
	}
	return CreateAdminUser(db, req.Username, req.Password)
}

func GetAdminSessionUser(db *gorm.DB, token string) (*models.AdminUser, *models.AdminSession, error) {
	if db == nil {
		return nil, nil, errors.New("db is nil")
	}
	if token == "" {
		return nil, nil, errors.New("token empty")
	}

	var session models.AdminSession
	if err := db.Where("token = ?", token).First(&session).Error; err != nil {
		return nil, nil, err
	}
	if time.Now().After(session.ExpiresAt) {
		_ = DeleteAdminSession(db, token)
		return nil, nil, errors.New("session expired")
	}

	var user models.AdminUser
	if err := db.Where("id = ?", session.UserID).First(&user).Error; err != nil {
		return nil, nil, err
	}
	return &user, &session, nil
}
