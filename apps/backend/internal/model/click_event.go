package model

import (
	"time"

	"github.com/google/uuid"
)

type ClickEvent struct {
	ID        int64     `json:"id"`
	LinkID    uuid.UUID `json:"linkId"`
	ClickedAt time.Time `json:"clickedAt"`

	IPHash   string  `json:"-"`
	Country  *string `json:"country,omitempty"`
	Region   *string `json:"region,omitempty"`
	City     *string `json:"city,omitempty"`
	Timezone *string `json:"timezone,omitempty"`

	Browser        *string `json:"browser,omitempty"`
	BrowserVersion *string `json:"browserVersion,omitempty"`
	OS             *string `json:"os,omitempty"`
	Device         *string `json:"device,omitempty"`

	IsBot     bool    `json:"isBot"`
	Referer   *string `json:"referer,omitempty"`
	Language  *string `json:"language,omitempty"`
	UserAgent *string `json:"userAgent,omitempty"`
}
