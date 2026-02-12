package controllers

import (
	"bre_new_backend/config"
	"bre_new_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLatestNews(c *gin.Context) {
	// Find the latest batch
	var lastBatch models.BatchLog
	result := config.DB.Order("created_at desc").First(&lastBatch)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "No news generated yet",
			"rows": []interface{}{},
		})
		return
	}

	var news []models.NewsItem
	config.DB.Where("batch_id = ?", lastBatch.ID).Find(&news)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"rows": news,
		"batch": lastBatch,
	})
}

func GetLatestAnalysis(c *gin.Context) {
	days := c.Query("days") // 3 or 7
	
	var analysisType models.AnalysisType
	if days == "7" {
		analysisType = models.Analysis7Day
	} else {
		analysisType = models.Analysis3Day // Default to 3
	}

	var analysis models.Analysis
	result := config.DB.Where("type = ?", analysisType).Order("created_at desc").First(&analysis)
	
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "No analysis found",
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": analysis,
	})
}
