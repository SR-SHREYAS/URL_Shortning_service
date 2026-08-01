package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/lib/job"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
	"github.com/gin-gonic/gin"
)

type RedirectHandler struct {
	Handler
	redirects *service.RedirectService
}

func NewRedirectHandler(s *server.Server, redirects *service.RedirectService) *RedirectHandler {
	return &RedirectHandler{Handler: NewHandler(s), redirects: redirects}
}
func (h *RedirectHandler) Resolve(c *gin.Context) {
	link, err := h.redirects.Resolve(c.Request.Context(), c.Param("shortCode"), c.Query("pw"))
	if err != nil {
		var httpErr *errs.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusUnauthorized && link != nil {
			c.HTML(http.StatusUnauthorized, "password_prompt.html", gin.H{"shortCode": c.Param("shortCode")})
			return
		}
		respondErr(c, err)
		return
	}
	h.redirects.Enqueue(c.Request.Context(), link, job.ClickPayload{IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), Referer: c.Request.Referer(), AcceptLanguage: c.GetHeader("Accept-Language"), ClickedAt: time.Now().UTC()})
	c.Redirect(http.StatusFound, link.OriginalURL)
}
