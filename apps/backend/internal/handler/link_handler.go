package handler

import (
	"encoding/csv"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/model"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/repository"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/server"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LinkHandler struct {
	Handler
	links     *service.LinkService
	cfgDomain string
}

func NewLinkHandler(s *server.Server, links *service.LinkService) *LinkHandler {
	return &LinkHandler{Handler: NewHandler(s), links: links, cfgDomain: s.Config.Service.Domain}
}

type createLinkRequest struct {
	URL         string     `json:"url" binding:"required"`
	Alias       string     `json:"alias"`
	Expiry      *int       `json:"expiry"`
	Password    string     `json:"password"`
	Tags        []string   `json:"tags"`
	ClickLimit  *int       `json:"clickLimit"`
	ActivatesAt *time.Time `json:"activatesAt"`
}
type updateLinkRequest struct {
	URL, Password                    *string
	Tags                             *[]string
	IsFavorite, IsArchived, IsActive *bool
	ClickLimit                       *int
	ActivatesAt, ExpiresAt           *time.Time
}

func (r updateLinkRequest) input() service.UpdateLinkInput {
	return service.UpdateLinkInput{URL: r.URL, Password: r.Password, Tags: r.Tags, IsFavorite: r.IsFavorite, IsArchived: r.IsArchived, IsActive: r.IsActive, ClickLimit: r.ClickLimit, ActivatesAt: r.ActivatesAt, ExpiresAt: r.ExpiresAt}
}
func linkResponse(l *model.Link, domain string) gin.H {
	return gin.H{"id": l.ID, "shortCode": l.ShortCode, "shortUrl": "https://" + domain + "/" + l.ShortCode, "originalUrl": l.OriginalURL, "tags": l.Tags, "isActive": l.IsActive, "isFavorite": l.IsFavorite, "isArchived": l.IsArchived, "clickLimit": l.ClickLimit, "activatesAt": l.ActivatesAt, "expiresAt": l.ExpiresAt, "createdAt": l.CreatedAt, "updatedAt": l.UpdatedAt, "qrCode": "/api/v1/links/" + l.ID.String() + "/qrcode"}
}
func (h *LinkHandler) Create(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	var r createLinkRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		respondErr(c, serviceBad(e))
		return
	}
	l, e := h.links.Create(c.Request.Context(), u, service.CreateLinkInput{URL: r.URL, Alias: r.Alias, ExpiryHours: r.Expiry, Password: r.Password, Tags: r.Tags, ClickLimit: r.ClickLimit, ActivatesAt: r.ActivatesAt})
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusCreated, linkResponse(l, h.cfgDomain))
}
func serviceBad(e error) error { return errs.NewBadRequestError(e.Error(), false, nil, nil, nil) }
func filters(c *gin.Context) repository.ListFilters {
	f := repository.ListFilters{Search: c.Query("search"), Tag: c.Query("tag"), Sort: c.Query("sort")}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	for key, target := range map[string]**bool{"active": &f.Active, "expired": &f.Expired} {
		if raw, ok := c.GetQuery(key); ok {
			v, e := strconv.ParseBool(raw)
			if e == nil {
				*target = &v
			}
		}
	}
	return f
}
func (h *LinkHandler) List(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	f := filters(c)
	ls, total, e := h.links.List(c.Request.Context(), u, f, false)
	if e != nil {
		respondErr(c, e)
		return
	}
	data := make([]gin.H, 0, len(ls))
	for i := range ls {
		data = append(data, linkResponse(&ls[i], h.cfgDomain))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "page": f.Page, "limit": f.Limit, "total": total, "totalPages": (total + f.Limit - 1) / f.Limit})
}
func (h *LinkHandler) Get(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	l, e := h.links.Get(c.Request.Context(), u, id)
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusOK, linkResponse(l, h.cfgDomain))
}
func (h *LinkHandler) Update(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	var r updateLinkRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		respondErr(c, serviceBad(e))
		return
	}
	l, e := h.links.Update(c.Request.Context(), u, id, r.input())
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusOK, linkResponse(l, h.cfgDomain))
}
func (h *LinkHandler) Delete(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	if e := h.links.Delete(c.Request.Context(), u, id); e != nil {
		respondErr(c, e)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *LinkHandler) Toggle(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	var r struct {
		IsActive bool `json:"isActive"`
	}
	if e := c.ShouldBindJSON(&r); e != nil {
		respondErr(c, serviceBad(e))
		return
	}
	if e := h.links.Toggle(c.Request.Context(), u, id, r.IsActive); e != nil {
		respondErr(c, e)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *LinkHandler) Duplicate(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	id, ok := requestID(c)
	if !ok {
		return
	}
	l, e := h.links.Duplicate(c.Request.Context(), u, id)
	if e != nil {
		respondErr(c, e)
		return
	}
	c.JSON(http.StatusCreated, linkResponse(l, h.cfgDomain))
}
func (h *LinkHandler) BulkDelete(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	var r struct {
		IDs []uuid.UUID `json:"ids" binding:"required"`
	}
	if e := c.ShouldBindJSON(&r); e != nil {
		respondErr(c, serviceBad(e))
		return
	}
	if e := h.links.BulkDelete(c.Request.Context(), u, r.IDs); e != nil {
		respondErr(c, e)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *LinkHandler) Export(c *gin.Context) {
	u, ok := appUser(c)
	if !ok {
		return
	}
	ls, _, e := h.links.List(c.Request.Context(), u, filters(c), true)
	if e != nil {
		respondErr(c, e)
		return
	}
	if c.DefaultQuery("format", "csv") == "json" {
		c.JSON(http.StatusOK, ls)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=links.csv")
	c.Header("Content-Type", "text/csv")
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"id", "short_code", "original_url", "tags", "active", "expires_at", "created_at"})
	for _, l := range ls {
		exp := ""
		if l.ExpiresAt != nil {
			exp = l.ExpiresAt.Format(time.RFC3339)
		}
		_ = w.Write([]string{l.ID.String(), l.ShortCode, l.OriginalURL, strings.Join(l.Tags, "|"), strconv.FormatBool(l.IsActive), exp, l.CreatedAt.Format(time.RFC3339)})
	}
	w.Flush()
}
