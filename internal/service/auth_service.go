package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"eform/backend/config"
	"eform/backend/internal/domain"
	"eform/backend/internal/repository"
	"eform/backend/internal/utils"
	"eform/backend/pkg/auth"
	"eform/backend/pkg/storage"

	"github.com/google/uuid"
)

type AuthService struct {
	cfg     config.Config
	repos   *repository.Repositories
	storage *storage.LocalStorage
	jwt     *auth.Manager
}

func NewAuthService(cfg config.Config, repos *repository.Repositories, store *storage.LocalStorage, jwtManager *auth.Manager) *AuthService {
	return &AuthService{
		cfg:     cfg,
		repos:   repos,
		storage: store,
		jwt:     jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest, files map[string]*multipart.FileHeader, meta domain.RequestMeta) (*domain.AuthResponse, error) {
	upsertRequest := domain.EmployeeUpsertRequest(req)
	if err := ensureUniqueConstraints(ctx, s.repos, upsertRequest, nil); err != nil {
		return nil, err
	}

	role, err := s.repos.FindRoleByCode(ctx, domain.RoleUser)
	if err != nil {
		return nil, err
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Email:        strings.ToLower(req.Email),
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		Phone:        req.Phone,
		KTPNumber:    req.KTPNumber,
		Status:       domain.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	profile := buildProfile(upsertRequest)
	documents, err := buildDocuments(user.ID, files, s.storage, nil, true)
	if err != nil {
		return nil, err
	}

	if err := s.repos.CreateUser(ctx, &user, &profile, documents); err != nil {
		return nil, err
	}

	user.Role = role
	user.Profile = profile
	user.Documents = documents

	tokenPair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := createAuditLog(ctx, s.repos, &user.ID, &user.ID, "register", "user", "employee registration completed", user, meta); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: tokenPair,
		User:  mapEmployee(user),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest, meta domain.RequestMeta) (*domain.AuthResponse, error) {
	user, err := s.repos.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, domain.NewAppError(http.StatusUnauthorized, "invalid email or password", err)
		}
		return nil, err
	}

	if user.Status != domain.UserStatusActive {
		return nil, domain.NewAppError(http.StatusForbidden, "user account is inactive", nil)
	}

	if err := utils.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, domain.NewAppError(http.StatusUnauthorized, "invalid email or password", err)
	}

	loginAt := time.Now()
	if err := s.repos.SetLastLogin(ctx, user.ID, loginAt); err != nil {
		return nil, err
	}

	user.LastLoginAt = &loginAt
	tokenPair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := createAuditLog(ctx, s.repos, &user.ID, &user.ID, "login", "auth", "user login successful", map[string]string{"email": user.Email}, meta); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: tokenPair,
		User:  mapEmployee(user),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	claims, err := s.jwt.Parse(refreshToken)
	if err != nil {
		return domain.TokenPair{}, domain.NewAppError(http.StatusUnauthorized, "invalid refresh token", err)
	}

	if claims.Type != "refresh" {
		return domain.TokenPair{}, domain.NewAppError(http.StatusUnauthorized, "invalid token type", nil)
	}

	storedToken, err := s.repos.FindRefreshToken(ctx, auth.HashToken(refreshToken))
	if err != nil {
		return domain.TokenPair{}, domain.NewAppError(http.StatusUnauthorized, "refresh token not found", err)
	}

	if storedToken.RevokedAt != nil || storedToken.ExpiresAt.Before(time.Now()) {
		return domain.TokenPair{}, domain.NewAppError(http.StatusUnauthorized, "refresh token expired", nil)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return domain.TokenPair{}, err
	}

	user, err := s.repos.FindUserByID(ctx, userID)
	if err != nil {
		return domain.TokenPair{}, err
	}

	if err := s.repos.RevokeRefreshToken(ctx, auth.HashToken(refreshToken)); err != nil {
		return domain.TokenPair{}, err
	}

	return s.issueTokenPair(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, meta domain.RequestMeta) error {
	claims, err := s.jwt.Parse(refreshToken)
	if err != nil {
		return domain.NewAppError(http.StatusUnauthorized, "invalid refresh token", err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return err
	}

	if err := s.repos.RevokeRefreshToken(ctx, auth.HashToken(refreshToken)); err != nil {
		return err
	}

	return createAuditLog(ctx, s.repos, &userID, &userID, "logout", "auth", "user logout", nil, meta)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (map[string]string, error) {
	_, err := s.repos.FindUserByEmail(ctx, email)
	if err != nil && !repository.IsNotFound(err) {
		return nil, err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	token := hex.EncodeToString(tokenBytes)
	reset := domain.PasswordReset{
		Email:     strings.ToLower(email),
		TokenHash: auth.HashToken(token),
		ExpiresAt: time.Now().Add(30 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repos.CreatePasswordReset(ctx, &reset); err != nil {
		return nil, err
	}

	data := map[string]string{}
	if s.cfg.AppEnv != "production" {
		data["resetToken"] = token
	}

	return data, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token string, newPassword string, meta domain.RequestMeta) error {
	reset, err := s.repos.FindValidPasswordReset(ctx, auth.HashToken(token))
	if err != nil {
		return domain.NewAppError(http.StatusUnauthorized, "invalid or expired reset token", err)
	}

	user, err := s.repos.FindUserByEmail(ctx, reset.Email)
	if err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.repos.UpdatePassword(ctx, user.ID, passwordHash); err != nil {
		return err
	}

	if err := s.repos.MarkPasswordResetUsed(ctx, reset.ID); err != nil {
		return err
	}

	return createAuditLog(ctx, s.repos, &user.ID, &user.ID, "password_reset", "auth", "password reset completed", nil, meta)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req domain.ChangePasswordRequest, meta domain.RequestMeta) error {
	user, err := s.repos.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := utils.ComparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return domain.NewAppError(http.StatusUnauthorized, "current password is incorrect", err)
	}

	passwordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := s.repos.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return err
	}

	return createAuditLog(ctx, s.repos, &userID, &userID, "password_change", "auth", "password changed", nil, meta)
}

func (s *AuthService) ResetEmployeePassword(ctx context.Context, actorID uuid.UUID, targetID uuid.UUID, password string, meta domain.RequestMeta) error {
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	if err := s.repos.UpdatePassword(ctx, targetID, passwordHash); err != nil {
		return err
	}

	return createAuditLog(ctx, s.repos, &actorID, &targetID, "admin_password_reset", "user", "admin reset employee password", nil, meta)
}

func (s *AuthService) issueTokenPair(ctx context.Context, user domain.User) (domain.TokenPair, error) {
	tokenPair, err := s.jwt.GenerateTokenPair(domain.AuthClaims{
		UserID: user.ID,
		Role:   user.Role.Code,
		Email:  user.Email,
	})
	if err != nil {
		return domain.TokenPair{}, err
	}

	if err := s.repos.CreateRefreshToken(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: auth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		return domain.TokenPair{}, err
	}

	return tokenPair, nil
}
