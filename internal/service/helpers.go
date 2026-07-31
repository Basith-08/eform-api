package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"eform/backend/internal/domain"
	"eform/backend/internal/repository"
	"eform/backend/internal/utils"
	"eform/backend/pkg/auth"
	"eform/backend/pkg/storage"

	"github.com/google/uuid"
)

type baseService struct {
	repos   *repository.Repositories
	storage *storage.LocalStorage
	jwt     *auth.Manager
}

func calculateProfileCompletion(user domain.User) int {
	return utils.CalculateProfileCompletion(user)
}

func ensureUniqueConstraints(ctx context.Context, repos *repository.Repositories, req domain.EmployeeUpsertRequest, exceptID *uuid.UUID) error {
	emailExists, err := repos.ExistsByEmail(ctx, req.Email, exceptID)
	if err != nil {
		return err
	}

	if emailExists {
		return domain.NewAppError(http.StatusConflict, "email already exists", nil)
	}

	phoneExists, err := repos.ExistsByPhone(ctx, req.Phone, exceptID)
	if err != nil {
		return err
	}

	if phoneExists {
		return domain.NewAppError(http.StatusConflict, "phone already exists", nil)
	}

	ktpExists, err := repos.ExistsByKTP(ctx, req.KTPNumber, exceptID)
	if err != nil {
		return err
	}

	if ktpExists {
		return domain.NewAppError(http.StatusConflict, "ktp number already exists", nil)
	}

	return nil
}

func buildProfile(req domain.EmployeeUpsertRequest) domain.EmployeeProfile {
	return domain.EmployeeProfile{
		JoinDate:                 req.JoinDate,
		EntryDate:                req.EntryDate,
		BirthPlace:               req.BirthPlace,
		BirthDate:                req.BirthDate,
		Gender:                   strings.ToLower(req.Gender),
		Whatsapp:                 req.Whatsapp,
		BankAccount:              req.BankAccount,
		NPWPNumber:               req.NPWPNumber,
		Position:                 req.Position,
		KTPAddress:               req.KTPAddress,
		CurrentAddress:           req.CurrentAddress,
		Religion:                 req.Religion,
		MaritalStatus:            req.MaritalStatus,
		Dependents:               req.Dependents,
		Education:                req.Education,
		BloodType:                req.BloodType,
		InstituteName:            req.InstituteName,
		EmergencyContactName:     req.EmergencyContactName,
		EmergencyContactPhone:    req.EmergencyContactPhone,
		EmergencyContactAddress:  req.EmergencyContactAddress,
		EmergencyContactRelation: req.EmergencyContactRelation,
	}
}

func buildDocuments(userID uuid.UUID, files map[string]*multipart.FileHeader, store *storage.LocalStorage, existing []domain.Document, requireAll bool) ([]domain.Document, error) {
	existingByType := map[string]domain.Document{}
	for _, document := range existing {
		existingByType[document.Type] = document
	}

	requiredTypes := []string{domain.DocumentTypeCV, domain.DocumentTypeKTP, domain.DocumentTypeNPWP}
	result := make([]domain.Document, 0, len(requiredTypes))

	for _, documentType := range requiredTypes {
		file := files[documentType]
		if file != nil {
			document, err := store.SaveDocument(userID, documentType, file)
			if err != nil {
				return nil, err
			}

			if previous, ok := existingByType[documentType]; ok {
				_ = store.DeleteFile(previous.FilePath)
			}

			result = append(result, document)
			continue
		}

		existingDocument, ok := existingByType[documentType]
		if !ok && requireAll {
			return nil, domain.NewAppError(http.StatusBadRequest, fmt.Sprintf("%s document is required", strings.ToUpper(documentType)), nil)
		}

		if ok {
			result = append(result, existingDocument)
		}
	}

	return result, nil
}

func createAuditLog(ctx context.Context, repos *repository.Repositories, actorID *uuid.UUID, targetID *uuid.UUID, action, resource, description string, payload any, meta domain.RequestMeta) error {
	payloadBytes, _ := json.Marshal(payload)
	return repos.CreateAuditLog(ctx, &domain.AuditLog{
		ActorUserID:  actorID,
		TargetUserID: targetID,
		Action:       action,
		Resource:     resource,
		Description:  description,
		IPAddress:    meta.IPAddress,
		UserAgent:    meta.UserAgent,
		Payload:      string(payloadBytes),
		CreatedAt:    time.Now(),
	})
}
