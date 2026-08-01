package router

import (
	"github.com/gin-gonic/gin"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/handler"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/middleware"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
)

func NewRouter(
	s *server.Server,
	h *handler.Handlers,
	services *service.Services,
) *gin.Engine {
	middlewares := middleware.NewMiddlewares(s, services.AppUser)

	router := gin.New()
	router.LoadHTMLGlob("templates/*")

	// global middlewares
	router.Use(
		middlewares.Global.Recover(),
		middleware.RequestID(),
		middlewares.Global.CORS(),
		middlewares.Global.Secure(),
		middlewares.Tracing.NewRelicMiddleware(),
		middlewares.Tracing.EnhanceTracing(),
		middlewares.ContextEnhancer.EnhanceContext(),
		middlewares.Global.RequestLogger(),
		middlewares.Global.ErrorHandler(),
		middlewares.RateLimit.Limit(20, 20),
	)

	// register system routes
	registerSystemRoutes(router, h)

	// Clerk or API-key authentication resolves the local AppUserID before
	// handlers execute, so both clients use the same downstream code path.
	protected := router.Group("/api/v1", middlewares.APIKey.OptionalAuth(middlewares.Auth.RequireAuth()))
	user := protected.Group("/user")
	user.POST("/api-key", h.User.RotateAPIKey)
	user.DELETE("/api-key", h.User.RevokeAPIKey)
	user.DELETE("/account", h.User.DeleteAccount)

	links := protected.Group("/links")
	links.POST("", middlewares.RateLimit.Named("link_create"), h.Link.Create)
	links.GET("", h.Link.List)
	links.GET("/export", h.Link.Export)
	links.POST("/bulk-delete", h.Link.BulkDelete)
	links.GET("/:id", h.Link.Get)
	links.PATCH("/:id", h.Link.Update)
	links.DELETE("/:id", h.Link.Delete)
	links.PATCH("/:id/status", h.Link.Toggle)
	links.POST("/:id/duplicate", h.Link.Duplicate)
	links.GET("/:id/qrcode", h.QRCode.Get)
	links.GET("/:id/analytics", h.Analytics.LinkOverview)
	links.GET("/:id/clicks", h.Analytics.Clicks)
	links.GET("/:id/countries", h.Analytics.Breakdown("countries"))
	links.GET("/:id/devices", h.Analytics.Breakdown("devices"))
	links.GET("/:id/browsers", h.Analytics.Breakdown("browsers"))
	links.GET("/:id/referrers", h.Analytics.Breakdown("referrers"))
	links.GET("/:id/recent", h.Analytics.Recent)

	dashboard := protected.Group("/dashboard")
	dashboard.GET("/overview", h.Analytics.DashboardOverview)
	dashboard.GET("/analytics", h.Analytics.DashboardAnalytics)

	// Keep this unversioned: it is the redirect hot path, not an API call.
	router.GET("/:shortCode", middlewares.RateLimit.Named("redirect"), h.Redirect.Resolve)

	router.NoRoute(func(c *gin.Context) {
		middlewares.Global.HandleError(
			c,
			errs.NewNotFoundError("Route not found", false, nil),
		)
	})

	router.NoMethod(func(c *gin.Context) {
		middlewares.Global.HandleError(
			c,
			errs.NewNotFoundError("Route not found", false, nil),
		)
	})

	return router
}
