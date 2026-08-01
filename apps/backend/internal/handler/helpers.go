package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/errs"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/middleware"
)

func appUser(c *gin.Context) (uuid.UUID, bool) {
	id, ok := middleware.AppUserID(c)
	if !ok {
		_ = c.Error(errs.NewUnauthorizedError("Unauthorized", false))
		c.Abort()

		return uuid.Nil, false
	}

	return id, true
}

func requestID(c *gin.Context) (uuid.UUID, bool) {
	id, e := uuid.Parse(c.Param("id"))
	if e != nil {
		_ = c.Error(errs.NewBadRequestError("invalid link id", false, nil, nil, nil))

		return uuid.Nil, false
	}

	return id, true
}

func respondErr(c *gin.Context, e error) {
	_ = c.Error(e)

	c.Abort()
}
