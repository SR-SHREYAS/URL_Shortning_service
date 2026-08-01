package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
)

type UserHandler struct {
	Handler
	users *service.AppUserService
}

func NewUserHandler(s *server.Server, users *service.AppUserService) *UserHandler {
	return &UserHandler{
		Handler: NewHandler(s),
		users:   users,
	}
}

func (h *UserHandler) RotateAPIKey(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}

	key, e := h.users.RotateAPIKey(c.Request.Context(), u)
	if e != nil {
		respondErr(c, e)

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"apiKey": key,
	})
}

func (h *UserHandler) RevokeAPIKey(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}

	if e := h.users.RevokeAPIKey(c.Request.Context(), u); e != nil {
		respondErr(c, e)

		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}

	if e := h.users.DeleteAccount(c.Request.Context(), u); e != nil {
		respondErr(c, e)

		return
	}

	c.Status(http.StatusNoContent)
}
