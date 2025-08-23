package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://adju-sche.vercel.app"}, // フロントのURL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}
