package service

import (
	"context"
	"time"

	"eform/backend/internal/domain"
	"eform/backend/internal/repository"

	"github.com/google/uuid"
)

type DashboardService struct {
	repos *repository.Repositories
}

func NewDashboardService(repos *repository.Repositories) *DashboardService {
	return &DashboardService{repos: repos}
}

func (s *DashboardService) Admin(ctx context.Context) (*domain.AdminDashboardResponse, error) {
	totalEmployees, err := s.repos.CountEmployees(ctx)
	if err != nil {
		return nil, err
	}

	newEmployees, err := s.repos.CountNewEmployees(ctx, time.Now().AddDate(0, 0, -30))
	if err != nil {
		return nil, err
	}

	activeEmployees, err := s.repos.CountEmployeesByStatus(ctx, domain.UserStatusActive)
	if err != nil {
		return nil, err
	}

	inactiveEmployees, err := s.repos.CountEmployeesByStatus(ctx, domain.UserStatusInactive)
	if err != nil {
		return nil, err
	}

	latestEmployees, err := s.repos.LatestEmployees(ctx, 5)
	if err != nil {
		return nil, err
	}

	items := make([]domain.EmployeeResponse, 0, len(latestEmployees))
	for _, employee := range latestEmployees {
		items = append(items, mapEmployee(employee))
	}

	return &domain.AdminDashboardResponse{
		TotalEmployees:    totalEmployees,
		NewEmployees:      newEmployees,
		ActiveEmployees:   activeEmployees,
		InactiveEmployees: inactiveEmployees,
		LatestEmployees:   items,
	}, nil
}

func (s *DashboardService) User(ctx context.Context, userID string) (*domain.UserDashboardResponse, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewAppError(400, "invalid user identifier", err)
	}

	user, err := s.repos.FindUserByID(ctx, parsedID)
	if err != nil {
		return nil, err
	}

	documents := make([]domain.DocumentResponse, 0, len(user.Documents))
	for _, document := range user.Documents {
		documents = append(documents, mapDocument(document))
	}

	profile := mapEmployee(user)
	return &domain.UserDashboardResponse{
		Profile:           profile,
		ProfileCompletion: profile.ProfileCompletion,
		UploadedDocuments: documents,
	}, nil
}
