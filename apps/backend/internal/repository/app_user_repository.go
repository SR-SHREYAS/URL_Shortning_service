package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
)

type AppUserRepository interface {
	GetOrCreate(context.Context, string) (*model.AppUser, error)
	GetByAPIKey(context.Context, string) (*model.AppUser, error)
	SetAPIKey(context.Context, uuid.UUID, *string) error
	Delete(context.Context, uuid.UUID) error
}

type appUserRepository struct {
	db *pgxpool.Pool
}

func NewAppUserRepository(db *pgxpool.Pool) AppUserRepository {
	return &appUserRepository{
		db: db,
	}
}

func scanAppUser(row pgx.Row) (*model.AppUser, error) {
	var u model.AppUser

	if err := row.Scan(
		&u.ID,
		&u.ClerkUserID,
		&u.APIKey,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *appUserRepository) GetOrCreate(
	ctx context.Context,
	clerkID string,
) (*model.AppUser, error) {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO app_users (clerk_user_id) VALUES ($1) ON CONFLICT (clerk_user_id) DO NOTHING`,
		clerkID,
	)
	if err != nil {
		return nil, err
	}

	return scanAppUser(
		r.db.QueryRow(
			ctx,
			`SELECT id,clerk_user_id,api_key,created_at,updated_at FROM app_users WHERE clerk_user_id=$1`,
			clerkID,
		),
	)
}

func (r *appUserRepository) GetByAPIKey(
	ctx context.Context,
	key string,
) (*model.AppUser, error) {
	return scanAppUser(
		r.db.QueryRow(
			ctx,
			`SELECT id,clerk_user_id,api_key,created_at,updated_at FROM app_users WHERE api_key=$1`,
			key,
		),
	)
}

func (r *appUserRepository) SetAPIKey(
	ctx context.Context,
	id uuid.UUID,
	key *string,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE app_users SET api_key=$2,updated_at=now() WHERE id=$1`,
		id,
		key,
	)

	return err
}

func (r *appUserRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM app_users WHERE id=$1`,
		id,
	)

	return err
}
