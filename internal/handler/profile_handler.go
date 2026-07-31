package handler

import (
	"mime/multipart"
	"net/http"

	"eform/backend/internal/domain"
	"eform/backend/internal/service"
	"eform/backend/internal/validator"
	"eform/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	service   *service.EmployeeService
	validator *validator.Validator
}

func NewProfileHandler(service *service.EmployeeService, validator *validator.Validator) *ProfileHandler {
	return &ProfileHandler{
		service:   service,
		validator: validator,
	}
}

// Me godoc
// @Summary Get own profile
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /profile/me [get]
func (h *ProfileHandler) Me(c *gin.Context) {
	claims := currentClaims(c)
	data, err := h.service.Detail(c.Request.Context(), claims.UserID)
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "profile fetched successfully", data)
}

// UpdateMe godoc
// @Summary Update own profile
// @Tags Profile
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /profile/me [put]
func (h *ProfileHandler) UpdateMe(c *gin.Context) {
	req, err := validator.ParseEmployeeForm(c, false)
	if err != nil {
		writeError(c, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	files := map[string]*multipart.FileHeader{
		domain.DocumentTypeCV:   fileFromRequest(c, domain.DocumentTypeCV),
		domain.DocumentTypeKTP:  fileFromRequest(c, domain.DocumentTypeKTP),
		domain.DocumentTypeNPWP: fileFromRequest(c, domain.DocumentTypeNPWP),
	}

	claims := currentClaims(c)
	data, err := h.service.UpdateOwnProfile(c.Request.Context(), claims.UserID, req, files, requestMeta(c))
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "profile updated successfully", data)
}
