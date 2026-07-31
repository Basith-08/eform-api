package handler

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"eform/backend/internal/domain"
	"eform/backend/internal/service"
	"eform/backend/internal/validator"
	"eform/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	service     *service.EmployeeService
	authService *service.AuthService
	validator   *validator.Validator
}

func NewEmployeeHandler(service *service.EmployeeService, authService *service.AuthService, validator *validator.Validator) *EmployeeHandler {
	return &EmployeeHandler{
		service:     service,
		authService: authService,
		validator:   validator,
	}
}

// List godoc
// @Summary List employees
// @Tags Employees
// @Security BearerAuth
// @Produce json
// @Param search query string false "Search term"
// @Param status query string false "Status"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Param sortBy query string false "Sort column"
// @Param sortOrder query string false "Sort order"
// @Success 200 {object} response.Envelope
// @Router /employees [get]
func (h *EmployeeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := domain.EmployeeListFilter{
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		Role:      domain.RoleUser,
		Page:      page,
		Limit:     limit,
		SortBy:    c.DefaultQuery("sortBy", "createdAt"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	employees, total, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employees fetched successfully", gin.H{
		"items": employees,
		"pagination": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
		},
	})
}

// Detail godoc
// @Summary Get employee detail
// @Tags Employees
// @Security BearerAuth
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} response.Envelope
// @Router /employees/{id} [get]
func (h *EmployeeHandler) Detail(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}

	employee, err := h.service.Detail(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employee detail fetched successfully", employee)
}

// Create godoc
// @Summary Create employee
// @Tags Employees
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Success 201 {object} response.Envelope
// @Router /employees [post]
func (h *EmployeeHandler) Create(c *gin.Context) {
	req, err := validator.ParseEmployeeForm(c, true)
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
	employee, err := h.service.Create(c.Request.Context(), req, files, claims.UserID, requestMeta(c))
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "employee created successfully", employee)
}

// Update godoc
// @Summary Update employee
// @Tags Employees
// @Security BearerAuth
// @Accept mpfd
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} response.Envelope
// @Router /employees/{id} [put]
func (h *EmployeeHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}

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
	employee, err := h.service.Update(c.Request.Context(), id, req, files, claims.UserID, requestMeta(c))
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employee updated successfully", employee)
}

// Delete godoc
// @Summary Delete employee
// @Tags Employees
// @Security BearerAuth
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} response.Envelope
// @Router /employees/{id} [delete]
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}

	claims := currentClaims(c)
	if err := h.service.Delete(c.Request.Context(), id, claims.UserID, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employee deleted successfully", nil)
}

// Activate godoc
// @Summary Activate employee
// @Tags Employees
// @Security BearerAuth
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} response.Envelope
// @Router /employees/{id}/activate [patch]
func (h *EmployeeHandler) Activate(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}

	claims := currentClaims(c)
	if err := h.service.SetStatus(c.Request.Context(), id, claims.UserID, true, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employee activated successfully", nil)
}

// Deactivate godoc
// @Summary Deactivate employee
// @Tags Employees
// @Security BearerAuth
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} response.Envelope
// @Router /employees/{id}/deactivate [patch]
func (h *EmployeeHandler) Deactivate(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}

	claims := currentClaims(c)
	if err := h.service.SetStatus(c.Request.Context(), id, claims.UserID, false, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employee deactivated successfully", nil)
}

// ResetPassword godoc
// @Summary Reset employee password
// @Tags Employees
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param payload body domain.ResetEmployeePasswordRequest true "Reset password payload"
// @Success 200 {object} response.Envelope
// @Router /employees/{id}/reset-password [post]
func (h *EmployeeHandler) ResetPassword(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}

	var req domain.ResetEmployeePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid reset password payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	claims := currentClaims(c)
	if err := h.authService.ResetEmployeePassword(c.Request.Context(), claims.UserID, id, req.NewPassword, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "employee password reset successfully", nil)
}
