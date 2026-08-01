package service

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/config"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"net"
	"net/url"
	"strings"
	"time"
)

type CreateLinkInput struct {
	URL, Alias, Password string
	Tags                 []string
	ExpiryHours          *int
	ClickLimit           *int
	ActivatesAt          *time.Time
}
type UpdateLinkInput struct {
	URL, Password                    *string
	Tags                             *[]string
	IsFavorite, IsArchived, IsActive *bool
	ClickLimit                       *int
	ActivatesAt, ExpiresAt           *time.Time
}
type LinkService struct {
	repo  repository.LinkRepository
	cache *redis.Client
	cfg   *config.Config
}

func NewLinkService(r repository.LinkRepository, cache *redis.Client, c *config.Config) *LinkService {
	return &LinkService{repo: r, cache: cache, cfg: c}
}
func bad(msg string) error { return errs.NewBadRequestError(msg, false, nil, nil, nil) }
func conflict(msg string) error {
	code := "SHORT_CODE_TAKEN"
	return errs.NewBadRequestError(msg, true, &code, nil, nil)
}
func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", bad("URL is required")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, e := url.Parse(raw)
	if e != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", bad("URL must be a valid http(s) URL")
	}
	return u.String(), nil
}
func (s *LinkService) Create(ctx context.Context, u uuid.UUID, in CreateLinkInput) (*model.Link, error) {
	normalized, e := normalizeURL(in.URL)
	if e != nil {
		return nil, e
	}
	parsed, _ := url.Parse(normalized)
	domain := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(s.cfg.Service.Domain), "https://"), "http://")
	if domain != "" && strings.EqualFold(parsed.Hostname(), strings.Split(domain, ":")[0]) {
		return nil, bad("cannot shorten this service domain")
	}
	code := in.Alias
	if code != "" {
		if !validCode(code) {
			return nil, bad("alias may contain only letters, numbers, hyphens, and underscores")
		}
		if _, e = s.repo.GetByShortCode(ctx, code); e == nil {
			return nil, conflict("short code is already taken")
		} else if !errors.Is(e, pgx.ErrNoRows) {
			return nil, e
		}
	}
	if code == "" {
		available := false
		for i := 0; i < 5; i++ {
			code = randomCode(s.cfg.Service.ShortCodeLength)
			if _, e = s.repo.GetByShortCode(ctx, code); errors.Is(e, pgx.ErrNoRows) {
				available = true
				break
			}
			if e != nil {
				return nil, e
			}
		}
		if !available {
			return nil, conflict("could not allocate a unique short code")
		}
	}
	hours := s.cfg.Service.DefaultExpiryHours
	if in.ExpiryHours != nil {
		hours = *in.ExpiryHours
	}
	if hours <= 0 {
		return nil, bad("expiry must be positive")
	}
	expiry := time.Now().UTC().Add(time.Duration(hours) * time.Hour)
	l := &model.Link{UserID: u, ShortCode: code, OriginalURL: normalized, Tags: in.Tags, ClickLimit: in.ClickLimit, ActivatesAt: in.ActivatesAt, ExpiresAt: &expiry}
	if in.Password != "" {
		h, e := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if e != nil {
			return nil, e
		}
		x := string(h)
		l.PasswordHash = &x
	}
	if e = s.repo.Create(ctx, l); e != nil {
		return nil, e
	}
	return l, nil
}
func (s *LinkService) Get(ctx context.Context, u, id uuid.UUID) (*model.Link, error) {
	return s.repo.GetByID(ctx, u, id)
}
func (s *LinkService) List(ctx context.Context, u uuid.UUID, f repository.ListFilters, all bool) ([]model.Link, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Sort != "" && f.Sort != "created_at" && f.Sort != "clicks" {
		return nil, 0, bad("invalid sort field")
	}
	return s.repo.List(ctx, u, f, all)
}
func (s *LinkService) Update(ctx context.Context, u, id uuid.UUID, in UpdateLinkInput) (*model.Link, error) {
	l, e := s.repo.GetByID(ctx, u, id)
	if e != nil {
		return nil, e
	}
	if in.URL != nil {
		x, e := normalizeURL(*in.URL)
		if e != nil {
			return nil, e
		}
		l.OriginalURL = x
	}
	if in.Password != nil {
		if *in.Password == "" {
			l.PasswordHash = nil
		} else {
			h, e := bcrypt.GenerateFromPassword([]byte(*in.Password), bcrypt.DefaultCost)
			if e != nil {
				return nil, e
			}
			x := string(h)
			l.PasswordHash = &x
		}
	}
	if in.Tags != nil {
		l.Tags = *in.Tags
	}
	if in.IsFavorite != nil {
		l.IsFavorite = *in.IsFavorite
	}
	if in.IsArchived != nil {
		l.IsArchived = *in.IsArchived
	}
	if in.IsActive != nil {
		l.IsActive = *in.IsActive
	}
	if in.ClickLimit != nil {
		l.ClickLimit = in.ClickLimit
	}
	if in.ActivatesAt != nil {
		l.ActivatesAt = in.ActivatesAt
	}
	if in.ExpiresAt != nil {
		l.ExpiresAt = in.ExpiresAt
	}
	e = s.repo.Update(ctx, u, l)
	if e == nil {
		s.invalidate(ctx, l.ShortCode)
	}
	return l, e
}
func (s *LinkService) Delete(ctx context.Context, u, id uuid.UUID) error {
	link, err := s.repo.GetByID(ctx, u, id)
	if err != nil {
		return err
	}
	if err = s.repo.SoftDelete(ctx, u, id); err == nil {
		s.invalidate(ctx, link.ShortCode)
	}
	return err
}
func (s *LinkService) BulkDelete(ctx context.Context, u uuid.UUID, ids []uuid.UUID) error {
	links := make([]*model.Link, 0, len(ids))
	for _, id := range ids {
		link, err := s.repo.GetByID(ctx, u, id)
		if err == nil {
			links = append(links, link)
		}
	}
	if err := s.repo.BulkSoftDelete(ctx, u, ids); err != nil {
		return err
	}
	for _, link := range links {
		s.invalidate(ctx, link.ShortCode)
	}
	return nil
}
func (s *LinkService) Toggle(ctx context.Context, u, id uuid.UUID, active bool) error {
	link, err := s.repo.GetByID(ctx, u, id)
	if err != nil {
		return err
	}
	if err = s.repo.SetActive(ctx, u, id, active); err == nil {
		s.invalidate(ctx, link.ShortCode)
	}
	return err
}
func (s *LinkService) Duplicate(ctx context.Context, u, id uuid.UUID) (*model.Link, error) {
	l, e := s.repo.GetByID(ctx, u, id)
	if e != nil {
		return nil, e
	}
	return s.Create(ctx, u, CreateLinkInput{URL: l.OriginalURL, Tags: l.Tags, ClickLimit: l.ClickLimit})
}
func randomCode(n int) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	r := make([]byte, n)
	_, _ = rand.Read(r)
	for i := range b {
		b[i] = chars[int(r[i])%len(chars)]
	}
	return string(b)
}
func validCode(v string) bool {
	if len(v) < 3 || len(v) > 64 {
		return false
	}
	for _, r := range v {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return net.ParseIP(v) == nil
}

func (s *LinkService) invalidate(ctx context.Context, code string) {
	if s.cache != nil {
		_ = s.cache.Del(ctx, "link:"+code).Err()
	}
}
