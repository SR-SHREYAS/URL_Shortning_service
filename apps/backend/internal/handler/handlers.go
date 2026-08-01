package handler

import (
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
)

type Handlers struct {
	Health    *HealthHandler
	OpenAPI   *OpenAPIHandler
	User      *UserHandler
	Link      *LinkHandler
	Redirect  *RedirectHandler
	QRCode    *QRCodeHandler
	Analytics *AnalyticsHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:    NewHealthHandler(s),
		OpenAPI:   NewOpenAPIHandler(s),
		User:      NewUserHandler(s, services.AppUser),
		Link:      NewLinkHandler(s, services.Link),
		Redirect:  NewRedirectHandler(s, services.Redirect),
		QRCode:    NewQRCodeHandler(s, services.QRCode),
		Analytics: NewAnalyticsHandler(s, services.Analytics),
	}
}
