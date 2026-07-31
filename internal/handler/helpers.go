package handler

import (
	"errors"
	"net/http"

	"eform/backend/internal/domain"
	"eform/backend/internal/middleware"
	"eform/backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeError(c *gin.Context, err error) {
	var appError *domain.AppError
	if errors.As(err, &appError) {
		response.Error(c, appError.StatusCode, appError.Message)
		return
	}

	response.Error(c, http.StatusInternalServerError, "internal server error")
}

func requestMeta(c *gin.Context) domain.RequestMeta {
	return domain.RequestMeta{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func currentClaims(c *gin.Context) domain.AuthClaims {
	claims, _ := middleware.GetClaims(c)
	return claims
}

func parseUUIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		return uuid.Nil, domain.NewAppError(http.StatusBadRequest, "invalid identifier", err)
	}

	return id, nil
}
