package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
)

type AnalyticsService struct {
	repo  repository.AnalyticsRepository
	cache *redis.Client
}

func NewAnalyticsService(
	r repository.AnalyticsRepository,
	c *redis.Client,
) *AnalyticsService {
	return &AnalyticsService{
		repo:  r,
		cache: c,
	}
}

func period(rangeName string) (time.Time, time.Time) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)

	if rangeName == "today" {
		from = time.Date(
			to.Year(),
			to.Month(),
			to.Day(),
			0,
			0,
			0,
			0,
			time.UTC,
		)
	}

	if rangeName == "7d" {
		from = to.AddDate(0, 0, -7)
	}

	return from, to
}

func (s *AnalyticsService) Dashboard(
	ctx context.Context,
	u uuid.UUID,
	rangeName string,
) (repository.DashboardStats, error) {
	from, to := period(rangeName)

	key := "dash:" + u.String() + ":" + rangeName

	if s.cache != nil {
		if raw, e := s.cache.Get(ctx, key).Bytes(); e == nil {
			var v repository.DashboardStats

			if json.Unmarshal(raw, &v) == nil {
				return v, nil
			}
		}
	}

	v, e := s.repo.Dashboard(ctx, u, from, to)

	if e == nil && s.cache != nil {
		if raw, x := json.Marshal(v); x == nil {
			_ = s.cache.Set(ctx, key, raw, time.Minute).Err()
		}
	}

	return v, e
}

func (s *AnalyticsService) Heatmap(
	ctx context.Context,
	userID uuid.UUID,
	rangeName string,
) ([]repository.HeatmapPoint, error) {
	from, to := period(rangeName)
	return s.repo.Heatmap(ctx, userID, from, to)
}

func (s *AnalyticsService) LinkOverview(
	ctx context.Context,
	u uuid.UUID,
	id uuid.UUID,
	rangeName string,
) (repository.DashboardStats, error) {
	f, t := period(rangeName)

	return s.repo.LinkStats(ctx, u, id, f, t)
}

func (s *AnalyticsService) Series(
	ctx context.Context,
	u uuid.UUID,
	id uuid.UUID,
	rangeName string,
) ([]repository.Breakdown, error) {
	f, t := period(rangeName)

	return s.repo.Series(ctx, u, id, f, t)
}

func (s *AnalyticsService) Breakdown(
	ctx context.Context,
	u uuid.UUID,
	id uuid.UUID,
	kind string,
) ([]repository.Breakdown, error) {
	return s.repo.Breakdown(ctx, u, id, kind)
}

func (s *AnalyticsService) Recent(
	ctx context.Context,
	u uuid.UUID,
	id uuid.UUID,
) ([]model.ClickEvent, error) {
	return s.repo.Recent(ctx, u, id)
}
