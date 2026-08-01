package router

import (
	"github.com/gin-gonic/gin"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/handler"
)

func registerSystemRoutes(r *gin.Engine, h *handler.Handlers) {
	r.GET("/status", h.Health.CheckHealth)

	r.Static("/static", "static")

	r.GET("/docs", h.OpenAPI.ServeOpenAPIUI)
}
