package tests

import (
	"testing"

	"eform/backend/internal/domain"
	"eform/backend/internal/utils"
)

func TestCalculateProfileCompletion(t *testing.T) {
	user := domain.User{
		Email:     "user@example.com",
		FullName:  "Test User",
		Phone:     "0812345678",
		KTPNumber: "3173000000000001",
		Documents: []domain.Document{
			{Type: domain.DocumentTypeCV},
			{Type: domain.DocumentTypeKTP},
			{Type: domain.DocumentTypeNPWP},
		},
		Profile: domain.EmployeeProfile{
			Position:              "Engineer",
			CurrentAddress:        "Current Address",
			KTPAddress:            "KTP Address",
			BirthPlace:            "Jakarta",
			Gender:                "male",
			Whatsapp:              "08123456789",
			Education:             "Bachelor",
			InstituteName:         "State University",
			EmergencyContactName:  "Parent",
			EmergencyContactPhone: "08111111111",
		},
	}

	completion := utils.CalculateProfileCompletion(user)
	if completion != 100 {
		t.Fatalf("expected profile completion 100, got %d", completion)
	}
}
