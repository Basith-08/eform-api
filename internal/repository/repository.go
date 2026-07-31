package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"eform/backend/internal/domain"
	"eform/backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repositories struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{db: db}
}

func (r *Repositories) WithDB(db *gorm.DB) *Repositories {
	return &Repositories{db: db}
}

func (r *Repositories) DB() *gorm.DB {
	return r.db
}

func (r *Repositories) FindRoleByCode(ctx context.Context, code string) (domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	return role, err
}

func (r *Repositories) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Profile").
		Preload("Documents").
		Where("LOWER(email) = ?", strings.ToLower(email)).
		First(&user).Error
	return user, err
}

func (r *Repositories) FindUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Profile").
		Preload("Documents").
		First(&user, "id = ?", userID).Error
	return user, err
}

func (r *Repositories) ExistsByEmail(ctx context.Context, email string, exceptID *uuid.UUID) (bool, error) {
	return r.exists(ctx, "email", strings.ToLower(email), exceptID)
}

func (r *Repositories) ExistsByPhone(ctx context.Context, phone string, exceptID *uuid.UUID) (bool, error) {
	return r.exists(ctx, "phone", phone, exceptID)
}

func (r *Repositories) ExistsByKTP(ctx context.Context, ktp string, exceptID *uuid.UUID) (bool, error) {
	return r.exists(ctx, "ktp_number", ktp, exceptID)
}

func (r *Repositories) exists(ctx context.Context, column string, value string, exceptID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domain.User{}).Where(column+" = ?", value)
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repositories) CreateUser(ctx context.Context, user *domain.User, profile *domain.EmployeeProfile, documents []domain.Document) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		profile.UserID = user.ID
		if err := tx.Create(profile).Error; err != nil {
			return err
		}

		for index := range documents {
			documents[index].UserID = user.ID
		}

		if len(documents) > 0 {
			if err := tx.Create(&documents).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repositories) UpdateUserAggregate(ctx context.Context, user *domain.User, profile *domain.EmployeeProfile, documents []domain.Document) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", user.ID).Delete(&domain.Document{}).Error; err != nil {
			return err
		}

		profile.UserID = user.ID
		if err := tx.Where(&domain.EmployeeProfile{UserID: user.ID}).Assign(profile).FirstOrCreate(profile).Error; err != nil {
			return err
		}

		for index := range documents {
			documents[index].UserID = user.ID
		}

		if len(documents) > 0 {
			if err := tx.Create(&documents).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repositories) SetLastLogin(ctx context.Context, userID uuid.UUID, loginAt time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("last_login_at", loginAt).Error
}

func (r *Repositories) ListEmployees(ctx context.Context, filter domain.EmployeeListFilter) ([]domain.User, int64, error) {
	allowedSorts := map[string]string{
		"fullName":  "users.full_name",
		"email":     "users.email",
		"phone":     "users.phone",
		"position":  "employee_profiles.position",
		"status":    "users.status",
		"joinDate":  "employee_profiles.join_date",
		"createdAt": "users.created_at",
	}

	normalizedSort := allowedSorts[filter.SortBy]
	if normalizedSort == "" {
		normalizedSort = "users.created_at"
	}

	sortOrder := "DESC"
	if strings.EqualFold(filter.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	query := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Joins("LEFT JOIN employee_profiles ON employee_profiles.user_id = users.id").
		Preload("Role").
		Preload("Profile").
		Preload("Documents")

	if filter.Role != "" {
		query = query.Joins("JOIN roles ON roles.id = users.role_id").Where("roles.code = ?", filter.Role)
	}

	if filter.Status != "" {
		query = query.Where("users.status = ?", filter.Status)
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(users.full_name) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(users.phone) LIKE ? OR LOWER(COALESCE(employee_profiles.position, '')) LIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []domain.User
	if err := query.Order(normalizedSort + " " + sortOrder).
		Limit(utils.NormalizeLimit(filter.Limit)).
		Offset((utils.NormalizePage(filter.Page) - 1) * utils.NormalizeLimit(filter.Limit)).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *Repositories) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.Document{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", userID).Delete(&domain.EmployeeProfile{}).Error; err != nil {
			return err
		}

		return tx.Delete(&domain.User{}, "id = ?", userID).Error
	})
}

func (r *Repositories) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("status", status).Error
}

func (r *Repositories) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

func (r *Repositories) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *Repositories) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repositories) FindRefreshToken(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	return token, err
}

func (r *Repositories) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", time.Now()).
		Error
}

func (r *Repositories) CreatePasswordReset(ctx context.Context, reset *domain.PasswordReset) error {
	return r.db.WithContext(ctx).Create(reset).Error
}

func (r *Repositories) FindValidPasswordReset(ctx context.Context, tokenHash string) (domain.PasswordReset, error) {
	var reset domain.PasswordReset
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&reset).Error
	return reset, err
}

func (r *Repositories) MarkPasswordResetUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.PasswordReset{}).Where("id = ?", id).Update("used_at", now).Error
}

func (r *Repositories) CountEmployees(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.code = ?", domain.RoleUser).
		Count(&count).
		Error
	return count, err
}

func (r *Repositories) CountEmployeesByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.code = ? AND users.status = ?", domain.RoleUser, status).
		Count(&count).
		Error
	return count, err
}

func (r *Repositories) CountNewEmployees(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.code = ? AND users.created_at >= ?", domain.RoleUser, since).
		Count(&count).
		Error
	return count, err
}

func (r *Repositories) LatestEmployees(ctx context.Context, limit int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.code = ?", domain.RoleUser).
		Preload("Role").
		Preload("Profile").
		Preload("Documents").
		Order("users.created_at DESC").
		Limit(limit).
		Find(&users).
		Error
	return users, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
