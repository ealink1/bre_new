package controllers

import (
	"bre_new_backend/config"
	"bre_new_backend/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := ""
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "unauthorized",
			})
			c.Abort()
			return
		}

		user, session, err := services.GetAdminSessionUser(config.DB, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "unauthorized",
			})
			c.Abort()
			return
		}

		c.Set("adminUser", user)
		c.Set("adminSession", session)
		c.Next()
	}
}

