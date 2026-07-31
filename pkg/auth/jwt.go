package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"eform/backend/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func NewManager(secret string, accessTokenTTL, refreshTokenTTL time.Duration) *Manager {
	return &Manager{
		secret:          []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (m *Manager) GenerateTokenPair(authClaims domain.AuthClaims) (domain.TokenPair, error) {
	accessToken, err := m.generateToken(authClaims, "access", m.accessTokenTTL)
	if err != nil {
		return domain.TokenPair{}, err
	}

	refreshToken, err := m.generateToken(authClaims, "refresh", m.refreshTokenTTL)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.accessTokenTTL.Seconds()),
	}, nil
}

func (m *Manager) generateToken(authClaims domain.AuthClaims, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: authClaims.UserID.String(),
		Role:   authClaims.Role,
		Email:  authClaims.Email,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   authClaims.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
