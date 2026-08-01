package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
)

type ClickRepository interface {
	Create(context.Context, *model.ClickEvent) error
}

type clickRepository struct {
	db *pgxpool.Pool
}

func NewClickRepository(db *pgxpool.Pool) ClickRepository {
	return &clickRepository{
		db: db,
	}
}

func (r *clickRepository) Create(
	ctx context.Context,
	e *model.ClickEvent,
) error {
	return r.db.QueryRow(
		ctx,
		`INSERT INTO click_events (link_id,clicked_at,ip_hash,country,region,city,timezone,browser,browser_version,os,device,is_bot,referer,language,user_agent) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
		e.LinkID,
		e.ClickedAt,
		e.IPHash,
		e.Country,
		e.Region,
		e.City,
		e.Timezone,
		e.Browser,
		e.BrowserVersion,
		e.OS,
		e.Device,
		e.IsBot,
		e.Referer,
		e.Language,
		e.UserAgent,
	).Scan(&e.ID)
}
