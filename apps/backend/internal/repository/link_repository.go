package repository

import (
	"context"
	"fmt"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type ListFilters struct {
	Search, Tag, Sort string
	Active, Expired   *bool
	Page, Limit       int
}
type LinkRepository interface {
	Create(context.Context, *model.Link) error
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*model.Link, error)
	GetByShortCode(context.Context, string) (*model.Link, error)
	List(context.Context, uuid.UUID, ListFilters, bool) ([]model.Link, int, error)
	Update(context.Context, uuid.UUID, *model.Link) error
	SoftDelete(context.Context, uuid.UUID, uuid.UUID) error
	BulkSoftDelete(context.Context, uuid.UUID, []uuid.UUID) error
	SetActive(context.Context, uuid.UUID, uuid.UUID, bool) error
	CountClicks(context.Context, uuid.UUID) (int, error)
}
type linkRepository struct{ db *pgxpool.Pool }

func NewLinkRepository(db *pgxpool.Pool) LinkRepository { return &linkRepository{db} }

const linkColumns = `id,user_id,short_code,original_url,password_hash,tags,is_active,is_favorite,is_archived,click_limit,activates_at,expires_at,deleted_at,created_at,updated_at`

func scanLink(row pgx.Row) (*model.Link, error) {
	var l model.Link
	err := row.Scan(&l.ID, &l.UserID, &l.ShortCode, &l.OriginalURL, &l.PasswordHash, &l.Tags, &l.IsActive, &l.IsFavorite, &l.IsArchived, &l.ClickLimit, &l.ActivatesAt, &l.ExpiresAt, &l.DeletedAt, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}
func (r *linkRepository) Create(ctx context.Context, l *model.Link) error {
	return r.db.QueryRow(ctx, `INSERT INTO links (user_id,short_code,original_url,password_hash,tags,click_limit,activates_at,expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+linkColumns, l.UserID, l.ShortCode, l.OriginalURL, l.PasswordHash, l.Tags, l.ClickLimit, l.ActivatesAt, l.ExpiresAt).Scan(&l.ID, &l.UserID, &l.ShortCode, &l.OriginalURL, &l.PasswordHash, &l.Tags, &l.IsActive, &l.IsFavorite, &l.IsArchived, &l.ClickLimit, &l.ActivatesAt, &l.ExpiresAt, &l.DeletedAt, &l.CreatedAt, &l.UpdatedAt)
}
func (r *linkRepository) GetByID(ctx context.Context, user, id uuid.UUID) (*model.Link, error) {
	return scanLink(r.db.QueryRow(ctx, `SELECT `+linkColumns+` FROM links WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, user))
}
func (r *linkRepository) GetByShortCode(ctx context.Context, code string) (*model.Link, error) {
	return scanLink(r.db.QueryRow(ctx, `SELECT `+linkColumns+` FROM links WHERE short_code=$1 AND deleted_at IS NULL`, code))
}
func (r *linkRepository) List(ctx context.Context, user uuid.UUID, f ListFilters, all bool) ([]model.Link, int, error) {
	where, args := []string{"user_id=$1", "deleted_at IS NULL"}, []any{user}
	n := 2
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(original_url ILIKE $%d OR short_code ILIKE $%d)", n, n))
		args = append(args, "%"+f.Search+"%")
		n++
	}
	if f.Tag != "" {
		where = append(where, fmt.Sprintf("tags @> ARRAY[$%d]::text[]", n))
		args = append(args, f.Tag)
		n++
	}
	if f.Active != nil {
		where = append(where, fmt.Sprintf("is_active=$%d", n))
		args = append(args, *f.Active)
		n++
	}
	if f.Expired != nil && *f.Expired {
		where = append(where, "expires_at < now()")
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM links WHERE "+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	order := "created_at DESC"
	if f.Sort == "clicks" {
		order = "(SELECT COUNT(*) FROM click_events ce WHERE ce.link_id=links.id) DESC"
	}
	q := "SELECT " + linkColumns + " FROM links WHERE " + clause + " ORDER BY " + order
	if !all {
		if f.Limit <= 0 {
			f.Limit = 20
		}
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, (f.Page-1)*f.Limit)
	}
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.Link{}
	for rows.Next() {
		l, e := scanLink(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, *l)
	}
	return out, total, rows.Err()
}
func (r *linkRepository) Update(ctx context.Context, user uuid.UUID, l *model.Link) error {
	cmd, err := r.db.Exec(ctx, `UPDATE links SET original_url=$3,password_hash=$4,tags=$5,is_active=$6,is_favorite=$7,is_archived=$8,click_limit=$9,activates_at=$10,expires_at=$11,updated_at=now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, l.ID, user, l.OriginalURL, l.PasswordHash, l.Tags, l.IsActive, l.IsFavorite, l.IsArchived, l.ClickLimit, l.ActivatesAt, l.ExpiresAt)
	if err == nil && cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *linkRepository) SoftDelete(ctx context.Context, user, id uuid.UUID) error {
	cmd, err := r.db.Exec(ctx, `UPDATE links SET deleted_at=now(),updated_at=now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, user)
	if err == nil && cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *linkRepository) BulkSoftDelete(ctx context.Context, user uuid.UUID, ids []uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE links SET deleted_at=now(),updated_at=now() WHERE user_id=$1 AND id=ANY($2) AND deleted_at IS NULL`, user, ids)
	return err
}
func (r *linkRepository) SetActive(ctx context.Context, user, id uuid.UUID, a bool) error {
	cmd, err := r.db.Exec(ctx, `UPDATE links SET is_active=$3,updated_at=now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, user, a)
	if err == nil && cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *linkRepository) CountClicks(ctx context.Context, id uuid.UUID) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM click_events WHERE link_id=$1`, id).Scan(&n)
	return n, err
}
