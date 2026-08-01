package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/config"
	qr "github.com/SR-SHREYAS/URL_Shortning_service/internal/lib/qrcode"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
)

type QRCodeService struct {
	links repository.LinkRepository
	cfg   *config.Config
}

func NewQRCodeService(
	l repository.LinkRepository,
	c *config.Config,
) *QRCodeService {
	return &QRCodeService{
		links: l,
		cfg:   c,
	}
}

func (s *QRCodeService) PNG(
	ctx context.Context,
	u uuid.UUID,
	id uuid.UUID,
) ([]byte, error) {
	l, e := s.links.GetByID(ctx, u, id)
	if e != nil {
		return nil, e
	}

	return qr.PNG("https://" + s.cfg.Service.Domain + "/" + l.ShortCode)
}
