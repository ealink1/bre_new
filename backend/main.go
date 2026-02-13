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
	_ = services.EnsureAdminUser(config.DB)

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
		api.GET("/sites/categories", controllers.GetSiteCategories)
	}

	admin := api.Group("/admin")
	{
		// admin.POST("/setup", controllers.AdminSetup) // 暂时不开放此功能
		admin.POST("/login", controllers.AdminLogin)
	}

	adminAuthed := api.Group("/admin")
	adminAuthed.Use(controllers.AdminAuthMiddleware())
	{
		adminAuthed.POST("/logout", controllers.AdminLogout)
		adminAuthed.POST("/trigger-update", func(c *gin.Context) {
			go services.RunUpdateTask()
			c.JSON(200, gin.H{"code": 200, "msg": "success"})
		})

		adminAuthed.GET("/users", controllers.AdminUserList)
		adminAuthed.POST("/users", controllers.AdminUserCreate)
		adminAuthed.PATCH("/users/:id/password", controllers.AdminUserSetPassword)
		adminAuthed.DELETE("/users/:id", controllers.AdminUserDelete)

		adminAuthed.GET("/site-categories", controllers.AdminSiteCategoryList)
		adminAuthed.POST("/site-categories", controllers.AdminSiteCategoryCreate)
		adminAuthed.PATCH("/site-categories/:id", controllers.AdminSiteCategoryUpdate)
		adminAuthed.DELETE("/site-categories/:id", controllers.AdminSiteCategoryDelete)

		adminAuthed.GET("/sites", controllers.AdminSiteList)
		adminAuthed.POST("/sites", controllers.AdminSiteCreate)
		adminAuthed.PATCH("/sites/:id", controllers.AdminSiteUpdate)
		adminAuthed.DELETE("/sites/:id", controllers.AdminSiteDelete)

		adminAuthed.GET("/batches", controllers.AdminBatchList)
		adminAuthed.GET("/batches/:id/news", controllers.AdminBatchNewsList)
		adminAuthed.DELETE("/batches/:id", controllers.AdminBatchDelete)

		adminAuthed.GET("/news", controllers.AdminNewsList)
		adminAuthed.POST("/news", controllers.AdminNewsCreate)
		adminAuthed.PATCH("/news/:id", controllers.AdminNewsUpdate)
		adminAuthed.DELETE("/news/:id", controllers.AdminNewsDelete)

		adminAuthed.GET("/analysis", controllers.AdminAnalysisList)
		adminAuthed.POST("/analysis", controllers.AdminAnalysisCreate)
		adminAuthed.PATCH("/analysis/:id", controllers.AdminAnalysisUpdate)
		adminAuthed.DELETE("/analysis/:id", controllers.AdminAnalysisDelete)
	}

	// 4. Run
	r.Run(":" + config.AppConfig.System.Port)
}
