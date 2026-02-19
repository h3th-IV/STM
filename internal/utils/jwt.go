package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims holds custom claims for access tokens.
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// RefreshClaims holds claims for refresh tokens.
type RefreshClaims struct {
	TokenID string `json:"tid"`
	UserID  uint   `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTService handles JWT creation and validation.
type JWTService struct {
	Secret        string
	Issuer        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// NewJWTService creates a new JWT service.
func NewJWTService(secret, issuer string, accessMins, refreshDays int) *JWTService {
	return &JWTService{
		Secret:        secret,
		Issuer:        issuer,
		AccessExpiry:  time.Duration(accessMins) * time.Minute,
		RefreshExpiry: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

// GenerateAccessToken creates a new access token for the user.
func (s *JWTService) GenerateAccessToken(userID uint, role, email string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Secret))
}

// GenerateRefreshToken creates a new refresh token.
func (s *JWTService) GenerateRefreshToken(userID uint, tokenID string) (string, error) {
	claims := &RefreshClaims{
		TokenID: tokenID,
		UserID:  userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.Issuer,
			ID:        tokenID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Secret))
}

// ValidateAccessToken parses and validates an access token, returns claims.
func (s *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken parses and validates a refresh token.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
