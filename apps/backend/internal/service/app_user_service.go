package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
)

type AppUserService struct {
	repo repository.AppUserRepository
}

func NewAppUserService(r repository.AppUserRepository) *AppUserService {
	return &AppUserService{
		repo: r,
	}
}

func (s *AppUserService) GetOrCreate(
	ctx context.Context,
	clerkID string,
) (uuid.UUID, error) {
	u, e := s.repo.GetOrCreate(ctx, clerkID)
	if e != nil {
		return uuid.Nil, e
	}

	return u.ID, nil
}

func (s *AppUserService) AuthenticateAPIKey(
	ctx context.Context,
	key string,
) (uuid.UUID, error) {
	u, e := s.repo.GetByAPIKey(ctx, key)
	if e != nil {
		return uuid.Nil, e
	}

	return u.ID, nil
}

func (s *AppUserService) RotateAPIKey(
	ctx context.Context,
	id uuid.UUID,
) (string, error) {
	b := make([]byte, 32)

	if _, e := rand.Read(b); e != nil {
		return "", e
	}

	key := "mrc_" + base64.RawURLEncoding.EncodeToString(b)

	return key, s.repo.SetAPIKey(ctx, id, &key)
}

func (s *AppUserService) RevokeAPIKey(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.SetAPIKey(ctx, id, nil)
}

func (s *AppUserService) DeleteAccount(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.Delete(ctx, id)
}
