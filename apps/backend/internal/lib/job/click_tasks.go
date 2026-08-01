package job

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	ua "github.com/SR-SHREYAS/URL_Shortning_service/internal/lib/useragent"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
)

const TaskClickEnrich = "click:enrich"

type ClickPayload struct {
	LinkID         uuid.UUID `json:"link_id"`
	IP             string
	UserAgent      string
	Referer        string
	AcceptLanguage string
	ClickedAt      time.Time `json:"clicked_at"`
}

type ClickRecorder interface {
	Create(context.Context, *model.ClickEvent) error
}

func NewClickEnrichTask(p ClickPayload) (*asynq.Task, error) {
	b, e := json.Marshal(p)
	if e != nil {
		return nil, e
	}

	return asynq.NewTask(
		TaskClickEnrich,
		b,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
	), nil
}

func (j *JobService) InitClickHandler(r ClickRecorder) {
	j.clicks = r
}

func (j *JobService) handleClickEnrichTask(ctx context.Context, t *asynq.Task) error {
	var p ClickPayload

	if e := json.Unmarshal(t.Payload(), &p); e != nil {
		return e
	}

	sum := sha256.Sum256([]byte(
		p.ClickedAt.UTC().Format("2006-01-02") + j.cfg.Auth.SecretKey,
	))

	ip := sha256.Sum256(
		append([]byte(p.IP), sum[:]...),
	)

	loc := j.geo.Lookup(p.IP)
	d := ua.Parse(p.UserAgent)

	event := &model.ClickEvent{
		LinkID:    p.LinkID,
		ClickedAt: p.ClickedAt,
		IPHash:    hex.EncodeToString(ip[:]),
		IsBot:     d.IsBot,
	}

	event.Country = stringPtr(loc.Country)
	event.Region = stringPtr(loc.Region)
	event.City = stringPtr(loc.City)
	event.Timezone = stringPtr(loc.Timezone)

	event.Browser = stringPtr(d.Browser)
	event.BrowserVersion = stringPtr(d.BrowserVersion)
	event.OS = stringPtr(d.OS)
	event.Device = stringPtr(d.Device)

	ref := classifyReferer(p.Referer)
	event.Referer = &ref

	if p.AcceptLanguage != "" {
		v := strings.TrimSpace(
			strings.Split(p.AcceptLanguage, ",")[0],
		)
		event.Language = &v
	}

	event.UserAgent = stringPtr(p.UserAgent)

	if j.clicks == nil {
		return nil
	}

	return j.clicks.Create(ctx, event)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func classifyReferer(raw string) string {
	if raw == "" {
		return "Direct"
	}

	u, e := url.Parse(raw)
	if e != nil {
		return "Unknown"
	}

	h := strings.ToLower(u.Hostname())

	switch {
	case strings.Contains(h, "google."):
		return "Google"

	case h == "github.com" || strings.HasSuffix(h, ".github.com"):
		return "GitHub"

	case strings.Contains(h, "linkedin.com"):
		return "LinkedIn"

	case h == "x.com" || strings.Contains(h, "twitter.com"):
		return "Twitter"

	default:
		return "Unknown"
	}
}
