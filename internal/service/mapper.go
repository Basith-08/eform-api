package service

import "eform/backend/internal/domain"

func mapDocument(document domain.Document) domain.DocumentResponse {
	return domain.DocumentResponse{
		ID:        document.ID,
		Type:      document.Type,
		FileName:  document.FileName,
		FilePath:  document.FilePath,
		MimeType:  document.MimeType,
		SizeBytes: document.SizeBytes,
	}
}

func mapEmployee(user domain.User) domain.EmployeeResponse {
	documents := make([]domain.DocumentResponse, 0, len(user.Documents))
	for _, document := range user.Documents {
		documents = append(documents, mapDocument(document))
	}

	return domain.EmployeeResponse{
		ID:                       user.ID,
		Role:                     user.Role.Code,
		Status:                   user.Status,
		Email:                    user.Email,
		FullName:                 user.FullName,
		Phone:                    user.Phone,
		KTPNumber:                user.KTPNumber,
		JoinDate:                 user.Profile.JoinDate,
		EntryDate:                user.Profile.EntryDate,
		BirthPlace:               user.Profile.BirthPlace,
		BirthDate:                user.Profile.BirthDate,
		Gender:                   user.Profile.Gender,
		Whatsapp:                 user.Profile.Whatsapp,
		BankAccount:              user.Profile.BankAccount,
		NPWPNumber:               user.Profile.NPWPNumber,
		Position:                 user.Profile.Position,
		KTPAddress:               user.Profile.KTPAddress,
		CurrentAddress:           user.Profile.CurrentAddress,
		Religion:                 user.Profile.Religion,
		MaritalStatus:            user.Profile.MaritalStatus,
		Dependents:               user.Profile.Dependents,
		Education:                user.Profile.Education,
		BloodType:                user.Profile.BloodType,
		InstituteName:            user.Profile.InstituteName,
		EmergencyContactName:     user.Profile.EmergencyContactName,
		EmergencyContactPhone:    user.Profile.EmergencyContactPhone,
		EmergencyContactAddress:  user.Profile.EmergencyContactAddress,
		EmergencyContactRelation: user.Profile.EmergencyContactRelation,
		LastLoginAt:              user.LastLoginAt,
		ProfileCompletion:        calculateProfileCompletion(user),
		Documents:                documents,
		CreatedAt:                user.CreatedAt,
		UpdatedAt:                user.UpdatedAt,
	}
}
