package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusPending  = "pending"

	DocumentTypeCV   = "cv"
	DocumentTypeKTP  = "ktp"
	DocumentTypeNPWP = "npwp"
)

type Role struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Code      string         `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	ID           uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	RoleID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"roleId"`
	Role         Role            `json:"role"`
	Email        string          `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string          `gorm:"size:255;not null" json:"-"`
	FullName     string          `gorm:"size:255;not null" json:"fullName"`
	Phone        string          `gorm:"size:50;uniqueIndex;not null" json:"phone"`
	KTPNumber    string          `gorm:"size:50;uniqueIndex;not null" json:"ktpNumber"`
	Status       string          `gorm:"size:30;default:'active';index;not null" json:"status"`
	LastLoginAt  *time.Time      `json:"lastLoginAt"`
	Profile      EmployeeProfile `json:"profile"`
	Documents    []Document      `json:"documents"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

type EmployeeProfile struct {
	ID                       uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID                   uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`
	JoinDate                 *time.Time     `json:"joinDate"`
	EntryDate                *time.Time     `json:"entryDate"`
	BirthPlace               string         `gorm:"size:150" json:"birthPlace"`
	BirthDate                *time.Time     `json:"birthDate"`
	Gender                   string         `gorm:"size:30" json:"gender"`
	Whatsapp                 string         `gorm:"size:50" json:"whatsapp"`
	BankAccount              string         `gorm:"size:100" json:"bankAccount"`
	NPWPNumber               string         `gorm:"size:50" json:"npwpNumber"`
	Position                 string         `gorm:"size:100" json:"position"`
	KTPAddress               string         `gorm:"type:text" json:"ktpAddress"`
	CurrentAddress           string         `gorm:"type:text" json:"currentAddress"`
	Religion                 string         `gorm:"size:50" json:"religion"`
	MaritalStatus            string         `gorm:"size:50" json:"maritalStatus"`
	Dependents               int            `gorm:"default:0" json:"dependents"`
	Education                string         `gorm:"size:100" json:"education"`
	BloodType                string         `gorm:"size:10" json:"bloodType"`
	InstituteName            string         `gorm:"size:150" json:"instituteName"`
	EmergencyContactName     string         `gorm:"size:150" json:"emergencyContactName"`
	EmergencyContactPhone    string         `gorm:"size:50" json:"emergencyContactPhone"`
	EmergencyContactAddress  string         `gorm:"type:text" json:"emergencyContactAddress"`
	EmergencyContactRelation string         `gorm:"size:100" json:"emergencyContactRelation"`
	CreatedAt                time.Time      `json:"createdAt"`
	UpdatedAt                time.Time      `json:"updatedAt"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}

type Document struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	Type      string         `gorm:"size:20;not null;uniqueIndex:idx_user_document_type" json:"type"`
	FileName  string         `gorm:"size:255;not null" json:"fileName"`
	FilePath  string         `gorm:"size:255;not null" json:"filePath"`
	MimeType  string         `gorm:"size:100;not null" json:"mimeType"`
	SizeBytes int64          `gorm:"not null" json:"sizeBytes"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ActorUserID  *uuid.UUID `gorm:"type:uuid;index" json:"actorUserId"`
	TargetUserID *uuid.UUID `gorm:"type:uuid;index" json:"targetUserId"`
	Action       string     `gorm:"size:100;index;not null" json:"action"`
	Resource     string     `gorm:"size:100;index;not null" json:"resource"`
	Description  string     `gorm:"type:text;not null" json:"description"`
	IPAddress    string     `gorm:"size:64" json:"ipAddress"`
	UserAgent    string     `gorm:"size:255" json:"userAgent"`
	Payload      string     `gorm:"type:text" json:"payload"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type PasswordReset struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email     string         `gorm:"size:255;index;not null" json:"email"`
	TokenHash string         `gorm:"size:255;uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expiresAt"`
	UsedAt    *time.Time     `json:"usedAt"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type RefreshToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"userId"`
	TokenHash string         `gorm:"size:255;uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expiresAt"`
	RevokedAt *time.Time     `json:"revokedAt"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuthClaims struct {
	UserID uuid.UUID `json:"userId"`
	Role   string    `json:"role"`
	Email  string    `json:"email"`
}

type RegisterRequest struct {
	Email                    string     `json:"email" validate:"required,email"`
	Password                 string     `json:"password" validate:"required,min=8"`
	FullName                 string     `json:"fullName" validate:"required,min=3,max=255"`
	Phone                    string     `json:"phone" validate:"required,min=8,max=50"`
	KTPNumber                string     `json:"ktpNumber" validate:"required,min=8,max=50"`
	JoinDate                 *time.Time `json:"joinDate"`
	EntryDate                *time.Time `json:"entryDate"`
	BirthPlace               string     `json:"birthPlace" validate:"max=150"`
	BirthDate                *time.Time `json:"birthDate"`
	Gender                   string     `json:"gender" validate:"omitempty,oneof=male female other"`
	Whatsapp                 string     `json:"whatsapp" validate:"max=50"`
	BankAccount              string     `json:"bankAccount" validate:"max=100"`
	NPWPNumber               string     `json:"npwpNumber" validate:"max=50"`
	Position                 string     `json:"position" validate:"max=100"`
	KTPAddress               string     `json:"ktpAddress"`
	CurrentAddress           string     `json:"currentAddress"`
	Religion                 string     `json:"religion" validate:"max=50"`
	MaritalStatus            string     `json:"maritalStatus" validate:"max=50"`
	Dependents               int        `json:"dependents" validate:"min=0,max=20"`
	Education                string     `json:"education" validate:"max=100"`
	BloodType                string     `json:"bloodType" validate:"max=10"`
	InstituteName            string     `json:"instituteName" validate:"max=150"`
	EmergencyContactName     string     `json:"emergencyContactName" validate:"max=150"`
	EmergencyContactPhone    string     `json:"emergencyContactPhone" validate:"max=50"`
	EmergencyContactAddress  string     `json:"emergencyContactAddress"`
	EmergencyContactRelation string     `json:"emergencyContactRelation" validate:"max=100"`
}

type EmployeeUpsertRequest struct {
	Email                    string     `json:"email" validate:"required,email"`
	Password                 string     `json:"password" validate:"omitempty,min=8"`
	FullName                 string     `json:"fullName" validate:"required,min=3,max=255"`
	Phone                    string     `json:"phone" validate:"required,min=8,max=50"`
	KTPNumber                string     `json:"ktpNumber" validate:"required,min=8,max=50"`
	JoinDate                 *time.Time `json:"joinDate"`
	EntryDate                *time.Time `json:"entryDate"`
	BirthPlace               string     `json:"birthPlace" validate:"max=150"`
	BirthDate                *time.Time `json:"birthDate"`
	Gender                   string     `json:"gender" validate:"omitempty,oneof=male female other"`
	Whatsapp                 string     `json:"whatsapp" validate:"max=50"`
	BankAccount              string     `json:"bankAccount" validate:"max=100"`
	NPWPNumber               string     `json:"npwpNumber" validate:"max=50"`
	Position                 string     `json:"position" validate:"max=100"`
	KTPAddress               string     `json:"ktpAddress"`
	CurrentAddress           string     `json:"currentAddress"`
	Religion                 string     `json:"religion" validate:"max=50"`
	MaritalStatus            string     `json:"maritalStatus" validate:"max=50"`
	Dependents               int        `json:"dependents" validate:"min=0,max=20"`
	Education                string     `json:"education" validate:"max=100"`
	BloodType                string     `json:"bloodType" validate:"max=10"`
	InstituteName            string     `json:"instituteName" validate:"max=150"`
	EmergencyContactName     string     `json:"emergencyContactName" validate:"max=150"`
	EmergencyContactPhone    string     `json:"emergencyContactPhone" validate:"max=50"`
	EmergencyContactAddress  string     `json:"emergencyContactAddress"`
	EmergencyContactRelation string     `json:"emergencyContactRelation" validate:"max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,min=8"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

type ResetEmployeePasswordRequest struct {
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

type EmployeeListFilter struct {
	Search    string
	Status    string
	Role      string
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}

type RequestMeta struct {
	IPAddress string
	UserAgent string
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type DocumentResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	MimeType  string    `json:"mimeType"`
	SizeBytes int64     `json:"sizeBytes"`
}

type EmployeeResponse struct {
	ID                       uuid.UUID          `json:"id"`
	Role                     string             `json:"role"`
	Status                   string             `json:"status"`
	Email                    string             `json:"email"`
	FullName                 string             `json:"fullName"`
	Phone                    string             `json:"phone"`
	KTPNumber                string             `json:"ktpNumber"`
	JoinDate                 *time.Time         `json:"joinDate"`
	EntryDate                *time.Time         `json:"entryDate"`
	BirthPlace               string             `json:"birthPlace"`
	BirthDate                *time.Time         `json:"birthDate"`
	Gender                   string             `json:"gender"`
	Whatsapp                 string             `json:"whatsapp"`
	BankAccount              string             `json:"bankAccount"`
	NPWPNumber               string             `json:"npwpNumber"`
	Position                 string             `json:"position"`
	KTPAddress               string             `json:"ktpAddress"`
	CurrentAddress           string             `json:"currentAddress"`
	Religion                 string             `json:"religion"`
	MaritalStatus            string             `json:"maritalStatus"`
	Dependents               int                `json:"dependents"`
	Education                string             `json:"education"`
	BloodType                string             `json:"bloodType"`
	InstituteName            string             `json:"instituteName"`
	EmergencyContactName     string             `json:"emergencyContactName"`
	EmergencyContactPhone    string             `json:"emergencyContactPhone"`
	EmergencyContactAddress  string             `json:"emergencyContactAddress"`
	EmergencyContactRelation string             `json:"emergencyContactRelation"`
	LastLoginAt              *time.Time         `json:"lastLoginAt"`
	ProfileCompletion        int                `json:"profileCompletion"`
	Documents                []DocumentResponse `json:"documents"`
	CreatedAt                time.Time          `json:"createdAt"`
	UpdatedAt                time.Time          `json:"updatedAt"`
}

type AuthResponse struct {
	Token TokenPair        `json:"token"`
	User  EmployeeResponse `json:"user"`
}

type AdminDashboardResponse struct {
	TotalEmployees    int64              `json:"totalEmployees"`
	NewEmployees      int64              `json:"newEmployees"`
	ActiveEmployees   int64              `json:"activeEmployees"`
	InactiveEmployees int64              `json:"inactiveEmployees"`
	LatestEmployees   []EmployeeResponse `json:"latestEmployees"`
}

type UserDashboardResponse struct {
	Profile           EmployeeResponse   `json:"profile"`
	ProfileCompletion int                `json:"profileCompletion"`
	UploadedDocuments []DocumentResponse `json:"uploadedDocuments"`
}
