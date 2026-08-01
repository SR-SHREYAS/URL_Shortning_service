package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
)

type RateLimitMiddleware struct {
	server   *server.Server
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
}

func NewRateLimitMiddleware(s *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		server:   s,
		visitors: make(map[string]*rate.Limiter),
	}
}

func (r *RateLimitMiddleware) RecordRateLimitHit(endpoint string) {
	if r.server.LoggerService != nil && r.server.LoggerService.GetApplication() != nil {
		r.server.LoggerService.GetApplication().RecordCustomEvent(
			"RateLimitHit",
			map[string]interface{}{
				"endpoint": endpoint,
			},
		)
	}
}

func (r *RateLimitMiddleware) Limit(
	requestsPerSecond rate.Limit,
	burst int,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		limiter := r.getLimiter(
			"global:"+clientIP,
			requestsPerSecond,
			burst,
		)

		if !limiter.Allow() {
			r.RecordRateLimitHit(c.FullPath())

			r.server.Logger.Warn().
				Str("request_id", GetRequestID(c)).
				Str("path", c.FullPath()).
				Str("method", c.Request.Method).
				Str("ip", clientIP).
				Msg("rate limit exceeded")

			_ = c.Error(errs.TooManyRequestsError("Rate limit exceeded", false))

			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				errs.HTTPError{
					Code:     "TOO_MANY_REQUESTS",
					Message:  "Rate limit exceeded",
					Status:   http.StatusTooManyRequests,
					Override: false,
				},
			)

			return
		}

		c.Next()
	}
}

func (r *RateLimitMiddleware) Named(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		perMin := r.server.Config.RateLimit.RedirectPerMin

		if name == "link_create" {
			perMin = r.server.Config.RateLimit.LinkCreatePerMin

			if id, ok := AppUserID(c); ok {
				key = id.String()
			}
		}

		if !r.allowNamed(c.Request.Context(), name+":"+key, perMin) {
			r.RecordRateLimitHit(c.FullPath())

			_ = c.Error(errs.TooManyRequestsError("Rate limit exceeded", false))

			c.Abort()

			return
		}

		c.Next()
	}
}

func (r *RateLimitMiddleware) allowNamed(ctx context.Context, key string, limit int) bool {
	if r.server.Redis != nil {
		redisKey := "ratelimit:" + key
		count, err := r.server.Redis.Incr(ctx, redisKey).Result()
		if err == nil {
			if count == 1 {
				_ = r.server.Redis.Expire(ctx, redisKey, time.Minute).Err()
			}

			return count <= int64(limit)
		}
	}

	return r.getLimiter(key, rate.Limit(float64(limit)/60), limit).Allow()
}

func (r *RateLimitMiddleware) getLimiter(
	clientIP string,
	requestsPerSecond rate.Limit,
	burst int,
) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, exists := r.visitors[clientIP]
	if !exists {
		limiter = rate.NewLimiter(requestsPerSecond, burst)
		r.visitors[clientIP] = limiter
	}

	return limiter
}
