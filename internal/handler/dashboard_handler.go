package handler

import (
	"net/http"

	"eform/backend/internal/service"
	"eform/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// Admin godoc
// @Summary Get admin dashboard
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /dashboard/admin [get]
func (h *DashboardHandler) Admin(c *gin.Context) {
	data, err := h.service.Admin(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin dashboard fetched successfully", data)
}

// User godoc
// @Summary Get user dashboard
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /dashboard/user [get]
func (h *DashboardHandler) User(c *gin.Context) {
	claims := currentClaims(c)
	data, err := h.service.User(c.Request.Context(), claims.UserID.String())
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "user dashboard fetched successfully", data)
}
