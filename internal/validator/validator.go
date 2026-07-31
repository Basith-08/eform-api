package validator

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eform/backend/internal/domain"

	"github.com/gin-gonic/gin"
	playground "github.com/go-playground/validator/v10"
)

type Validator struct {
	engine *playground.Validate
}

func New() *Validator {
	return &Validator{
		engine: playground.New(),
	}
}

func (v *Validator) Struct(payload any) error {
	if err := v.engine.Struct(payload); err != nil {
		var validationErrors playground.ValidationErrors
		if errors.As(err, &validationErrors) {
			messages := make([]string, 0, len(validationErrors))
			for _, validationError := range validationErrors {
				messages = append(messages, fmt.Sprintf("%s is %s", strings.ToLower(validationError.Field()), validationError.Tag()))
			}

			return domain.NewAppError(http.StatusBadRequest, strings.Join(messages, ", "), err)
		}

		return domain.NewAppError(http.StatusBadRequest, "invalid request payload", err)
	}

	return nil
}

func ParseEmployeeForm(c *gin.Context, passwordRequired bool) (domain.EmployeeUpsertRequest, error) {
	req := domain.EmployeeUpsertRequest{
		Email:                    strings.TrimSpace(c.PostForm("email")),
		Password:                 c.PostForm("password"),
		FullName:                 strings.TrimSpace(c.PostForm("fullName")),
		Phone:                    strings.TrimSpace(c.PostForm("phone")),
		KTPNumber:                strings.TrimSpace(c.PostForm("ktpNumber")),
		BirthPlace:               strings.TrimSpace(c.PostForm("birthPlace")),
		Gender:                   strings.TrimSpace(c.PostForm("gender")),
		Whatsapp:                 strings.TrimSpace(c.PostForm("whatsapp")),
		BankAccount:              strings.TrimSpace(c.PostForm("bankAccount")),
		NPWPNumber:               strings.TrimSpace(c.PostForm("npwpNumber")),
		Position:                 strings.TrimSpace(c.PostForm("position")),
		KTPAddress:               strings.TrimSpace(c.PostForm("ktpAddress")),
		CurrentAddress:           strings.TrimSpace(c.PostForm("currentAddress")),
		Religion:                 strings.TrimSpace(c.PostForm("religion")),
		MaritalStatus:            strings.TrimSpace(c.PostForm("maritalStatus")),
		Education:                strings.TrimSpace(c.PostForm("education")),
		BloodType:                strings.TrimSpace(c.PostForm("bloodType")),
		InstituteName:            strings.TrimSpace(c.PostForm("instituteName")),
		EmergencyContactName:     strings.TrimSpace(c.PostForm("emergencyContactName")),
		EmergencyContactPhone:    strings.TrimSpace(c.PostForm("emergencyContactPhone")),
		EmergencyContactAddress:  strings.TrimSpace(c.PostForm("emergencyContactAddress")),
		EmergencyContactRelation: strings.TrimSpace(c.PostForm("emergencyContactRelation")),
	}

	var err error
	req.JoinDate, err = parseDate(c.PostForm("joinDate"))
	if err != nil {
		return req, err
	}

	req.EntryDate, err = parseDate(c.PostForm("entryDate"))
	if err != nil {
		return req, err
	}

	req.BirthDate, err = parseDate(c.PostForm("birthDate"))
	if err != nil {
		return req, err
	}

	req.Dependents, err = parseInt(c.PostForm("dependents"))
	if err != nil {
		return req, err
	}

	if passwordRequired && len(req.Password) < 8 {
		return req, domain.NewAppError(http.StatusBadRequest, "password must be at least 8 characters", nil)
	}

	return req, nil
}

func parseDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, domain.NewAppError(http.StatusBadRequest, "invalid date format, use YYYY-MM-DD", err)
	}

	return &parsed, nil
}

func parseInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, domain.NewAppError(http.StatusBadRequest, "invalid numeric value", err)
	}

	return parsed, nil
}
