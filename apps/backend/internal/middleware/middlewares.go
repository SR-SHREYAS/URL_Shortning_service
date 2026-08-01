package middleware

import (
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
)

type Middlewares struct {
	Global          *GlobalMiddlewares
	Auth            *AuthMiddleware
	ContextEnhancer *ContextEnhancer
	Tracing         *TracingMiddleware
	RateLimit       *RateLimitMiddleware
	APIKey          *APIKeyMiddleware
}

func NewMiddlewares(s *server.Server, users *service.AppUserService) *Middlewares {
	// Get New Relic application instance from server
	var nrApp *newrelic.Application

	if s.LoggerService != nil {
		nrApp = s.LoggerService.GetApplication()
	}

	return &Middlewares{
		Global:          NewGlobalMiddlewares(s),
		Auth:            NewAuthMiddleware(s, users),
		APIKey:          NewAPIKeyMiddleware(users),
		ContextEnhancer: NewContextEnhancer(s),
		Tracing:         NewTracingMiddleware(s, nrApp),
		RateLimit:       NewRateLimitMiddleware(s),
	}
}
