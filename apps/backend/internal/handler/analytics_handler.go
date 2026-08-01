package handler

import (
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AnalyticsHandler struct {
	Handler
	analytics *service.AnalyticsService
}

func NewAnalyticsHandler(s *server.Server, analytics *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{Handler: NewHandler(s), analytics: analytics}
}
func (h *AnalyticsHandler) DashboardOverview(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	v, e := h.analytics.Dashboard(c.Request.Context(), u, c.DefaultQuery("range", "30d"))
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *AnalyticsHandler) DashboardAnalytics(c *gin.Context) {
	userID, ok := appUser(c)
	if !ok {
		return
	}

	rangeName := c.DefaultQuery("range", "30d")
	stats, err := h.analytics.Dashboard(c.Request.Context(), userID, rangeName)
	if err != nil {
		respondErr(c, err)
		return
	}
	heatmap, err := h.analytics.Heatmap(c.Request.Context(), userID, rangeName)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"overview": stats, "heatmap": heatmap})
}
func (h *AnalyticsHandler) LinkOverview(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	v, e := h.analytics.LinkOverview(c.Request.Context(), u, id, c.DefaultQuery("range", "30d"))
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *AnalyticsHandler) Clicks(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	v, e := h.analytics.Series(c.Request.Context(), u, id, c.DefaultQuery("range", "30d"))
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": v})
}
func (h *AnalyticsHandler) Breakdown(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := appUser(c)
		if !ok {
			return
		}
		id, ok := requestID(c)
		if !ok {
			return
		}
		v, e := h.analytics.Breakdown(c.Request.Context(), u, id, kind)
		if e != nil {
			respondErr(c, e)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": v})
	}
}
func (h *AnalyticsHandler) Recent(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	v, e := h.analytics.Recent(c.Request.Context(), u, id)
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": v})
}
