package handler

import (
	"net/http"
	"os"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/gin-gonic/gin"
)

type OpenAPIHandler struct {
	Handler
}

func NewOpenAPIHandler(s *server.Server) *OpenAPIHandler {
	return &OpenAPIHandler{
		Handler: NewHandler(s),
	}
}

func (h *OpenAPIHandler) ServeOpenAPIUI(c *gin.Context) {
	templateBytes, err := os.ReadFile("static/openapi.html")

	c.Writer.Header().Set("Cache-Control", "no-cache")

	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", templateBytes)
}
