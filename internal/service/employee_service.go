package service

import (
	"context"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"eform/backend/internal/domain"
	"eform/backend/internal/repository"
	"eform/backend/internal/utils"
	"eform/backend/pkg/storage"

	"github.com/google/uuid"
)

type EmployeeService struct {
	repos   *repository.Repositories
	storage *storage.LocalStorage
}

func NewEmployeeService(repos *repository.Repositories, store *storage.LocalStorage) *EmployeeService {
	return &EmployeeService{
		repos:   repos,
		storage: store,
	}
}

func (s *EmployeeService) List(ctx context.Context, filter domain.EmployeeListFilter) ([]domain.EmployeeResponse, int64, error) {
	users, total, err := s.repos.ListEmployees(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]domain.EmployeeResponse, 0, len(users))
	for _, user := range users {
		result = append(result, mapEmployee(user))
	}

	return result, total, nil
}

func (s *EmployeeService) Detail(ctx context.Context, userID uuid.UUID) (*domain.EmployeeResponse, error) {
	user, err := s.repos.FindUserByID(ctx, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, domain.NewAppError(http.StatusNotFound, "employee not found", err)
		}
		return nil, err
	}

	response := mapEmployee(user)
	return &response, nil
}

func (s *EmployeeService) Create(ctx context.Context, req domain.EmployeeUpsertRequest, files map[string]*multipart.FileHeader, actorID uuid.UUID, meta domain.RequestMeta) (*domain.EmployeeResponse, error) {
	if len(req.Password) < 8 {
		return nil, domain.NewAppError(http.StatusBadRequest, "password must be at least 8 characters", nil)
	}

	if err := ensureUniqueConstraints(ctx, s.repos, req, nil); err != nil {
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
	profile := buildProfile(req)
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

	if err := createAuditLog(ctx, s.repos, &actorID, &user.ID, "create", "employee", "employee created by admin", user, meta); err != nil {
		return nil, err
	}

	response := mapEmployee(user)
	return &response, nil
}

func (s *EmployeeService) Update(ctx context.Context, userID uuid.UUID, req domain.EmployeeUpsertRequest, files map[string]*multipart.FileHeader, actorID uuid.UUID, meta domain.RequestMeta) (*domain.EmployeeResponse, error) {
	user, err := s.repos.FindUserByID(ctx, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, domain.NewAppError(http.StatusNotFound, "employee not found", err)
		}
		return nil, err
	}

	if err := ensureUniqueConstraints(ctx, s.repos, req, &userID); err != nil {
		return nil, err
	}

	user.Email = strings.ToLower(req.Email)
	user.FullName = req.FullName
	user.Phone = req.Phone
	user.KTPNumber = req.KTPNumber
	user.UpdatedAt = time.Now()

	if req.Password != "" {
		passwordHash, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = passwordHash
	}

	profile := buildProfile(req)
	documents, err := buildDocuments(user.ID, files, s.storage, user.Documents, false)
	if err != nil {
		return nil, err
	}

	if err := s.repos.UpdateUserAggregate(ctx, &user, &profile, documents); err != nil {
		return nil, err
	}

	user.Profile = profile
	user.Documents = documents

	if err := createAuditLog(ctx, s.repos, &actorID, &userID, "update", "employee", "employee updated", user, meta); err != nil {
		return nil, err
	}

	response := mapEmployee(user)
	return &response, nil
}

func (s *EmployeeService) Delete(ctx context.Context, userID uuid.UUID, actorID uuid.UUID, meta domain.RequestMeta) error {
	if err := s.repos.DeleteUser(ctx, userID); err != nil {
		return err
	}

	return createAuditLog(ctx, s.repos, &actorID, &userID, "delete", "employee", "employee deleted", nil, meta)
}

func (s *EmployeeService) SetStatus(ctx context.Context, userID uuid.UUID, actorID uuid.UUID, active bool, meta domain.RequestMeta) error {
	status := domain.UserStatusInactive
	action := "deactivate"
	description := "employee deactivated"
	if active {
		status = domain.UserStatusActive
		action = "activate"
		description = "employee activated"
	}

	if err := s.repos.UpdateUserStatus(ctx, userID, status); err != nil {
		return err
	}

	return createAuditLog(ctx, s.repos, &actorID, &userID, action, "employee", description, map[string]string{"status": status}, meta)
}

func (s *EmployeeService) UpdateOwnProfile(ctx context.Context, userID uuid.UUID, req domain.EmployeeUpsertRequest, files map[string]*multipart.FileHeader, meta domain.RequestMeta) (*domain.EmployeeResponse, error) {
	user, err := s.repos.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(user.Email, req.Email) {
		return nil, domain.NewAppError(http.StatusForbidden, "email cannot be changed from profile page", nil)
	}

	if err := ensureUniqueConstraints(ctx, s.repos, req, &userID); err != nil {
		return nil, err
	}

	user.FullName = req.FullName
	user.Phone = req.Phone
	user.KTPNumber = req.KTPNumber
	user.UpdatedAt = time.Now()

	profile := buildProfile(req)
	documents, err := buildDocuments(user.ID, files, s.storage, user.Documents, false)
	if err != nil {
		return nil, err
	}

	if err := s.repos.UpdateUserAggregate(ctx, &user, &profile, documents); err != nil {
		return nil, err
	}

	user.Profile = profile
	user.Documents = documents

	if err := createAuditLog(ctx, s.repos, &userID, &userID, "update", "profile", "user updated own profile", user, meta); err != nil {
		return nil, err
	}

	response := mapEmployee(user)
	return &response, nil
}
