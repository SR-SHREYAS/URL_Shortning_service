package repository

import (
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
)

type Repositories struct {
	AppUser   AppUserRepository
	Link      LinkRepository
	Click     ClickRepository
	Analytics AnalyticsRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		AppUser:   NewAppUserRepository(s.DB.Pool),
		Link:      NewLinkRepository(s.DB.Pool),
		Click:     NewClickRepository(s.DB.Pool),
		Analytics: NewAnalyticsRepository(s.DB.Pool),
	}
}
