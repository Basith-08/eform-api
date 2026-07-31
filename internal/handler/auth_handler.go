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

type AuthHandler struct {
	service   *service.AuthService
	validator *validator.Validator
}

func NewAuthHandler(service *service.AuthService, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		service:   service,
		validator: validator,
	}
}

// Register godoc
// @Summary Register new employee
// @Tags Auth
// @Accept mpfd
// @Produce json
// @Param email formData string true "Email"
// @Param password formData string true "Password"
// @Param fullName formData string true "Full Name"
// @Param phone formData string true "Phone Number"
// @Param ktpNumber formData string true "KTP Number"
// @Param cv formData file true "CV document"
// @Param ktp formData file true "KTP image"
// @Param npwp formData file true "NPWP image"
// @Success 201 {object} response.Envelope
// @Failure 400 {object} response.Envelope
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	req, err := validator.ParseEmployeeForm(c, true)
	if err != nil {
		writeError(c, err)
		return
	}

	registerRequest := domain.RegisterRequest(req)
	if err := h.validator.Struct(registerRequest); err != nil {
		writeError(c, err)
		return
	}

	files := map[string]*multipart.FileHeader{
		domain.DocumentTypeCV:   fileFromRequest(c, domain.DocumentTypeCV),
		domain.DocumentTypeKTP:  fileFromRequest(c, domain.DocumentTypeKTP),
		domain.DocumentTypeNPWP: fileFromRequest(c, domain.DocumentTypeNPWP),
	}

	authResponse, err := h.service.Register(c.Request.Context(), registerRequest, files, requestMeta(c))
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "registration successful", authResponse)
}

// Login godoc
// @Summary Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body domain.LoginRequest true "Login payload"
// @Success 200 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid login payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	authResponse, err := h.service.Login(c.Request.Context(), req, requestMeta(c))
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "login successful", authResponse)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body domain.RefreshRequest true "Refresh payload"
// @Success 200 {object} response.Envelope
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid refresh payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	tokenPair, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "token refreshed", tokenPair)
}

// Logout godoc
// @Summary Logout
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body domain.RefreshRequest true "Logout payload"
// @Success 200 {object} response.Envelope
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid logout payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "logout successful", nil)
}

// ForgotPassword godoc
// @Summary Forgot password
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body domain.ForgotPasswordRequest true "Forgot password payload"
// @Success 200 {object} response.Envelope
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid forgot password payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	data, err := h.service.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "if the email exists, reset instructions have been generated", data)
}

// ResetPassword godoc
// @Summary Reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body domain.ResetPasswordRequest true "Reset password payload"
// @Success 200 {object} response.Envelope
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid reset password payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "password reset successful", nil)
}

// ChangePassword godoc
// @Summary Change own password
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body domain.ChangePasswordRequest true "Change password payload"
// @Success 200 {object} response.Envelope
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid change password payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		writeError(c, err)
		return
	}

	claims := currentClaims(c)
	if err := h.service.ChangePassword(c.Request.Context(), claims.UserID, req, requestMeta(c)); err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "password changed successfully", nil)
}

func fileFromRequest(c *gin.Context, name string) *multipart.FileHeader {
	file, err := c.FormFile(name)
	if err != nil {
		return nil
	}

	return file
}
