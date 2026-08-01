package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/config"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/lib/job"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
)

type RedirectService struct {
	repo  repository.LinkRepository
	redis *redis.Client
	jobs  *asynq.Client
	cfg   *config.Config
}

// cachedLink contains precisely the redirect hot-path fields. PasswordHash is
// intentionally included, unlike model.Link's JSON representation, so cache
// hits enforce password protection exactly like database reads.
type cachedLink struct {
	ID           uuid.UUID  `json:"id"`
	OriginalURL  string     `json:"original_url"`
	IsActive     bool       `json:"is_active"`
	PasswordHash *string    `json:"password_hash"`
	ClickLimit   *int       `json:"click_limit"`
	ActivatesAt  *time.Time `json:"activates_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

func NewRedirectService(
	r repository.LinkRepository,
	cache *redis.Client,
	jobs *asynq.Client,
	c *config.Config,
) *RedirectService {
	return &RedirectService{
		repo:  r,
		redis: cache,
		jobs:  jobs,
		cfg:   c,
	}
}

func (s *RedirectService) Resolve(
	ctx context.Context,
	code,
	pw string,
) (*model.Link, error) {
	var l model.Link

	key := "link:" + code

	if s.redis != nil {
		if raw, e := s.redis.Get(ctx, key).Result(); e == nil {
			var cached cachedLink
			if json.Unmarshal([]byte(raw), &cached) == nil {
				l.ID = cached.ID
				l.OriginalURL = cached.OriginalURL
				l.IsActive = cached.IsActive
				l.PasswordHash = cached.PasswordHash
				l.ClickLimit = cached.ClickLimit
				l.ActivatesAt = cached.ActivatesAt
				l.ExpiresAt = cached.ExpiresAt
			}
		}
	}

	if l.ID == uuid.Nil {
		x, e := s.repo.GetByShortCode(ctx, code)
		if e != nil {
			return nil, e
		}

		l = *x

		if s.redis != nil {
			cached := cachedLink{
				ID:           l.ID,
				OriginalURL:  l.OriginalURL,
				IsActive:     l.IsActive,
				PasswordHash: l.PasswordHash,
				ClickLimit:   l.ClickLimit,
				ActivatesAt:  l.ActivatesAt,
				ExpiresAt:    l.ExpiresAt,
			}
			if raw, e := json.Marshal(cached); e == nil {
				_ = s.redis.Set(
					ctx,
					key,
					raw,
					time.Duration(s.cfg.Service.LinkCacheTTLSeconds)*time.Second,
				).Err()
			}
		}
	}

	now := time.Now()

	if !l.IsActive ||
		(l.ExpiresAt != nil && l.ExpiresAt.Before(now)) ||
		(l.ActivatesAt != nil && l.ActivatesAt.After(now)) {
		return nil, errs.NewNotFoundError("Link not found", false, nil)
	}

	if l.ClickLimit != nil {
		n, e := s.repo.CountClicks(ctx, l.ID)
		if e != nil {
			return nil, e
		}

		if n >= *l.ClickLimit {
			return nil, errs.NewNotFoundError("Link not found", false, nil)
		}
	}

	if l.PasswordHash != nil {
		if pw == "" ||
			bcrypt.CompareHashAndPassword(
				[]byte(*l.PasswordHash),
				[]byte(pw),
			) != nil {
			return &l, errs.NewUnauthorizedError("Password required", false)
		}
	}

	return &l, nil
}

func (s *RedirectService) Enqueue(
	ctx context.Context,
	l *model.Link,
	p job.ClickPayload,
) {
	if s.jobs == nil {
		return
	}

	p.LinkID = l.ID

	t, e := job.NewClickEnrichTask(p)
	if e == nil {
		_, _ = s.jobs.EnqueueContext(ctx, t)
	}
}
