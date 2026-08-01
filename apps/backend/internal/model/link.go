package model

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ShortCode string    `json:"shortCode"`

	OriginalURL  string  `json:"originalUrl"`
	PasswordHash *string `json:"-"`

	Tags []string `json:"tags"`

	IsActive   bool `json:"isActive"`
	IsFavorite bool `json:"isFavorite"`
	IsArchived bool `json:"isArchived"`

	ClickLimit  *int       `json:"clickLimit,omitempty"`
	ActivatesAt *time.Time `json:"activatesAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	DeletedAt   *time.Time `json:"-"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
