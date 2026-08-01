package handler

import (
	"net/http"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
	"github.com/gin-gonic/gin"
)

type QRCodeHandler struct {
	Handler
	qrcodes *service.QRCodeService
}

func NewQRCodeHandler(s *server.Server, qrcodes *service.QRCodeService) *QRCodeHandler {
	return &QRCodeHandler{Handler: NewHandler(s), qrcodes: qrcodes}
}

func (h *QRCodeHandler) Get(c *gin.Context) {
	userID, ok := appUser(c)
	if !ok {
		return
	}
	linkID, ok := requestID(c)
	if !ok {
		return
	}
	png, err := h.qrcodes.PNG(c.Request.Context(), userID, linkID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}
