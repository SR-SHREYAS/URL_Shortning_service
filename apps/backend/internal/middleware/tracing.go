package middleware

import (
	"github.com/gin-gonic/gin"
	nrgin "github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/integrations/nrpkgerrors"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
)

type TracingMiddleware struct {
	server *server.Server
	nrApp  *newrelic.Application
}

func NewTracingMiddleware(
	s *server.Server,
	nrApp *newrelic.Application,
) *TracingMiddleware {
	return &TracingMiddleware{
		server: s,
		nrApp:  nrApp,
	}
}

// NewRelicMiddleware returns the New Relic middleware for Gin
func (tm *TracingMiddleware) NewRelicMiddleware() gin.HandlerFunc {
	if tm.nrApp == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return nrgin.Middleware(tm.nrApp)
}

// EnhanceTracing adds custom attributes to New Relic transactions
func (tm *TracingMiddleware) EnhanceTracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get New Relic transaction from context
		txn := newrelic.FromContext(c.Request.Context())
		if txn == nil {
			c.Next()

			return
		}

		// service.name and service.environment are already set in logger and New Relic config
		txn.AddAttribute("http.real_ip", c.ClientIP())
		txn.AddAttribute("http.user_agent", c.Request.UserAgent())

		// Add request ID if available
		if requestID := GetRequestID(c); requestID != "" {
			txn.AddAttribute("request.id", requestID)
		}

		// Add user context if available
		if userID, exists := c.Get("user_id"); exists {
			if userIDStr, ok := userID.(string); ok {
				txn.AddAttribute("user.id", userIDStr)
			}
		}

		// Execute next handler
		c.Next()

		// Record error if any with enhanced stack traces
		if len(c.Errors) > 0 {
			txn.NoticeError(nrpkgerrors.Wrap(c.Errors.Last().Err))
		}

		// Add response status
		txn.AddAttribute("http.status_code", c.Writer.Status())
	}
}
