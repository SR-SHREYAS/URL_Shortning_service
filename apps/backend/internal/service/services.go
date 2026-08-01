package service

import (
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/lib/job"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
)

type Services struct {
	Auth      *AuthService
	Job       *job.JobService
	AppUser   *AppUserService
	Link      *LinkService
	Redirect  *RedirectService
	QRCode    *QRCodeService
	Analytics *AnalyticsService
}

func NewServices(
	s *server.Server,
	repos *repository.Repositories,
) (*Services, error) {
	authService := NewAuthService(s)
	appUserService := NewAppUserService(repos.AppUser)

	return &Services{
		Job:       s.Job,
		Auth:      authService,
		AppUser:   appUserService,
		Link:      NewLinkService(repos.Link, s.Redis, s.Config),
		Redirect:  NewRedirectService(repos.Link, s.Redis, s.Job.Client, s.Config),
		QRCode:    NewQRCodeService(repos.Link, s.Config),
		Analytics: NewAnalyticsService(repos.Analytics, s.Redis),
	}, nil
}
