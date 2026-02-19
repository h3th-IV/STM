package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// AuthRateLimit returns a rate of 100 requests per 5 minutes.
func AuthRateLimit() limiter.Rate {
	return limiter.Rate{
		Period: 5 * time.Minute,
		Limit:  100,
	}
}

// RateLimiter creates a rate limit middleware (e.g., 100 req/5min per IP).
// Each call creates a new store - for production, reuse store/instance.
func RateLimiter(rate limiter.Rate) gin.HandlerFunc {
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		key := c.ClientIP()
		context, err := instance.Get(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			c.Abort()
			return
		}

		c.Next()
	}
}
