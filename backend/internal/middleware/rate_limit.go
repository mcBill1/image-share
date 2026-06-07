package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen int64
}

var (
	visitors      = make(map[string]*visitor)
	visitorsMutex sync.Mutex
)

func getVisitor(ip string, r rate.Limit, b int) *rate.Limiter {
	visitorsMutex.Lock()
	defer visitorsMutex.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(r, b)
		visitors[ip] = &visitor{limiter: limiter}
		return limiter
	}
	return v.limiter
}

func RateLimitMiddleware(r rate.Limit, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip, r, burst)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func GuestRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(rate.Limit(30), 30)
}

func AdminRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(rate.Limit(100), 100)
}