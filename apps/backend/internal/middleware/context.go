package middleware

import (
	"context"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/logger"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
)

const (
	UserIDKey    = "user_id"
	AppUserIDKey = "app_user_id"
	UserRoleKey  = "user_role"
	LoggerKey    = "logger"
)

type ContextEnhancer struct {
	server *server.Server
}

func AppUserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(AppUserIDKey)
	if !ok {
		return uuid.Nil, false
	}

	id, ok := v.(uuid.UUID)

	return id, ok && id != uuid.Nil
}

func NewContextEnhancer(s *server.Server) *ContextEnhancer {
	return &ContextEnhancer{
		server: s,
	}
}

func (ce *ContextEnhancer) EnhanceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract request ID
		requestID := GetRequestID(c)

		// Create enhanced logger with request context
		contextLogger := ce.server.Logger.
			With().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.FullPath()).
			Str("ip", c.ClientIP()).
			Logger()

		// Add trace context if available
		if txn := newrelic.FromContext(c.Request.Context()); txn != nil {
			contextLogger = logger.WithTraceContext(contextLogger, txn)
		}

		// Extract user information from JWT token or session
		if userID := ce.extractUserID(c); userID != "" {
			contextLogger = contextLogger.With().
				Str("user_id", userID).
				Logger()
		}

		if userRole := ce.extractUserRole(c); userRole != "" {
			contextLogger = contextLogger.With().
				Str("user_role", userRole).
				Logger()
		}

		// Store the enhanced logger in context
		c.Set(LoggerKey, &contextLogger)

		// Create a new context with the logger
		ctx := context.WithValue(c.Request.Context(), LoggerKey, &contextLogger)

		c.Request = c.Request.WithContext(ctx)
	}
}

func (ce *ContextEnhancer) extractUserID(c *gin.Context) string {
	// Check if user_id was already set by auth middleware (Clerk)
	if userID, exists := c.Get("user_id"); exists {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			return userIDStr
		}
	}

	return ""
}

func (ce *ContextEnhancer) extractUserRole(c *gin.Context) string {
	// Check if user_role was set by auth middleware (Clerk)
	if userRole, exists := c.Get("user_role"); exists {
		if userRoleStr, ok := userRole.(string); ok && userRoleStr != "" {
			return userRoleStr
		}
	}

	return ""
}

func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(UserIDKey); exists {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr
		}
	}

	return ""
}

func GetLogger(c *gin.Context) *zerolog.Logger {
	if loggerValue, exists := c.Get(LoggerKey); exists {
		if logger, ok := loggerValue.(*zerolog.Logger); ok {
			return logger
		}
	}

	// Fallback to a basic logger if not found
	logger := zerolog.Nop()

	return &logger
}
