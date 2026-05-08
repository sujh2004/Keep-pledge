package middleware

import (
	"net/http"
	"sync"
	"time"

	"keep-pledge/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	count     int
	resetTime time.Time
}

func RateLimit(limitPerMinute int) gin.HandlerFunc {
	if limitPerMinute <= 0 {
		limitPerMinute = 60
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.resetTime) {
			b = &bucket{count: 0, resetTime: now.Add(time.Minute)}
			buckets[ip] = b
		}
		b.count++
		allowed := b.count <= limitPerMinute
		mu.Unlock()

		if !allowed {
			response.Fail(c, http.StatusTooManyRequests, response.CodeValidationFailed, "请求过于频繁")
			c.Abort()
			return
		}
		c.Next()
	}
}
