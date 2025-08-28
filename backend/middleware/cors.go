package middleware

import (
	"taraftar-analizi/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins: []string{
			config.AppConfig.FrontendURL,
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"https://localhost:3000",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language",
			"Cache-Control", "Connection", "Host", "Pragma", "Referer", "User-Agent",
			"X-API-Key", "Accept-Charset",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Type", "Cache-Control", "Last-Modified",
			"Content-Encoding",
		},
		AllowCredentials: true,
		MaxAge:          86400, // 24 hours
	}

	if config.AppConfig.GinMode == "debug" {
		corsConfig.AllowAllOrigins = true
	}

	return cors.New(corsConfig)
}