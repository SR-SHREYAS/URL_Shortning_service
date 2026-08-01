package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/sqlerr"
)

type GlobalMiddlewares struct {
	server *server.Server
}

func NewGlobalMiddlewares(s *server.Server) *GlobalMiddlewares {
	return &GlobalMiddlewares{
		server: s,
	}
}

func (global *GlobalMiddlewares) CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: global.server.Config.Server.CORSAllowedOrigins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-API-Key",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
	})
}

func (global *GlobalMiddlewares) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		statusCode := c.Writer.Status()

		var requestErr error
		if len(c.Errors) > 0 {
			requestErr = c.Errors.Last().Err
		}

		logger := GetLogger(c)

		var e *zerolog.Event

		switch {
		case statusCode >= 500:
			e = logger.Error().Err(requestErr)

		case statusCode >= 400:
			e = logger.Warn()

		default:
			e = logger.Info()
		}

		if requestID := GetRequestID(c); requestID != "" {
			e = e.Str("request_id", requestID)
		}

		if userID := GetUserID(c); userID != "" {
			e = e.Str("user_id", userID)
		}

		e.
			Dur("latency", time.Since(start)).
			Int("status", statusCode).
			Str("method", c.Request.Method).
			Str("uri", c.Request.RequestURI).
			Str("host", c.Request.Host).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("API")
	}
}

func (global *GlobalMiddlewares) Recover() gin.HandlerFunc {
	return gin.Recovery()
}

func (global *GlobalMiddlewares) Secure() gin.HandlerFunc {
	return secure.New(secure.Config{
		SSLRedirect:          false,
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
	})
}

func (global *GlobalMiddlewares) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		global.handleError(c, c.Errors.Last().Err)
	}
}

func (global *GlobalMiddlewares) HandleError(c *gin.Context, err error) {
	global.handleError(c, err)
}

func (global *GlobalMiddlewares) handleError(c *gin.Context, err error) {
	// First try to handle database errors and convert them to appropriate HTTP errors
	originalErr := err

	// Try to handle known database errors
	// Only do this for errors that haven't already been converted to HTTPError
	var httpErr *errs.HTTPError

	if !errors.As(err, &httpErr) {
		// Here we call our sqlerr handler which will convert database errors
		// to appropriate application errors
		err = sqlerr.HandleError(err)
	}

	// Now process the possibly converted error
	var status int
	var code string
	var message string
	var fieldErrors []errs.FieldError
	var action *errs.Action

	switch {
	case errors.As(err, &httpErr):
		status = httpErr.Status
		code = httpErr.Code
		message = httpErr.Message
		fieldErrors = httpErr.Errors
		action = httpErr.Action

	default:
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(
			http.StatusText(http.StatusInternalServerError),
		)
		message = http.StatusText(http.StatusInternalServerError)
	}

	// Log the original error to help with debugging
	// Use enhanced logger from context which already includes request_id, method, path, ip, user context, and trace context
	logger := *GetLogger(c)

	logger.Error().
		Stack().
		Err(originalErr).
		Int("status", status).
		Str("error_code", code).
		Msg(message)

	if !c.Writer.Written() {
		c.AbortWithStatusJSON(status, errs.HTTPError{
			Code:     code,
			Message:  message,
			Status:   status,
			Override: httpErr != nil && httpErr.Override,
			Errors:   fieldErrors,
			Action:   action,
		})
	}
}
