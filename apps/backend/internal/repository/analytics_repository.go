package repository

import (
	"context"
	"fmt"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type DashboardStats struct{ TotalLinks, ActiveLinks, ExpiredLinks, TotalClicks, UniqueVisitors, BotClicks, TodayClicks int64 }
type Breakdown struct {
	Label  string `json:"label"`
	Clicks int64  `json:"clicks"`
}

type HeatmapPoint struct {
	DayOfWeek int   `json:"dayOfWeek"`
	Hour      int   `json:"hour"`
	Clicks    int64 `json:"clicks"`
}
type AnalyticsRepository interface {
	Dashboard(context.Context, uuid.UUID, time.Time, time.Time) (DashboardStats, error)
	LinkStats(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) (DashboardStats, error)
	Series(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) ([]Breakdown, error)
	Breakdown(context.Context, uuid.UUID, uuid.UUID, string) ([]Breakdown, error)
	Recent(context.Context, uuid.UUID, uuid.UUID) ([]model.ClickEvent, error)
	Heatmap(context.Context, uuid.UUID, time.Time, time.Time) ([]HeatmapPoint, error)
}
type analyticsRepository struct{ db *pgxpool.Pool }

func NewAnalyticsRepository(db *pgxpool.Pool) AnalyticsRepository { return &analyticsRepository{db} }
func (r *analyticsRepository) Dashboard(ctx context.Context, u uuid.UUID, from, to time.Time) (s DashboardStats, err error) {
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FILTER (WHERE deleted_at IS NULL),COUNT(*) FILTER (WHERE is_active AND deleted_at IS NULL),COUNT(*) FILTER (WHERE expires_at < now() AND deleted_at IS NULL) FROM links WHERE user_id=$1`, u).Scan(&s.TotalLinks, &s.ActiveLinks, &s.ExpiredLinks)
	if err != nil {
		return
	}
	err = r.db.QueryRow(ctx, `SELECT COUNT(*),COUNT(DISTINCT ce.ip_hash),COUNT(*) FILTER (WHERE ce.is_bot),COUNT(*) FILTER (WHERE ce.clicked_at >= current_date) FROM click_events ce JOIN links l ON l.id=ce.link_id WHERE l.user_id=$1 AND ce.clicked_at >= $2 AND ce.clicked_at <= $3`, u, from, to).Scan(&s.TotalClicks, &s.UniqueVisitors, &s.BotClicks, &s.TodayClicks)
	return
}
func (r *analyticsRepository) LinkStats(ctx context.Context, u, id uuid.UUID, from, to time.Time) (s DashboardStats, err error) {
	err = r.db.QueryRow(ctx, `SELECT COUNT(*),COUNT(DISTINCT ce.ip_hash),COUNT(*) FILTER (WHERE ce.is_bot),COUNT(*) FILTER (WHERE ce.clicked_at>=current_date) FROM click_events ce JOIN links l ON l.id=ce.link_id WHERE l.user_id=$1 AND l.id=$2 AND ce.clicked_at >= $3 AND ce.clicked_at <= $4`, u, id, from, to).Scan(&s.TotalClicks, &s.UniqueVisitors, &s.BotClicks, &s.TodayClicks)
	return
}

func (r *analyticsRepository) Heatmap(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]HeatmapPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT EXTRACT(DOW FROM ce.clicked_at)::int,
		       EXTRACT(HOUR FROM ce.clicked_at)::int,
		       COUNT(*)
		FROM click_events ce
		JOIN links l ON l.id = ce.link_id
		WHERE l.user_id = $1 AND ce.clicked_at >= $2 AND ce.clicked_at <= $3
		GROUP BY 1, 2
		ORDER BY 1, 2`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := []HeatmapPoint{}
	for rows.Next() {
		var point HeatmapPoint
		if err := rows.Scan(&point.DayOfWeek, &point.Hour, &point.Clicks); err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	return points, rows.Err()
}
func (r *analyticsRepository) Series(ctx context.Context, u, id uuid.UUID, from, to time.Time) ([]Breakdown, error) {
	rows, e := r.db.Query(ctx, `SELECT to_char(clicked_at,'YYYY-MM-DD'),COUNT(*) FROM click_events ce JOIN links l ON l.id=ce.link_id WHERE l.user_id=$1 AND l.id=$2 AND clicked_at >= $3 AND clicked_at <= $4 GROUP BY 1 ORDER BY 1`, u, id, from, to)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	return scanBreakdowns(rows)
}
func (r *analyticsRepository) Breakdown(ctx context.Context, u, id uuid.UUID, col string) ([]Breakdown, error) {
	allowed := map[string]string{"countries": "country", "devices": "device", "browsers": "browser", "referrers": "referer"}
	field, ok := allowed[col]
	if !ok {
		return nil, fmt.Errorf("invalid breakdown")
	}
	rows, e := r.db.Query(ctx, `SELECT COALESCE(`+field+`,'Unknown'),COUNT(*) FROM click_events ce JOIN links l ON l.id=ce.link_id WHERE l.user_id=$1 AND l.id=$2 GROUP BY 1 ORDER BY 2 DESC LIMIT 10`, u, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	return scanBreakdowns(rows)
}
func scanBreakdowns(rows pgx.Rows) ([]Breakdown, error) {
	out := []Breakdown{}
	for rows.Next() {
		var b Breakdown
		if e := rows.Scan(&b.Label, &b.Clicks); e != nil {
			return nil, e
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (r *analyticsRepository) Recent(ctx context.Context, u, id uuid.UUID) ([]model.ClickEvent, error) {
	rows, e := r.db.Query(ctx, `SELECT ce.id,ce.link_id,ce.clicked_at,ce.ip_hash,ce.country,ce.region,ce.city,ce.timezone,ce.browser,ce.browser_version,ce.os,ce.device,ce.is_bot,ce.referer,ce.language,ce.user_agent FROM click_events ce JOIN links l ON l.id=ce.link_id WHERE l.user_id=$1 AND l.id=$2 ORDER BY ce.clicked_at DESC LIMIT 20`, u, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.ClickEvent{}
	for rows.Next() {
		var x model.ClickEvent
		e = rows.Scan(&x.ID, &x.LinkID, &x.ClickedAt, &x.IPHash, &x.Country, &x.Region, &x.City, &x.Timezone, &x.Browser, &x.BrowserVersion, &x.OS, &x.Device, &x.IsBot, &x.Referer, &x.Language, &x.UserAgent)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
