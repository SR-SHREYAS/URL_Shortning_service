package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
)

type APIKeyMiddleware struct {
	users *service.AppUserService
}

func NewAPIKeyMiddleware(users *service.AppUserService) *APIKeyMiddleware {
	return &APIKeyMiddleware{users}
}

// OptionalAuth accepts an API key when supplied; Clerk remains the fallback.
func (m *APIKeyMiddleware) OptionalAuth(clerk gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")

		if key == "" {
			clerk(c)

			return
		}

		id, e := m.users.AuthenticateAPIKey(c.Request.Context(), key)
		if e != nil {
			_ = c.Error(errs.NewUnauthorizedError("Invalid API key", false))

			c.Abort()

			return
		}

		c.Set(AppUserIDKey, id)

		c.Next()
	}
}
