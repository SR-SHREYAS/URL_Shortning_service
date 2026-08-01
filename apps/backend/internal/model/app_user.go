package model

import (
	"time"

	"github.com/google/uuid"
)

type AppUser struct {
	ID          uuid.UUID `json:"id"`
	ClerkUserID string    `json:"clerkUserId"`
	APIKey      *string   `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
