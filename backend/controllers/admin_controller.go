package controllers

import (
	"bre_new_backend/config"
	"bre_new_backend/models"
	"bre_new_backend/services"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func AdminSetup(c *gin.Context) {
	var req services.SetupAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	user, err := services.SetupAdminUser(config.DB, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	session, err := services.CreateAdminSession(config.DB, user.ID, 30*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "create session failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"token": session.Token,
			"user":  gin.H{"id": user.ID, "username": user.Username},
		},
	})
}

func AdminLogin(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}

	user, err := services.ValidateAdminLogin(config.DB, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid credentials"})
		return
	}

	session, err := services.CreateAdminSession(config.DB, user.ID, 30*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "create session failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"token": session.Token,
			"user":  gin.H{"id": user.ID, "username": user.Username},
		},
	})
}

func AdminLogout(c *gin.Context) {
	sessionAny, ok := c.Get("adminSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	session, _ := sessionAny.(*models.AdminSession)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}

	_ = services.DeleteAdminSession(config.DB, session.Token)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

type AdminUserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminUserPasswordRequest struct {
	Password string `json:"password"`
}

func AdminUserList(c *gin.Context) {
	var rows []models.AdminUser
	if err := config.DB.Order("id asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminUserCreate(c *gin.Context) {
	var req AdminUserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	user, err := services.CreateAdminUser(config.DB, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": gin.H{"id": user.ID, "username": user.Username}})
}

func AdminUserSetPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	var req AdminUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}

	var user models.AdminUser
	if err := config.DB.Where("id = ?", uint(id)).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "update failed"})
		return
	}
	if err := config.DB.Model(&models.AdminUser{}).Where("id = ?", user.ID).Update("password_hash", string(hash)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "update failed"})
		return
	}
	_ = config.DB.Where("user_id = ?", user.ID).Delete(&models.AdminSession{}).Error
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

func AdminUserDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}

	currentAny, ok := c.Get("adminUser")
	if ok {
		if current, ok := currentAny.(*models.AdminUser); ok && current != nil {
			if current.ID == uint(id) {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "cannot delete current user"})
				return
			}
		}
	}

	tx := config.DB.Begin()
	if err := tx.Where("user_id = ?", uint(id)).Delete(&models.AdminSession{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	if err := tx.Where("id = ?", uint(id)).Delete(&models.AdminUser{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

type SiteCategoryUpsertRequest struct {
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

func AdminSiteCategoryList(c *gin.Context) {
	var rows []models.SiteCategory
	if err := config.DB.Order("sort asc, id asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminSiteCategoryCreate(c *gin.Context) {
	var req SiteCategoryUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	row := models.SiteCategory{Name: req.Name, Sort: req.Sort}
	if err := config.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "create failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": row})
}

func AdminSiteCategoryUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	var req SiteCategoryUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["sort"] = req.Sort
	if err := config.DB.Model(&models.SiteCategory{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

func AdminSiteCategoryDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	tx := config.DB.Begin()
	if err := tx.Where("category_id = ?", uint(id)).Delete(&models.SiteItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	if err := tx.Where("id = ?", uint(id)).Delete(&models.SiteCategory{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

type SiteUpsertRequest struct {
	CategoryID  uint   `json:"category_id"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Sort        int    `json:"sort"`
}

func AdminSiteList(c *gin.Context) {
	var rows []models.SiteItem
	q := config.DB.Order("sort asc, id asc")
	if categoryIDStr := c.Query("categoryId"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil && categoryID > 0 {
			q = q.Where("category_id = ?", uint(categoryID))
		}
	}
	if err := q.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminSiteCreate(c *gin.Context) {
	var req SiteUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CategoryID == 0 || req.Name == "" || req.Url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	row := models.SiteItem{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Url:         req.Url,
		Description: req.Description,
		Icon:        req.Icon,
		Sort:        req.Sort,
	}
	if err := config.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "create failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": row})
}

func AdminSiteUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	var req SiteUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	updates := map[string]interface{}{}
	if req.CategoryID != 0 {
		updates["category_id"] = req.CategoryID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Url != "" {
		updates["url"] = req.Url
	}
	updates["description"] = req.Description
	updates["icon"] = req.Icon
	updates["sort"] = req.Sort
	if err := config.DB.Model(&models.SiteItem{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

func AdminSiteDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	if err := config.DB.Where("id = ?", uint(id)).Delete(&models.SiteItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

func parseTimeFlexible(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return &t, nil
	}
	return nil, errors.New("invalid time")
}

func AdminBatchList(c *gin.Context) {
	createdAtStartStr := c.Query("createdAtStart")
	createdAtEndStr := c.Query("createdAtEnd")
	batchType := c.Query("type")

	q := config.DB.Model(&models.BatchLog{}).Order("created_at desc")

	if batchType != "" {
		q = q.Where("type = ?", batchType)
	}
	if createdAtStart, err := parseTimeFlexible(createdAtStartStr); err == nil && createdAtStart != nil {
		q = q.Where("created_at >= ?", *createdAtStart)
	}
	if createdAtEnd, err := parseTimeFlexible(createdAtEndStr); err == nil && createdAtEnd != nil {
		q = q.Where("created_at <= ?", *createdAtEnd)
	}

	var rows []models.BatchLog
	if err := q.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminBatchNewsList(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	var rows []models.NewsItem
	if err := config.DB.Where("batch_id = ?", uint(id)).Order("id asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminBatchDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	tx := config.DB.Begin()
	if err := tx.Where("batch_id = ?", uint(id)).Delete(&models.NewsItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	if err := tx.Where("batch_id = ?", uint(id)).Delete(&models.Analysis{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	if err := tx.Where("id = ?", uint(id)).Delete(&models.BatchLog{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

type NewsUpsertRequest struct {
	BatchID uint   `json:"batch_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Url     string `json:"url"`
	Source  string `json:"source"`
}

func AdminNewsList(c *gin.Context) {
	batchIDStr := c.Query("batchId")
	keyword := strings.TrimSpace(c.Query("keyword"))
	createdAtStartStr := c.Query("createdAtStart")
	createdAtEndStr := c.Query("createdAtEnd")

	q := config.DB.Model(&models.NewsItem{}).Order("created_at desc, id desc")
	if batchIDStr != "" {
		if batchID, err := strconv.ParseUint(batchIDStr, 10, 64); err == nil && batchID > 0 {
			q = q.Where("batch_id = ?", uint(batchID))
		}
	}
	if keyword != "" {
		q = q.Where("title like ?", "%"+keyword+"%")
	}
	if createdAtStart, err := parseTimeFlexible(createdAtStartStr); err == nil && createdAtStart != nil {
		q = q.Where("created_at >= ?", *createdAtStart)
	}
	if createdAtEnd, err := parseTimeFlexible(createdAtEndStr); err == nil && createdAtEnd != nil {
		q = q.Where("created_at <= ?", *createdAtEnd)
	}

	var rows []models.NewsItem
	if err := q.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminNewsCreate(c *gin.Context) {
	var req NewsUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.BatchID == 0 || strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	row := models.NewsItem{
		BatchID:   req.BatchID,
		Title:     strings.TrimSpace(req.Title),
		Content:   req.Content,
		Url:       req.Url,
		Source:    req.Source,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "create failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": row})
}

func AdminNewsUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	var req NewsUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.BatchID != 0 {
		updates["batch_id"] = req.BatchID
	}
	if title := strings.TrimSpace(req.Title); title != "" {
		updates["title"] = title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Url != "" {
		updates["url"] = req.Url
	}
	if req.Source != "" {
		updates["source"] = req.Source
	}

	if len(updates) == 1 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
		return
	}
	if err := config.DB.Model(&models.NewsItem{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

func AdminNewsDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	if err := config.DB.Where("id = ?", uint(id)).Delete(&models.NewsItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

type AnalysisUpsertRequest struct {
	BatchID uint   `json:"batch_id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func AdminAnalysisList(c *gin.Context) {
	batchIDStr := c.Query("batchId")
	analysisType := strings.TrimSpace(c.Query("type"))
	createdAtStartStr := c.Query("createdAtStart")
	createdAtEndStr := c.Query("createdAtEnd")

	q := config.DB.Model(&models.Analysis{}).Order("created_at desc, id desc")
	if batchIDStr != "" {
		if batchID, err := strconv.ParseUint(batchIDStr, 10, 64); err == nil && batchID > 0 {
			q = q.Where("batch_id = ?", uint(batchID))
		}
	}
	if analysisType != "" {
		q = q.Where("type = ?", analysisType)
	}
	if createdAtStart, err := parseTimeFlexible(createdAtStartStr); err == nil && createdAtStart != nil {
		q = q.Where("created_at >= ?", *createdAtStart)
	}
	if createdAtEnd, err := parseTimeFlexible(createdAtEndStr); err == nil && createdAtEnd != nil {
		q = q.Where("created_at <= ?", *createdAtEnd)
	}

	var rows []models.Analysis
	if err := q.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "query failed", "rows": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "rows": rows})
}

func AdminAnalysisCreate(c *gin.Context) {
	var req AnalysisUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.BatchID == 0 || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	if req.Type != string(models.Analysis3Day) && req.Type != string(models.Analysis7Day) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	row := models.Analysis{
		BatchID:   req.BatchID,
		Type:      models.AnalysisType(req.Type),
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "create failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": row})
}

func AdminAnalysisUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	var req AnalysisUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.BatchID != 0 {
		updates["batch_id"] = req.BatchID
	}
	if req.Type != "" {
		if req.Type != string(models.Analysis3Day) && req.Type != string(models.Analysis7Day) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
			return
		}
		updates["type"] = req.Type
	}
	if strings.TrimSpace(req.Content) != "" {
		updates["content"] = req.Content
	}
	if len(updates) == 1 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
		return
	}

	if err := config.DB.Model(&models.Analysis{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

func AdminAnalysisDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "bad request"})
		return
	}
	if err := config.DB.Where("id = ?", uint(id)).Delete(&models.Analysis{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}
