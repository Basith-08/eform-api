package utils

import "eform/backend/internal/domain"

func CalculateProfileCompletion(user domain.User) int {
	completed := 0
	total := 15

	fields := []bool{
		user.Email != "",
		user.FullName != "",
		user.Phone != "",
		user.KTPNumber != "",
		user.Profile.Position != "",
		user.Profile.CurrentAddress != "",
		user.Profile.KTPAddress != "",
		user.Profile.BirthPlace != "",
		user.Profile.Gender != "",
		user.Profile.Whatsapp != "",
		user.Profile.Education != "",
		user.Profile.InstituteName != "",
		user.Profile.EmergencyContactName != "",
		user.Profile.EmergencyContactPhone != "",
		len(user.Documents) >= 3,
	}

	for _, ok := range fields {
		if ok {
			completed++
		}
	}

	return int(float64(completed) / float64(total) * 100)
}
