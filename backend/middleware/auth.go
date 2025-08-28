package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"taraftar-analizi/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	clients map[string]*rate.Limiter
	mutex   sync.RWMutex
}

var rateLimiter = &RateLimiter{
	clients: make(map[string]*rate.Limiter),
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mutex.RLock()
	limiter, exists := rl.clients[ip]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		limiter = rate.NewLimiter(rate.Every(time.Second), 50) // 50 requests per second
		rl.clients[ip] = limiter
		rl.mutex.Unlock()
	}

	return limiter
}

func RateLimitMiddleware() gin.HandlerFunc {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			rateLimiter.mutex.Lock()
			rateLimiter.clients = make(map[string]*rate.Limiter)
			rateLimiter.mutex.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rateLimiter.getLimiter(ip)

		if !limiter.Allow() {
			log.Printf("Rate limit exceeded for IP: %s", ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/webhook") || 
		   strings.Contains(c.Request.URL.Path, "/n8n") {
			
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.Query("api_key")
			}

			if apiKey == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Unauthorized",
					"message": "API key is required for webhook endpoints",
				})
				c.Abort()
				return
			}

			if apiKey != config.AppConfig.APISecret {
				log.Printf("Invalid API key attempt from IP: %s", c.ClientIP())
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Unauthorized", 
					"message": "Invalid API key",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
			c.Abort()
			return
		}

		if tokenString != config.AppConfig.APISecret {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader != "" {
			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
			if tokenString == config.AppConfig.APISecret {
				c.Set("authenticated", true)
			}
		}
		
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			log.Printf("Error in %s %s: %v", c.Request.Method, c.Request.URL.Path, err.Err)
			
			if c.Writer.Written() {
				return
			}

			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation Error",
					"message": err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": "Something went wrong",
				})
			}
		}
	}
}