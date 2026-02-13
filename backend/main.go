package main

import (
	"bre_new_backend/config"
	"bre_new_backend/controllers"
	"bre_new_backend/services"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	// 1. Init Config & DB
	config.InitConfig()
	config.InitDB()

	// 2. Setup Cron
	c := cron.New()
	// Run at 8:00, 12:00, 18:00
	// Cron spec: Second Minute Hour Dom Month Dow
	// robfig/cron standard parser is Minute Hour Dom Month Dow (5 fields)
	// With Seconds field (6 fields) if using cron.WithSeconds()
	// Default standard parser is 5 fields.

	// Morning 8:00
	_, err := c.AddFunc("0 8 * * *", func() {
		services.RunUpdateTask()
	})
	if err != nil {
		fmt.Println("Error scheduling morning task:", err)
	}

	// Noon 12:00
	_, err = c.AddFunc("0 12 * * *", func() {
		services.RunUpdateTask()
	})
	if err != nil {
		fmt.Println("Error scheduling noon task:", err)
	}

	// Evening 18:00
	_, err = c.AddFunc("0 18 * * *", func() {
		services.RunUpdateTask()
	})
	if err != nil {
		fmt.Println("Error scheduling evening task:", err)
	}

	c.Start()

	// Optional: Run immediately on startup if DB is empty for demo purposes
	// go services.RunUpdateTask()

	// 3. Setup Router
	r := gin.Default()

	// Traffic Counter Middleware
	r.Use(func(c *gin.Context) {
		// 排除 OPTIONS 请求和静态资源（如果有）
		if c.Request.Method != "OPTIONS" {
			services.RecordVisit(c.ClientIP())
		}
		c.Next()
	})

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		requestHeaders := c.GetHeader("Access-Control-Request-Headers")
		if requestHeaders != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", requestHeaders)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := r.Group("/api")
	{
		api.GET("/news/latest", controllers.GetLatestNews)
		api.GET("/analysis/latest", controllers.GetLatestAnalysis)
		// Admin trigger for testing
		api.POST("/admin/trigger-update", func(c *gin.Context) {
			go services.RunUpdateTask()
			c.JSON(200, gin.H{"message": "Update task triggered"})
		})
	}

	// 4. Run
	r.Run(":" + config.AppConfig.System.Port)
}
